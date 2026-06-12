package domain

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{
			name: "ErrInvalidSchemaVersion",
			err:  ErrInvalidSchemaVersion,
			msg:  "invalid schema version: expected 1",
		},
		{
			name: "ErrInvalidProjectData",
			err:  ErrInvalidProjectData,
			msg:  "invalid project data",
		},
		{
			name: "ErrProjectNotFound",
			err:  ErrProjectNotFound,
			msg:  "project not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("expected message %q, got %q", tt.msg, tt.err.Error())
			}
			if !errors.Is(tt.err, tt.err) {
				t.Errorf("expected errors.Is to match itself")
			}
		})
	}
}
