package infrastructure_test

import (
	"testing"

	"service-short-link/internal/domain"
	"service-short-link/internal/infrastructure/services"

	"github.com/stretchr/testify/assert"
)

func TestQRCodeGenerator_Basic(t *testing.T) {
	gen := services.NewQRCodeGenerator()
	png, err := gen.Generate("https://example.com", 128)
	assert.NoError(t, err)
	assert.Greater(t, len(png), 0)
}

func TestQRCodeGenerator_WithOptions(t *testing.T) {
	gen := services.NewQRCodeGenerator()
	png, err := gen.GenerateWithOptions("https://example.com", 0, domain.QRCodeOptions{Size: 200, ErrorLevel: domain.QRErrorLevelHigh, Format: domain.QRFormatPNG})
	assert.NoError(t, err)
	assert.Greater(t, len(png), 0)
}

func TestQRCodeGenerator_UnsupportedFormats(t *testing.T) {
	gen := services.NewQRCodeGenerator()
	_, err := gen.GenerateWithOptions("https://example.com", 128, domain.QRCodeOptions{Format: domain.QRFormatJPEG})
	assert.Error(t, err)
	_, err = gen.GenerateWithOptions("https://example.com", 128, domain.QRCodeOptions{Format: domain.QRFormatSVG})
	assert.Error(t, err)
}
