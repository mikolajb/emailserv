package main

import "testing"

func Test_validate(t *testing.T) {
	m := &Message{
		Sender:       "abc",
		Recipients:   []string{"def", "def@abc.com"},
		CCRecipients: []string{"ghi@abc.com", "ghi"},
	}

	expected := map[string]string{
		"sender":        "not a valid email",
		"recipients":    "contains an invalid email address",
		"cc_recipients": "contains an invalid email address",
	}

	for _, ve := range validate(m) {
		message, ok := expected[ve.Field]
		if ok {
			if message != ve.Error {
				t.Errorf("expected message '%s' for field '%s' but got '%s'", message, ve.Field, ve.Error)
			}
			delete(expected, ve.Field)
		} else {
			t.Errorf("unexpected error '%s' for field '%s'", ve.Error, ve.Field)
		}
	}

	for field, error := range expected {
		t.Errorf("invalid field %s was accepted, expected error: %s", field, error)
	}
}
