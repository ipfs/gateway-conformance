package specs

import (
	"fmt"
)

type maturity string

const (
	wip        maturity = "wip"
	draft      maturity = "draft"
	reliable   maturity = "reliable"
	stable     maturity = "stable"
	permanent  maturity = "permanent"
	deprecated maturity = "deprecated"
)

func (m maturity) isMature() bool {
	switch m {
	case reliable, stable, permanent:
		return true
	default:
		return false
	}
}

type Spec string

const (
	SubdomainGateway Spec = "subdomain-gateway"
)

// All specs should be listed here.
var specMaturity = map[Spec]maturity{
	SubdomainGateway: stable,
}

func (s Spec) IsMature() bool {
	return specMaturity[s].isMature()
}

var specEnabled = map[Spec]bool{}

func (s Spec) IsEnabled() bool {
	// If the spec was explicitly enabled or disabled, use that.
	// Otherwise, use the maturity level.
	if enabled, ok := specEnabled[s]; ok {
		return enabled
	} else {
		return s.IsMature()
	}
}

func (s Spec) Enable() {
	specEnabled[s] = true
}

func (s Spec) Disable() {
	specEnabled[s] = false
}

func All() []Spec {
	specs := []Spec{}
	for spec := range specMaturity {
		specs = append(specs, spec)
	}
	return specs
}

func FromString(name string) (Spec, error) {
	for _, spec := range All() {
		if string(spec) == name {
			return spec, nil
		}
	}
	return "", fmt.Errorf("unknown spec: %s", name)
}
