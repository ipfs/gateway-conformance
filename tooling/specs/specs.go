package specs

import "fmt"

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

type Spec interface {
	Name() string
	IsEnabled() bool
	IsMature() bool
	Enable()
	Disable()
}

type Leaf struct {
	name     string
	maturity maturity
}

func (l Leaf) Name() string {
	return l.name
}

func (l Leaf) IsMature() bool {
	return l.maturity.isMature()
}

func (l Leaf) IsEnabled() bool {
	// If the spec was explicitly enabled or disabled, use that.
	// Otherwise, use the maturity level.
	if enabled, ok := specEnabled[l]; ok {
		return enabled
	} else {
		return l.IsMature()
	}
}

func (s Leaf) Enable() {
	specEnabled[s] = true
}

func (s Leaf) Disable() {
	specEnabled[s] = false
}

type Collection struct {
	name     string
	children []Spec
}

func (c Collection) Name() string {
	return c.name
}

func (c Collection) IsEnabled() bool {
	for _, s := range c.children {
		if !s.IsEnabled() {
			return false
		}
	}

	return true
}

func (c Collection) IsMature() bool {
	for _, s := range c.children {
		if !s.IsMature() {
			return false
		}
	}

	return true
}

func (c Collection) Enable() {
	for _, s := range c.children {
		s.Enable()
	}
}

func (c Collection) Disable() {
	for _, s := range c.children {
		s.Disable()
	}
}

var (
	TrustlessBlockGateway = Leaf{"trustless-block-gateway", stable}
	TrustlessCARGateway   = Leaf{"trustless-car-gateway", stable}
	TrustlessIPNSGateway  = Leaf{"trustless-ipns-gateway", stable}
	TrustlessGateway      = Collection{"trustless-gateway", []Spec{TrustlessBlockGateway, TrustlessCARGateway, TrustlessIPNSGateway}}
	PathGateway           = Leaf{"path-gateway", stable}
	DNSLinkGateway        = Leaf{"dnslink-gateway", stable}
	SubdomainGateway      = Leaf{"subdomain-gateway", stable}
	RedirectsFile         = Leaf{"redirects-file", stable} // TODO: Can be run for path gateway, dnslink and subdomain gateways, but not by itself.
	IPNSResolver          = Leaf{"ipns-resolver", stable}
)

// All specs MUST be listed here.
var specs = []Spec{
	TrustlessBlockGateway,
	TrustlessCARGateway,
	TrustlessIPNSGateway,
	TrustlessGateway,
	PathGateway,
	DNSLinkGateway,
	SubdomainGateway,
	RedirectsFile,
	IPNSResolver,
}

var specEnabled = map[Spec]bool{}

func All() []Spec {
	return specs
}

func FromString(name string) (Spec, error) {
	for _, spec := range All() {
		if spec.Name() == name {
			return spec, nil
		}
	}
	return nil, fmt.Errorf("unknown spec: %s", name)
}
