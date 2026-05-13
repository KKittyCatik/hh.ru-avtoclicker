package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HHClientID      string
	HHClientSecret  string
	LLMProvider     string
	OpenAIAPIKey    string
	DeepSeekAPIKey  string
	DailyApplyLimit int
	MinSalary       int
	ScheduleFilter  string
	ExcludeAgencies bool
	AccountsFile    string
	Port            string
}

func Load() (Config, error) {
	cfg := Config{
		HHClientID:      strings.TrimSpace(os.Getenv("HH_CLIENT_ID")),
		HHClientSecret:  strings.TrimSpace(os.Getenv("HH_CLIENT_SECRET")),
		LLMProvider:     strings.ToLower(strings.TrimSpace(os.Getenv("LLM_PROVIDER"))),
		OpenAIAPIKey:    strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		DeepSeekAPIKey:  strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		DailyApplyLimit: 50,
		ScheduleFilter:  "any",
		AccountsFile:    strings.TrimSpace(os.Getenv("ACCOUNTS_FILE")),
		Port:            "8080",
	}

	if cfg.LLMProvider == "" {
		cfg.LLMProvider = "openai"
	}
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" {
		cfg.Port = p
	}
	if s := strings.TrimSpace(os.Getenv("DAILY_APPLY_LIMIT")); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return Config{}, fmt.Errorf("parse DAILY_APPLY_LIMIT: %w", err)
		}
		cfg.DailyApplyLimit = v
	}
	if s := strings.TrimSpace(os.Getenv("MIN_SALARY")); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return Config{}, fmt.Errorf("parse MIN_SALARY: %w", err)
		}
		cfg.MinSalary = v
	}
	if s := strings.ToLower(strings.TrimSpace(os.Getenv("SCHEDULE_FILTER"))); s != "" {
		switch s {
		case "remote", "hybrid", "office", "any":
			cfg.ScheduleFilter = s
		default:
			return Config{}, fmt.Errorf("invalid SCHEDULE_FILTER: %s", s)
		}
	}
	if s := strings.TrimSpace(os.Getenv("EXCLUDE_AGENCIES")); s != "" {
		v, err := strconv.ParseBool(s)
		if err != nil {
			return Config{}, fmt.Errorf("parse EXCLUDE_AGENCIES: %w", err)
		}
		cfg.ExcludeAgencies = v
	}

	if cfg.AccountsFile == "" {
		return Config{}, fmt.Errorf("ACCOUNTS_FILE is required")
	}
	if cfg.LLMProvider != "openai" && cfg.LLMProvider != "deepseek" {
		return Config{}, fmt.Errorf("invalid LLM_PROVIDER: %s", cfg.LLMProvider)
	}

	return cfg, nil
}
