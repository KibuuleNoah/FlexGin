package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
)

func PrintPretty(data interface{}) {
	// MarshalIndent adds spacing and line breaks to make it human-readable
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting data: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateID(length int) (string, error) {
	bytes := make([]byte, length)

	// Read cryptographically secure random bytes
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Map random bytes to our specific character set
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes), nil
}

func ToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}

func ToNullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{
		Bool:  *b,
		Valid: true,
	}
}
