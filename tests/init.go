package tests

import (
	"flag"
	"regexp"
	"strings"

	"github.com/ipfs/gateway-conformance/tooling/specs"
)

type specsFlag string

func (s *specsFlag) String() string {
	return string(*s)
}

func (s *specsFlag) Set(value string) error {
	names := strings.Split(value, ",")
	var only, enable, disable = []specs.Spec{}, []specs.Spec{}, []specs.Spec{}
	for _, name := range names {
		spec, err := specs.FromString(regexp.MustCompile(`^[-+]`).ReplaceAllString(name, ""))
		if err != nil {
			return err
		}
		if strings.HasPrefix(name, "+") {
			// If a spec from the input is prefixed with a +,
			// it will be explicitly enabled.
			enable = append(enable, spec)
		} else if strings.HasPrefix(name, "-") {
			// If a spec from the input is prefixed with a -,
			// it will be explicitly disabled.
			disable = append(disable, spec)
		} else {
			// If a spec from the input is not prefixed with a + or -,
			// only the specified specs will be enabled.
			only = append(only, spec)
		}
	}
	if len(only) > 0 {
		// If any specs from the input are unprefixed,
		// disable all specs and then enable only the specified specs.
		for _, spec := range specs.All() {
			spec.Disable()
		}
		for _, spec := range only {
			spec.Enable()
		}
	}
	// If some specs from the input are prefixed with a + or -,
	// enable the specs prefixed with + and then disable the specs prefixed with -.
	for _, spec := range enable {
		spec.Enable()
	}
	for _, spec := range disable {
		spec.Disable()
	}
	*s = specsFlag(value)
	return nil
}

var specsFlagValue specsFlag

func init() {
	flag.Var(&specsFlagValue, "specs", "A comma-separated list of specs to be tested. Accepts a spec (test only this spec), a +spec (test also this immature spec), or a -spec (do not test this mature spec). Defaults to all mature specs.")
}
