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
	"hh-autoresponder/internal/config"
	"hh-autoresponder/internal/filters"
	"hh-autoresponder/internal/hh"
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
		logger.Info("no active accounts on startup, waiting for oauth login")
	}

	auth := hh.NewAuthManager(
		cfg.HHClientID,
		cfg.HHClientSecret,
		cfg.HHRedirectURI,
		[]string{"resumes", "negotiations", "vacancy_response"},
		logger,
		&acct.NeedsReauth,
	)
	auth.SetToken(acct.Token)
	hhClient := hh.NewClient(&http.Client{Timeout: 60 * time.Second}, auth, logger)

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
	applyWorker := worker.NewApplyWorker(hhClient, llmClient, dailyLimiter, filterChain, stats, hub, logger)
	replyWorker := worker.NewReplyWorker(hhClient, llmClient, hub, logger)

	resumeID := ""
	if len(acct.ResumeIDs) > 0 {
		resumeID = acct.ResumeIDs[0]
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	handlers := &httptransport.Handlers{
		Ctx:         ctx,
		Auth:        auth,
		ApplyWorker: applyWorker,
		ReplyWorker: replyWorker,
		Stats:       stats,
		Accounts:    acctMgr,
		HHClient:    hhClient,
		SearchURLs:  acct.SearchURLs,
		ResumeID:    resumeID,
	}
	server := httptransport.NewServer(cfg.Port, handlers, hub)

	sched := scheduler.New(hhClient, dailyLimiter, acct.ResumeIDs, logger)
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
