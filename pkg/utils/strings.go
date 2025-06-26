package utils

import (
	"fmt"
	"strings"
	"time"
)

// FormatMessage formats a message with timestamp
func FormatMessage(msg string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] %s", timestamp, msg)
}

// ToUpperCase converts a string to uppercase
func ToUpperCase(s string) string {
	return strings.ToUpper(s)
}

// ToLowerCase converts a string to lowercase
func ToLowerCase(s string) string {
	return strings.ToLower(s)
}

// TrimSpaces removes leading and trailing whitespace
func TrimSpaces(s string) string {
	return strings.TrimSpace(s)
}
