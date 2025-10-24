package infrastructure_test

import (
	"testing"

	"service-short-link/internal/infrastructure/services"

	"github.com/stretchr/testify/assert"
)

func TestURLValidator_Normalize_And_Safety(t *testing.T) {
	v := services.NewURLValidator()

	// Normalize adds scheme and lowercases host
	norm, err := v.Normalize("Example.COM/path")
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/path", norm)

	// Valid and safe https
	assert.True(t, v.IsValid("https://example.com"))
	assert.True(t, v.IsSafeURL("https://example.com"))

	// Invalid / unsafe cases
	assert.False(t, v.IsValid("javascript:alert(1)"))
	assert.False(t, v.IsSafeURL("http://127.0.0.1"))
}

func BenchmarkURLValidator_Normalize(b *testing.B) {
	v := services.NewURLValidator()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = v.Normalize("Example.COM/path?q=1")
	}
}
