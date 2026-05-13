package monitor

import "sync"

type Stats struct {
	AppliesToday int `json:"applies_today"`
	Invitations  int `json:"invitations"`
	Views        int `json:"views"`
}

type Collector struct {
	mu    sync.RWMutex
	stats Stats
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) Snapshot() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

func (c *Collector) IncApplies() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.AppliesToday++
}

func (c *Collector) SetInvitations(v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.Invitations = v
}

func (c *Collector) SetViews(v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.Views = v
}

func (c *Collector) ResetApplies() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats.AppliesToday = 0
}
