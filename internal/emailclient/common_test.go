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
			t.Run("single-recipients", func(t *testing.T) {
				var options []EmailOption
				for _, r := range c.ccRecipients {
					options = append(options, WithCCRecipient(r))
				}
				for _, r := range c.bccRecipients {
					options = append(options, WithBCCRecipient(r))
				}
				options = append(options, WithBody(c.body))

				result := processOptions(options...)

				compareOptions(t, c.expected, result)
			})
			t.Run("recipients-as-groups", func(t *testing.T) {
				result := processOptions(
					WithCCRecipients(c.ccRecipients),
					WithBCCRecipients(c.bccRecipients),
					WithBody(c.body),
				)

				compareOptions(t, c.expected, result)
			})
		})
	}
}

func compareOptions(t *testing.T, o1, o2 *emailOptions) {
	t.Helper()

	if len(o1.ccRecipients) == len(o2.ccRecipients) {
		for i := range o1.ccRecipients {
			if o1.ccRecipients[i] != o2.ccRecipients[i] {
				t.Error("CC recipients do not match")
			}
		}
	} else {
		t.Error("CC recipients do not match")
	}

	if len(o1.bccRecipients) == len(o2.bccRecipients) {
		for i := range o1.bccRecipients {
			if o1.bccRecipients[i] != o2.bccRecipients[i] {
				t.Error("BCC recipients do not match")
			}
		}
	} else {
		t.Error("BCC recipients do not match")
	}

	if o1.body != o2.body {
		t.Errorf("body does not match")
	}
}
