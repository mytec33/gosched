package types

import (
	"encoding/json"
	"errors"
	"testing"

	"git.sr.ht/~mytec/gosched/internal/errs"
)

type TestConfig struct {
	Value ConfiguredInt `json:"value"`
}

func TestRequireNonNegative(t *testing.T) {
	tests := []struct {
		name      string
		valueJSON string
		wantErr   error
		wantValue int
	}{
		{name: "zero ok", valueJSON: `{ "value": 0 }`, wantErr: nil, wantValue: 0},
		{name: "greater than zero ok", valueJSON: `{ "value": 100 }`, wantErr: nil, wantValue: 100},
		{name: "less than zero", valueJSON: `{ "value": -1 }`, wantErr: errs.ErrNegativeNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ci TestConfig

			err := json.Unmarshal([]byte(tt.valueJSON), &ci)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("got error %v, want nil", err)
				}
				if got := ci.Value.Int(); got != tt.wantValue { // adjust accessor to your actual field/method names
					t.Fatalf("got value %d, want %d", got, tt.wantValue)
				}
				return
			}

			if err == nil {
				t.Fatalf("got nil error, want %v", tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}
