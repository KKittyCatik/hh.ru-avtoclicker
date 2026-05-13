package scheduler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/robfig/cron/v3"
	"hh-autoresponder/internal/hh"
	"hh-autoresponder/internal/limiter"
)

type Scheduler struct {
	cron         *cron.Cron
	hhClient     *hh.Client
	dailyLimiter *limiter.DailyLimiter
	logger       *slog.Logger
	resumeIDs    []string
}

func New(hhClient *hh.Client, dailyLimiter *limiter.DailyLimiter, resumeIDs []string, logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{cron: cron.New(), hhClient: hhClient, dailyLimiter: dailyLimiter, logger: logger, resumeIDs: resumeIDs}
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
	resumes, err := s.hhClient.ListResumes(ctx)
	if err != nil {
		return fmt.Errorf("list resumes for publish: %w", err)
	}
	allowed := make(map[string]struct{}, len(s.resumeIDs))
	for _, id := range s.resumeIDs {
		allowed[id] = struct{}{}
	}
	selected := make([]hh.Resume, 0, len(resumes))
	for _, r := range resumes {
		if len(allowed) > 0 {
			if _, ok := allowed[r.ID]; !ok {
				continue
			}
		}
		selected = append(selected, r)
	}
	published, err := s.hhClient.PublishReadyResumes(ctx, selected)
	if err != nil {
		return fmt.Errorf("publish ready resumes: %w", err)
	}
	for _, id := range published {
		s.logger.Info("resume published", "resume_id", id)
	}
	return nil
}
