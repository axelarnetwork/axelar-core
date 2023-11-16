package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHexDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   "",
			want:    []byte{},
			wantErr: false,
		},
		{
			name:    "valid input with 0x prefix",
			input:   "0x68656c6c6f",
			want:    []byte("hello"),
			wantErr: false,
		},
		{
			name:    "valid input without 0x prefix",
			input:   "68656c6c6f",
			want:    []byte("hello"),
			wantErr: false,
		},
		{
			name:    "invalid input with odd number of characters",
			input:   "68656c6c6",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid input with non-hex characters",
			input:   "68656c6c6z",
			want:    nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := HexDecode(test.input)
			if test.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, test.want, got)
		})
	}
}
