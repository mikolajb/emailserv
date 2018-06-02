package emailclient

import "testing"

func Test_processOptions(t *testing.T) {
	cases := map[string]struct {
		ccRecipients  []string
		bccRecipients []string
		body          string
		expected      *emailOptions
	}{
		"nothing": {
			expected: &emailOptions{},
		},
		"some": {
			bccRecipients: []string{"a"},
			expected: &emailOptions{
				bccRecipients: []string{"a"},
			},
		},
		"everything": {
			ccRecipients:  []string{"a", "b", "c"},
			bccRecipients: []string{"d", "e", "f", "g"},
			body:          "hijkl",
			expected: &emailOptions{
				ccRecipients:  []string{"a", "b", "c"},
				bccRecipients: []string{"d", "e", "f", "g"},
				body:          "hijkl",
			},
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			var options []EmailOption
			for _, r := range c.ccRecipients {
				options = append(options, WithCCRecipient(r))
			}
			for _, r := range c.bccRecipients {
				options = append(options, WithBCCRecipient(r))
			}
			options = append(options, WithBody(c.body))

			result := processOptions(options...)

			if len(c.expected.ccRecipients) == len(result.ccRecipients) {
				for i := range c.expected.ccRecipients {
					if c.expected.ccRecipients[i] != result.ccRecipients[i] {
						t.Error("CC recipients do not match")
					}
				}
			} else {
				t.Error("CC recipients do not match")
			}

			if len(c.expected.bccRecipients) == len(result.bccRecipients) {
				for i := range c.expected.bccRecipients {
					if c.expected.bccRecipients[i] != result.bccRecipients[i] {
						t.Error("BCC recipients do not match")
					}
				}
			} else {
				t.Error("BCC recipients do not match")
			}

			if c.expected.body != result.body {
				t.Errorf("body does not match")
			}
		})
	}
}
