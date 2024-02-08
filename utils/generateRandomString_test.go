package utils

import (
	"context"
	"testing"
)

func TestGenerateRandomStringWithContext_GeneratesStringOfCorrectLength(t *testing.T) {
	ctx := context.Background()
	length := 10
	result := GenerateRandomStringWithContext(ctx, length)
	if len(result) != length {
		t.Errorf("Expected string of length %d, got: %d", length, len(result))
	}
}

func TestGenerateRandomStringWithContext_GeneratesDifferentStrings(t *testing.T) {
	ctx := context.Background()
	length := 10
	result1 := GenerateRandomStringWithContext(ctx, length)
	result2 := GenerateRandomStringWithContext(ctx, length)
	if result1 == result2 {
		t.Errorf("Expected different strings, got: %s and %s", result1, result2)
	}
}

func TestGenerateRandomStringWithContext_GeneratesEmptyStringForZeroLength(t *testing.T) {
	ctx := context.Background()
	length := 0
	result := GenerateRandomStringWithContext(ctx, length)
	if len(result) != length {
		t.Errorf("Expected empty string, got: %s", result)
	}
}

func TestGenerateRandomStringWithContext_GeneratesStringWithValidCharacters(t *testing.T) {
	ctx := context.Background()
	length := 10
	result := GenerateRandomStringWithContext(ctx, length)
	for _, char := range result {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			t.Errorf("Expected alphanumeric characters, got: %c", char)
		}
	}
}
