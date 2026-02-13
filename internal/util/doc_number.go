package util

import (
	"fmt"
	"time"
)

// GenerateDocNumber generates a document number with LAL prefix and timestamp
// Format: LAL + DDMMYYYYHHMM (e.g., LAL130220260008 = 13/02/2026 00:08)
func GenerateDocNumber(prefix string) string {
	now := time.Now()
	// Format: DDMMYYYYHHMM
	timestamp := now.Format("020120061504") // Day Month Year Hour Minute - Fixed format layout
	return fmt.Sprintf("%s%s", prefix, timestamp)
}

// GenerateLALDocNumber generates a document number with LAL prefix
func GenerateLALDocNumber() string {
	return GenerateDocNumber("LAL")
}

// GenerateIncomingNumber generates an incoming document number
func GenerateIncomingNumber() string {
	return GenerateDocNumber("IN")
}

// GenerateOutgoingNumber generates an outgoing document number
func GenerateOutgoingNumber() string {
	return GenerateDocNumber("OUT")
}
