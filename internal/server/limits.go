package server

import "github.com/stockyard-dev/stockyard-muster/internal/license"

type Limits struct {
	MaxMembers   int
	MaxIncidents int
	Webhooks     bool
}

var freeLimits = Limits{MaxMembers: 5, MaxIncidents: 20, Webhooks: false}
var proLimits = Limits{MaxMembers: 0, MaxIncidents: 0, Webhooks: true}

func LimitsFor(info *license.Info) Limits {
	if info != nil && info.IsPro() {
		return proLimits
	}
	return freeLimits
}

func LimitReached(limit, current int) bool {
	if limit == 0 {
		return false
	}
	return current >= limit
}
