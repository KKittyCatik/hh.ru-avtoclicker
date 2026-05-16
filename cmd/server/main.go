package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hh-autoresponder/internal/account"
	"hh-autoresponder/internal/browser"
	"hh-autoresponder/internal/config"
	"hh-autoresponder/internal/filters"
	"hh-autoresponder/internal/limiter"
	"hh-autoresponder/internal/llm"
	"hh-autoresponder/internal/monitor"
	"hh-autoresponder/internal/scheduler"
	httptransport "hh-autoresponder/internal/transport/http"
	"hh-autoresponder/internal/transport/ws"
	"hh-autoresponder/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	acctMgr := account.NewManager(cfg.AccountsFile)
	if err := acctMgr.Load(); err != nil {
		logger.Error("load accounts", "error", err)
		os.Exit(1)
	}
	active := acctMgr.GetActive()
	var acct account.Account
	if len(active) > 0 {
		acct = active[0]
	} else {
		logger.Info("no active accounts on startup")
	}

	llmClient, err := llm.NewProvider(cfg.LLMProvider, cfg.OpenAIAPIKey, cfg.DeepSeekAPIKey)
	if err != nil {
		logger.Error("init llm provider", "error", err)
		os.Exit(1)
	}

	dupFilter, err := filters.NewDuplicateFilter("data/applied_vacancies.json")
	if err != nil {
		logger.Error("init duplicate filter", "error", err)
		os.Exit(1)
	}
	filterChain := filters.NewFilterChain(
		filters.SalaryFilter{MinSalary: cfg.MinSalary},
		filters.ScheduleFilter{Schedule: cfg.ScheduleFilter},
		filters.AgencyFilter{Enabled: cfg.ExcludeAgencies},
		filters.NewBlacklistFilter(acct.Preferences.BlacklistIDs, acct.Preferences.BlacklistNames),
		dupFilter,
	)

	dailyLimiter := limiter.NewDailyLimiter(cfg.DailyApplyLimit)
	stats := monitor.NewCollector()
	hub := ws.NewHub()

	resumeID := ""
	if len(acct.ResumeIDs) > 0 {
		resumeID = acct.ResumeIDs[0]
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bm, err := browser.NewBrowserManager(ctx)
	if err != nil {
		logger.Error("init browser manager", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := bm.Close(); err != nil {
			logger.Error("close browser manager", "error", err)
		}
	}()

	var browserCtx *browser.AccountContext
	if acct.ID != "" {
		browserCtx, err = bm.AddAccount(acct.ID)
		if err != nil {
			logger.Error("add account browser context", "account", acct.ID, "error", err)
		}
	}

	if browserCtx != nil {
		loggedIn, err := browserCtx.IsLoggedIn()
		if err != nil {
			logger.Warn("check browser session", "account", acct.ID, "error", err)
		}
		if !loggedIn {
			key, keyErr := account.LoadEncryptionKeyFromEnv()
			if keyErr != nil {
				logger.Error("load encryption key", "error", keyErr)
			} else {
				password, decErr := account.DecryptPassword(acct.Password, key)
				if decErr != nil {
					logger.Error("decrypt account password", "account", acct.ID, "error", decErr)
				} else if err := browserCtx.Login(acct.Email, password); err != nil {
					logger.Error("login failed", "account", acct.ID, "error", err)
				}
			}
		}
	}

	bm.SetNotifier(func(ctx context.Context, event string, payload any) error {
		return hub.Broadcast(ctx, event, payload)
	})
	applyWorker := worker.NewApplyWorker(browserCtx, llmClient, dailyLimiter, filterChain, stats, hub, logger)
	replyWorker := worker.NewReplyWorker(browserCtx, llmClient, hub, logger)

	handlers := &httptransport.Handlers{
		Ctx:         ctx,
		BrowserCtx:  browserCtx,
		ApplyWorker: applyWorker,
		ReplyWorker: replyWorker,
		Stats:       stats,
		Accounts:    acctMgr,
		SearchURLs:  acct.SearchURLs,
		ResumeID:    resumeID,
	}
	server := httptransport.NewServer(cfg.Port, handlers, hub)

	sched := scheduler.New(browserCtx, dailyLimiter, acct.ResumeIDs, logger)
	if err := sched.Start(ctx); err != nil {
		logger.Error("start scheduler", "error", err)
		os.Exit(1)
	}

	go func() {
		if err := dailyLimiter.StartAutoReset(ctx); err != nil {
			logger.Info("daily limiter auto reset stopped", "error", err)
		}
	}()

	go func() {
		logger.Info("server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown server", "error", err)
	}
	logger.Info("server stopped")
}
