package utils

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHexDecode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []byte
		expectedErr bool
	}{
		{
			name:        "empty input",
			input:       "",
			expected:    []byte{},
			expectedErr: false,
		},
		{
			name:        "empty input with 0x prefix",
			input:       "0x",
			expected:    []byte{},
			expectedErr: false,
		},
		{
			name:        "valid input with 0x prefix",
			input:       "0x68656c6c6f",
			expected:    []byte("hello"),
			expectedErr: false,
		},
		{
			name:        "valid input without 0x prefix",
			input:       "68656c6c6f",
			expected:    []byte("hello"),
			expectedErr: false,
		},
		{
			name:        "invalid input with odd number of characters",
			input:       "68656c6c6",
			expected:    nil,
			expectedErr: true,
		},
		{
			name:        "invalid input with non-hex characters",
			input:       "68656c6c6z",
			expected:    nil,
			expectedErr: true,
		},
		{
			name:        "valid input with uppercase hex characters",
			input:       "0x48656C6C6F",
			expected:    []byte("Hello"),
			expectedErr: false,
		},
		{
			name:        "valid input with mixed case hex characters",
			input:       "0x48656c6C6F",
			expected:    []byte("Hello"),
			expectedErr: false,
		},
		{
			name:        "input with leading zeros",
			input:       "0x00000068656c6c6f",
			expected:    append([]byte{0, 0, 0}, []byte("hello")...),
			expectedErr: false,
		},
		{
			name:        "input with trailing zeros",
			input:       "0x68656c6c6f000000",
			expected:    append([]byte("hello"), 0, 0, 0),
			expectedErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decoded, err := HexDecode(test.input)
			if test.expectedErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, test.expected, decoded)

			encoded := HexEncode(decoded)
			assert.Equal(t, "0x"+hex.EncodeToString(decoded), encoded)
		})
	}
}
