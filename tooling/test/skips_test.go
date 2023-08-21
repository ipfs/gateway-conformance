package test

import "testing"

func TestSplit(t *testing.T) {
	type test struct {
		name     string
		expected []string
	}

	tests := []test{
		{
			name: `TestA`,
			expected: []string{
				"TestA",
			},
		},
		{
			name: `TestA/With/Path`,
			expected: []string{
				"TestA",
				"With",
				"Path",
			},
		},
		{
			name: `TestA/With/Path/And/Slash/\\`,
			expected: []string{
				"TestA",
				"With",
				"Path",
				"And",
				"Slash",
				`\`,
			},
		},
		{
			name: `TestA/With/Pa\th\/An\\d/Slash/\\\\`,
			expected: []string{
				"TestA",
				"With",
				`Pa\th/An\d`,
				"Slash",
				`\\`,
			},
		},
	}

	for _, test := range tests {
		got := split(test.name)
		if len(got) != len(test.expected) {
			t.Errorf("split(%s) = %v, want %v", test.name, got, test.expected)
			continue
		}

		for i := range got {
			if got[i] != test.expected[i] {
				t.Errorf("split(%s) = %v, want %v (%s != %s)", test.name, got, test.expected, got[i], test.expected[i])
				break
			}
		}
	}
}

func TestSkips(t *testing.T) {
	type test struct {
		name     string
		skips    []string
		expected bool
	}

	tests := []test{
		{
			name:     "TestNeverSkipped",
			skips:    []string{},
			expected: false,
		},
		{
			name: "TestAlwaysSkipped",
			skips: []string{
				"TestAlwaysSkipped",
			},
			expected: true,
		},
		{
			name: "TestA",
			skips: []string{
				"TestA/With/Path",
			},
			expected: false,
		},
		{
			name: "TestA/With/Path",
			skips: []string{
				"TestA",
			},
			expected: true,
		},
		{
			name: "TestA/With/Path",
			skips: []string{
				"TestA/With.*out/Path",
			},
			expected: false,
		},
		{
			name: "TestA/With/Path",
			skips: []string{
				"Test.*",
			},
			expected: true,
		},
		{
			name: "TestA/With/Path",
			skips: []string{
				"Test.*/With/.*",
			},
			expected: true,
		},
		{
			name: "TestA/Without/Path",
			skips: []string{
				"Test.*/With/.*",
			},
			expected: false,
		},
		{
			name: "TestA/With/Path",
			skips: []string{
				"Test.*/With",
			},
			expected: true,
		},
	}

	for _, test := range tests {
		got := isSkipped(test.name, test.skips)
		if got != test.expected {
			t.Errorf("isSkipped(%s, %v) = %v, want %v", test.name, test.skips, got, test.expected)
		}
	}
}
