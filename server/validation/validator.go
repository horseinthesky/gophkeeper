package validation

import (
	"fmt"
	"regexp"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString

	ErrInvalidUsername = fmt.Errorf("must contain only lowercase letters, digits, or underscore")
)

type ErrValueIsTooShortOrTooLong struct {
	min, max int
	err      error
}

func (e ErrValueIsTooShortOrTooLong) Error() string {
	return fmt.Sprintf("must contain from %d-%d characters", e.min, e.max)
}

func NewValueError(min, max int, err error) error {
	return ErrValueIsTooShortOrTooLong{min, max, err}
}

func validateString(value string, minLength int, maxLength int) error {
	n := len(value)
	if n < minLength || n > maxLength {
		return NewValueError(minLength, maxLength, fmt.Errorf(""))
	}
	return nil
}

func ValidateUsername(value string) error {
	if err := validateString(value, 3, 30); err != nil {
		return err
	}
	if !isValidUsername(value) {
		return ErrInvalidUsername
	}
	return nil
}

func ValidatePassword(value string) error {
	return validateString(value, 6, 50)
}
