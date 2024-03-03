package util

import (
	"regexp"
	"strings"
	"unicode"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ValidatePassword(password string) (string, bool) {
	var (
		hasUpper     = false
		hasLower     = false
		hasNumber    = false
		hasSpecial   = false
		specialRunes = "!@#$%^&*()_+"
	)

	if len(password) < 8 {
		return "Password must be at least 8 characters", false
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	for _, char := range specialRunes {
		if strings.ContainsRune(password, char) {
			hasSpecial = true
		}
	}

	if !hasUpper {
		return "Password must contain at least one uppercase letter", false
	}

	if !hasLower {
		return "Password must contain at least one lowercase letter", false
	}

	if !hasNumber {
		return "Password must contain at least one number", false
	}

	if !hasSpecial {
		return "Password must contain at least one special character", false
	}

	return "", true
}
