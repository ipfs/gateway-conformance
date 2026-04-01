package main

import (
	"bytes"
	"strings"
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

func TestTransformSuiteEvents(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "suite pass renamed",
			input: `{"Action":"pass","Package":"example.com/pkg"}` + "\n",
			want:  `{"Action":"suite_pass","Package":"example.com/pkg"}` + "\n",
		},
		{
			name:  "suite fail renamed",
			input: `{"Action":"fail","Package":"example.com/pkg"}` + "\n",
			want:  `{"Action":"suite_fail","Package":"example.com/pkg"}` + "\n",
		},
		{
			name:  "test pass unchanged",
			input: `{"Action":"pass","Package":"example.com/pkg","Test":"TestFoo"}` + "\n",
			want:  `{"Action":"pass","Package":"example.com/pkg","Test":"TestFoo"}` + "\n",
		},
		{
			name:  "test fail unchanged",
			input: `{"Action":"fail","Package":"example.com/pkg","Test":"TestFoo"}` + "\n",
			want:  `{"Action":"fail","Package":"example.com/pkg","Test":"TestFoo"}` + "\n",
		},
		{
			name:  "other actions unchanged",
			input: `{"Action":"run","Package":"example.com/pkg"}` + "\n",
			want:  `{"Action":"run","Package":"example.com/pkg"}` + "\n",
		},
		{
			name: "mixed events",
			input: strings.Join([]string{
				`{"Action":"run","Package":"example.com/pkg","Test":"TestFoo"}`,
				`{"Action":"pass","Package":"example.com/pkg","Test":"TestFoo"}`,
				`{"Action":"pass","Package":"example.com/pkg"}`,
				"",
			}, "\n"),
			want: strings.Join([]string{
				`{"Action":"run","Package":"example.com/pkg","Test":"TestFoo"}`,
				`{"Action":"pass","Package":"example.com/pkg","Test":"TestFoo"}`,
				`{"Action":"suite_pass","Package":"example.com/pkg"}`,
				"",
			}, "\n"),
		},
		{
			name:  "output containing Test key does not prevent rename",
			input: `{"Action":"pass","Package":"example.com/pkg","Output":"no \"Test\": field found\n"}` + "\n",
			want:  `{"Action":"suite_pass","Package":"example.com/pkg","Output":"no \"Test\": field found\n"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(transformSuiteEvents([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestTransformWriterChunked(t *testing.T) {
	tests := []struct {
		name   string
		chunks []string
		want   string
	}{
		{
			name: "line split across two writes",
			chunks: []string{
				`{"Action":"pass","Pack`,
				`age":"example.com/pkg"}` + "\n",
			},
			want: `{"Action":"suite_pass","Package":"example.com/pkg"}` + "\n",
		},
		{
			name: "two lines in a single write",
			chunks: []string{
				`{"Action":"fail","Package":"example.com/pkg"}` + "\n" +
					`{"Action":"fail","Package":"example.com/pkg","Test":"TestFoo"}` + "\n",
			},
			want: `{"Action":"suite_fail","Package":"example.com/pkg"}` + "\n" +
				`{"Action":"fail","Package":"example.com/pkg","Test":"TestFoo"}` + "\n",
		},
		{
			name: "byte-at-a-time delivery",
			chunks: func() []string {
				line := `{"Action":"pass","Package":"example.com/pkg"}` + "\n"
				out := make([]string, len(line))
				for i, b := range []byte(line) {
					out[i] = string([]byte{b})
				}
				return out
			}(),
			want: `{"Action":"suite_pass","Package":"example.com/pkg"}` + "\n",
		},
		{
			name: "trailing data without newline is buffered",
			chunks: []string{
				`{"Action":"pass","Package":"example.com/pkg"}` + "\n",
				`{"Action":"incomplete"`,
			},
			want: `{"Action":"suite_pass","Package":"example.com/pkg"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tw := &transformWriter{w: &buf}
			for _, chunk := range tt.chunks {
				tw.Write([]byte(chunk))
			}
			got := buf.String()
			if got != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
