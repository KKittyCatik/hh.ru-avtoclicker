package filters

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"hh-autoresponder/internal/hh"
)

type VacancyFilter interface {
	Filter(v hh.Vacancy) (bool, string)
}

type FilterChain struct {
	filters []VacancyFilter
}

func NewFilterChain(filters ...VacancyFilter) *FilterChain {
	return &FilterChain{filters: filters}
}

func (f *FilterChain) Filter(v hh.Vacancy) (bool, string) {
	for _, filter := range f.filters {
		ok, reason := filter.Filter(v)
		if !ok {
			return false, reason
		}
	}
	return true, ""
}

type SalaryFilter struct {
	MinSalary int
}

func (f SalaryFilter) Filter(v hh.Vacancy) (bool, string) {
	if f.MinSalary <= 0 {
		return true, ""
	}
	if v.Salary.From < f.MinSalary {
		return false, fmt.Sprintf("salary %d less than %d", v.Salary.From, f.MinSalary)
	}
	return true, ""
}

type ScheduleFilter struct {
	Schedule string
}

func (f ScheduleFilter) Filter(v hh.Vacancy) (bool, string) {
	if f.Schedule == "" || f.Schedule == "any" {
		return true, ""
	}
	if strings.EqualFold(f.Schedule, v.Schedule) {
		return true, ""
	}
	return false, "schedule mismatch"
}

type AgencyFilter struct {
	Enabled bool
}

var agencyKeywords = []string{"кадровое", "рекрутинг", "staffing", "подбор персонала", "hr-агентство"}

func (f AgencyFilter) Filter(v hh.Vacancy) (bool, string) {
	if !f.Enabled {
		return true, ""
	}
	content := strings.ToLower(v.Employer.Name + " " + v.Description)
	for _, kw := range agencyKeywords {
		if strings.Contains(content, kw) {
			return false, "agency-like vacancy"
		}
	}
	return true, ""
}

type BlacklistFilter struct {
	CompanyIDs   map[string]struct{}
	CompanyNames map[string]struct{}
}

func NewBlacklistFilter(ids []string, names []string) BlacklistFilter {
	f := BlacklistFilter{CompanyIDs: map[string]struct{}{}, CompanyNames: map[string]struct{}{}}
	for _, id := range ids {
		f.CompanyIDs[strings.TrimSpace(id)] = struct{}{}
	}
	for _, name := range names {
		f.CompanyNames[strings.TrimSpace(name)] = struct{}{}
	}
	return f
}

func (f BlacklistFilter) Filter(v hh.Vacancy) (bool, string) {
	if _, ok := f.CompanyIDs[v.Employer.ID]; ok {
		return false, "company is blacklisted by id"
	}
	if _, ok := f.CompanyNames[v.Employer.Name]; ok {
		return false, "company is blacklisted by name"
	}
	return true, ""
}

type DuplicateFilter struct {
	mu      sync.Mutex
	path    string
	applied map[string]struct{}
}

func NewDuplicateFilter(path string) (*DuplicateFilter, error) {
	d := &DuplicateFilter{path: path, applied: map[string]struct{}{}}
	if err := d.load(); err != nil {
		return nil, fmt.Errorf("load duplicate history: %w", err)
	}
	return d, nil
}

func (f *DuplicateFilter) Filter(v hh.Vacancy) (bool, string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.applied[v.ID]; ok {
		return false, "duplicate vacancy"
	}
	return true, ""
}

func (f *DuplicateFilter) MarkApplied(vacancyID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.applied[vacancyID] = struct{}{}
	if err := f.persist(); err != nil {
		return fmt.Errorf("persist duplicate history: %w", err)
	}
	return nil
}

func (f *DuplicateFilter) load() error {
	if f.path == "" {
		return nil
	}
	b, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read duplicate file: %w", err)
	}
	var ids []string
	if err := json.Unmarshal(b, &ids); err != nil {
		return fmt.Errorf("unmarshal duplicate file: %w", err)
	}
	for _, id := range ids {
		f.applied[id] = struct{}{}
	}
	return nil
}

func (f *DuplicateFilter) persist() error {
	if f.path == "" {
		return nil
	}
	ids := make([]string, 0, len(f.applied))
	for id := range f.applied {
		ids = append(ids, id)
	}
	data, err := json.MarshalIndent(ids, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal duplicate file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return fmt.Errorf("create duplicate file dir: %w", err)
	}
	if err := os.WriteFile(f.path, data, 0o644); err != nil {
		return fmt.Errorf("write duplicate file: %w", err)
	}
	return nil
}
