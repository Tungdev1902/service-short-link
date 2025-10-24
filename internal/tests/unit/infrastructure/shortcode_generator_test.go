package infrastructure_test

import (
	"testing"

	"service-short-link/internal/infrastructure/services"

	"github.com/stretchr/testify/assert"
)

func TestShortCodeGenerator_Validity(t *testing.T) {
	gen := services.NewShortCodeGenerator()
	assert.True(t, gen.IsValid("abc12"))
	assert.False(t, gen.IsValid("a"))
	assert.False(t, gen.IsValid("admin"))
}

func BenchmarkShortCodeGenerator_Generate(b *testing.B) {
	gen := services.NewShortCodeGenerator()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = gen.GenerateWithLength(7)
	}
}

func TestShortCodeGenerator_GenerateWithBounds(t *testing.T) {
	gen := services.NewShortCodeGenerator()
	code := gen.GenerateWithLength(5)
	assert.Len(t, code, 5)
	code = gen.GenerateWithLength(50) // should clamp to max
	assert.LessOrEqual(t, len(code), 10)
}
