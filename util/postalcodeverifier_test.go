package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyPostalCodeValid(t *testing.T) {
	result := VerifyPostalCode("A1B2C3")
	assert.True(t, result)
}

func TestVerifyPostalCodeInvalidLength(t *testing.T) {
	result := VerifyPostalCode("A1B2C")
	assert.False(t, result)
}

func TestVerifyPostalCodeInvalidFormatNumbersInLetterPositions(t *testing.T) {
	result := VerifyPostalCode("123456")
	assert.False(t, result)
}

func TestVerifyPostalCodeInvalidFormatLettersInNumberPositions(t *testing.T) {
	result := VerifyPostalCode("ABCDEF")
	assert.False(t, result)
}

func TestVerifyPostalCodeEmptyString(t *testing.T) {
	result := VerifyPostalCode("")
	assert.False(t, result)
}
