package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		expected error
	}{
		{
			name:     "bob",
			expected: nil,
		},
		{
			name:     "bo",
			expected: ErrValueIsTooShortOrTooLong{3, 30, fmt.Errorf("")},
		},
		{
			name:     "bo-bo",
			expected: ErrInvalidUsername,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("test %s", tt.name), func(t *testing.T) {
			err := ValidateUsername(tt.name)
			require.Equal(t, tt.expected, err)
		})
	}
}
