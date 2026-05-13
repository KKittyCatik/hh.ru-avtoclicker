package filters

import (
	"testing"

	"hh-autoresponder/internal/hh"
)

func TestSalaryFilter(t *testing.T) {
	f := SalaryFilter{MinSalary: 200000}
	ok, _ := f.Filter(hh.Vacancy{Salary: hh.Salary{From: 150000}})
	if ok {
		t.Fatal("expected salary filter reject")
	}
}

func TestFilterChainStopsOnFirstReject(t *testing.T) {
	chain := NewFilterChain(SalaryFilter{MinSalary: 200000}, ScheduleFilter{Schedule: "remote"})
	ok, reason := chain.Filter(hh.Vacancy{Salary: hh.Salary{From: 100000}, Schedule: "remote"})
	if ok || reason == "" {
		t.Fatal("expected reject reason")
	}
}
