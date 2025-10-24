package domain_test

import (
	"testing"

	domain "service-short-link/internal/domain"
	"service-short-link/internal/tests/mocks"

	"github.com/stretchr/testify/assert"
)

func TestServices_Basic(t *testing.T) {
	gen := &mocks.ShortCodeGeneratorMock{
		GenerateWithLengthFunc: func(length int) string { return "abcde" },
		IsValidFunc:            func(shortCode string) bool { return len(shortCode) >= 5 },
	}
	assert.Equal(t, "abcde", gen.GenerateWithLength(5))
	assert.True(t, gen.IsValid("abcde"))
	assert.False(t, gen.IsValid("abc"))

	qrGen := &mocks.QRCodeGeneratorMock{
		GenerateFunc: func(data string, size int) ([]byte, error) { return []byte("qrdata"), nil },
		GenerateWithOptionsFunc: func(data string, size int, options domain.QRCodeOptions) ([]byte, error) {
			return []byte("qrdataopt"), nil
		},
	}
	data, err := qrGen.Generate("test", 100)
	assert.NoError(t, err)
	assert.Equal(t, []byte("qrdata"), data)
	data, err = qrGen.GenerateWithOptions("test", 100, domain.QRCodeOptions{})
	assert.NoError(t, err)
	assert.Equal(t, []byte("qrdataopt"), data)

	validator := &mocks.URLValidatorMock{
		IsValidFunc:   func(url string) bool { return url == "https://valid.com" },
		NormalizeFunc: func(url string) (string, error) { return url, nil },
		IsSafeURLFunc: func(url string) bool { return true },
	}
	assert.True(t, validator.IsValid("https://valid.com"))
	assert.False(t, validator.IsValid("invalid"))
	norm, err := validator.Normalize("https://valid.com")
	assert.NoError(t, err)
	assert.Equal(t, "https://valid.com", norm)
	assert.True(t, validator.IsSafeURL("any"))
}
