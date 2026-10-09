package application

import (
	"strings"
	"time"
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Обёртки, чтобы легко мокать в тестах.
var timeNow = time.Now
