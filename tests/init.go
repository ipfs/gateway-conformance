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
			enable = append(enable, spec)
		} else if strings.HasPrefix(name, "-") {
			disable = append(disable, spec)
		} else {
			only = append(only, spec)
		}
	}
	if len(only) > 0 {
		// disable all specs
		for _, spec := range specs.All() {
			spec.Disable()
		}
		// enable only the specified specs
		for _, spec := range only {
			spec.Enable()
		}
	} else {
		// enable the specified specs
		for _, spec := range enable {
			spec.Enable()
		}
		// disable the specified specs
		for _, spec := range disable {
			spec.Disable()
		}
	}
	*s = specsFlag(value)
	return nil
}

var specsFlagValue specsFlag

func init() {
	flag.Var(&specsFlagValue, "specs", "comma-separated list of specs to test")
}
