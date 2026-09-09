package example

import (
	"testing"

	"github.com/behzadsh/monq"
)

// TestUnexportedModelPaths reads the paths generated from an unexported model.
//
// The value and the type behind it cannot share a name, so this file compiling at all is the point: it is what
// fails when the generator gives both the same identifier. The assertions below are the ordinary path checks.
func TestUnexportedModelPaths(t *testing.T) {
	var doc sessionDoc
	if !doc.ExpiresAt.IsZero() {
		t.Fatal("a zero sessionDoc has a non-zero ExpiresAt")
	}

	tests := []struct {
		name string
		got  monq.FieldPath
		want monq.FieldPath
	}{
		{name: "an id", got: sessionDocPaths.ID, want: "_id"},
		{name: "a tagged field", got: sessionDocPaths.UserID, want: "user_id"},
		{name: "a nested document", got: sessionDocPaths.Device.Platform, want: "device.platform"},
		{name: "a date is a leaf", got: sessionDocPaths.ExpiresAt, want: "expires_at"},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if tt.got != tt.want {
					t.Fatalf("got %q, want %q", tt.got, tt.want)
				}
			},
		)
	}
}
