package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GenerateDocNumber generates a document number with LAL prefix and timestamp
// Format: PREFIX + DDMMYYYYHHMMSS + XXX (Random)
func GenerateDocNumber(prefix string) string {
	now := time.Now()
	// Format: DDMMYYYYHHMMSS
	timestamp := now.Format("02012006150405")
	
	// Add 3 random digits to ensure uniqueness within the same second
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(1000))
	return fmt.Sprintf("%s%s%03d", prefix, timestamp, randomNum)
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
