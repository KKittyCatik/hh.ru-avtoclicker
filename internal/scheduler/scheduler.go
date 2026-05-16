package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"hh-autoresponder/internal/browser"
	"hh-autoresponder/internal/limiter"
)

type Scheduler struct {
	cron         *cron.Cron
	browserCtx   *browser.AccountContext
	dailyLimiter *limiter.DailyLimiter
	logger       *slog.Logger
	resumeIDs    []string
}

func New(browserCtx *browser.AccountContext, dailyLimiter *limiter.DailyLimiter, resumeIDs []string, logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{cron: cron.New(), browserCtx: browserCtx, dailyLimiter: dailyLimiter, logger: logger, resumeIDs: resumeIDs}
}

func (s *Scheduler) Start(ctx context.Context) error {
	// Every 4 hours at minute 0.
	if _, err := s.cron.AddFunc("0 */4 * * *", func() {
		if err := s.publishResumes(ctx); err != nil {
			s.logger.Error("publish resumes job failed", "error", err)
		}
	}); err != nil {
		return fmt.Errorf("add publish cron job: %w", err)
	}
	if _, err := s.cron.AddFunc("0 0 * * *", func() {
		s.dailyLimiter.Reset()
		s.logger.Info("daily apply limit reset")
	}); err != nil {
		return fmt.Errorf("add limiter reset cron job: %w", err)
	}
	s.cron.Start()
	go func() {
		<-ctx.Done()
		s.cron.Stop()
	}()
	return nil
}

func (s *Scheduler) publishResumes(ctx context.Context) error {
	if s.browserCtx == nil {
		return fmt.Errorf("browser context is not configured")
	}
	for _, id := range s.resumeIDs {
		nextAt, err := s.browserCtx.GetNextPublishTime(ctx, id)
		if err != nil {
			return fmt.Errorf("get next publish time for %s: %w", id, err)
		}
		if time.Now().Before(nextAt) {
			s.logger.Info("resume publish skipped", "resume_id", id, "next_publish_at", nextAt)
			continue
		}
		if err := s.browserCtx.PublishResume(ctx, id); err != nil {
			if errors.Is(err, browser.ErrPublishNotAvailable) {
				s.logger.Info("resume publish not available yet", "resume_id", id)
				continue
			}
			return fmt.Errorf("publish resume %s: %w", id, err)
		}
		s.logger.Info("resume published", "resume_id", id)
	}
	return nil
}
