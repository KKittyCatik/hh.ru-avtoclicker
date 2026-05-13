package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"golang.org/x/sync/errgroup"

	"hh-autoresponder/internal/filters"
	"hh-autoresponder/internal/hh"
	"hh-autoresponder/internal/limiter"
	"hh-autoresponder/internal/llm"
	"hh-autoresponder/internal/monitor"
	"hh-autoresponder/internal/transport/ws"
)

type ApplyWorker struct {
	hhClient    *hh.Client
	llmClient   llm.LLMClient
	limiter     *limiter.DailyLimiter
	filterChain *filters.FilterChain
	stats       *monitor.Collector
	hub         *ws.Hub
	logger      *slog.Logger

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
}

func NewApplyWorker(hhClient *hh.Client, llmClient llm.LLMClient, limiter *limiter.DailyLimiter, chain *filters.FilterChain, stats *monitor.Collector, hub *ws.Hub, logger *slog.Logger) *ApplyWorker {
	if logger == nil {
		logger = slog.Default()
	}
	return &ApplyWorker{hhClient: hhClient, llmClient: llmClient, limiter: limiter, filterChain: chain, stats: stats, hub: hub, logger: logger}
}

func (w *ApplyWorker) Start(ctx context.Context, searchURLs []string, resumeID string) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("apply worker already running")
	}
	runCtx, cancel := context.WithCancel(ctx)
	w.running = true
	w.cancel = cancel
	w.mu.Unlock()

	go func() {
		defer func() {
			w.mu.Lock()
			w.running = false
			w.cancel = nil
			w.mu.Unlock()
		}()

		if err := w.run(runCtx, searchURLs, resumeID); err != nil {
			w.logger.Error("apply worker failed", "error", err)
		}
	}()
	return nil
}

func (w *ApplyWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		w.cancel()
	}
}

func (w *ApplyWorker) run(ctx context.Context, searchURLs []string, resumeID string) error {
	vacancies, err := w.hhClient.CollectVacanciesByURLs(ctx, searchURLs, 5)
	if err != nil {
		return fmt.Errorf("collect vacancies: %w", err)
	}

	g, gctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, 5)
	for _, vacancy := range vacancies {
		vacancy := vacancy
		g.Go(func() error {
			select {
			case sem <- struct{}{}:
			case <-gctx.Done():
				return fmt.Errorf("wait apply semaphore: %w", gctx.Err())
			}
			defer func() { <-sem }()

			ok, reason := w.filterChain.Filter(vacancy)
			if !ok {
				w.logger.Info("vacancy filtered", "vacancy_id", vacancy.ID, "company", vacancy.Employer.Name, "reason", reason)
				return nil
			}
			if !w.limiter.Allow() {
				_ = w.hub.Broadcast(gctx, "limit_reached", map[string]any{"vacancy_id": vacancy.ID})
				return nil
			}

			coverLetter, err := w.llmClient.Complete(gctx, "Ты соискатель. Напиши краткое сопроводительное письмо на русском.", fmt.Sprintf("Вакансия: %s\nОписание: %s", vacancy.Name, vacancy.Description))
			if err != nil {
				return fmt.Errorf("generate cover letter for %s: %w", vacancy.ID, err)
			}
			applyErr := w.hhClient.ApplyToVacancy(gctx, hh.ApplyRequest{ResumeID: resumeID, VacancyID: vacancy.ID, CoverLetter: coverLetter})
			result := hh.ApplyResult{VacancyID: vacancy.ID, Company: vacancy.Employer.Name, Status: "applied"}
			if applyErr != nil {
				result.Status = "failed"
				result.Reason = applyErr.Error()
				w.logger.Error("apply failed", "vacancy_id", vacancy.ID, "company", vacancy.Employer.Name, "error", applyErr)
			} else {
				w.stats.IncApplies()
				w.logger.Info("apply success", "vacancy_id", vacancy.ID, "company", vacancy.Employer.Name)
			}
			_ = w.hub.Broadcast(gctx, "apply_result", result)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("run apply worker: %w", err)
	}
	return nil
}
