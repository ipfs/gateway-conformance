package main

import (
	"testing"
)

// Test cases for isSubdomainPresetEnabled function
func TestIsSubdomainPresetEnabled(t *testing.T) {
	tests := []struct {
		specs           string
		expectedEnabled bool
		description     string
	}{
		// Test case 1: Empty specs and subdomain preset is enabled by default
		{
			specs:           "",
			expectedEnabled: true,
			description:     "Empty specs and subdomain preset is enabled by default",
		},
		// Test case 2: User provides "-subdomain", should be disabled explicitly
		{
			specs:           "-subdomain-gateway",
			expectedEnabled: false,
			description:     "User provides '-subdomain', should be disabled explicitly",
		},
		// Test case 3: User provides "+subdomain", should be enabled explicitly
		{
			specs:           "+subdomain-gateway",
			expectedEnabled: true,
			description:     "User provides '+subdomain', should be enabled explicitly",
		},
		// Test case 4: User provides "+other", should not affect subdomain preset
		{
			specs:           "+proxy-gateway",
			expectedEnabled: true,
			description:     "User provides '+proxy-gateway', should not affect subdomain-gateway preset default",
		},
		// Test case 5: User provides "other", subdomain preset should be enabled by default
		{
			specs:           "path-gateway",
			expectedEnabled: false,
			description:     "User provides 'path-gateway', subdomain preset should be disabled due to explicit (manual) list",
		},
		// Test case 6: User provides "-other,+subdomain", should be enabled due to +subdomain
		{
			specs:           "-path-gateway,+subdomain-gateway",
			expectedEnabled: true,
			description:     "User provides '-path-gateway,+subdomain-gateway', should be enabled due to +subdomain-gateway",
		},
		// Test case 7: User provides "+other,-subdomain", should be disabled due to -subdomain
		{
			specs:           "+path-gateway,-subdomain-gateway",
			expectedEnabled: false,
			description:     "User provides '+path-gateway,-subdomain-gateway', should be disabled due to -subdomain-gateway",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualEnabled := isSubdomainPresetEnabled(test.specs)
			if actualEnabled != test.expectedEnabled {
				t.Errorf("Expected isSubdomainPresetEnabled(%q) to be %v, but got %v",
					test.specs, test.expectedEnabled, actualEnabled)
			}
		})
	}
}
