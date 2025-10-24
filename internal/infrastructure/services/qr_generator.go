package services

import (
	"fmt"

	"service-short-link/internal/domain"
	"service-short-link/pkg/logger"

	"github.com/skip2/go-qrcode"
)

type qrCodeGenerator struct {
	defaultSize int
}

// NewQRCodeGenerator creates a new QR code generator
func NewQRCodeGenerator() domain.QRCodeGenerator {
	return &qrCodeGenerator{
		defaultSize: 256,
	}
}

// Generate creates a QR code with default options
func (g *qrCodeGenerator) Generate(data string, size int) ([]byte, error) {
	if size <= 0 {
		size = g.defaultSize
	}

	qrCode, err := qrcode.New(data, qrcode.High)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "QRGenerator.Generate: failed to create QR code", "data="+data, "size="+fmt.Sprintf("%d", size), "error_type=qr_code_creation_failed")
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	pngData, err := qrCode.PNG(size)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "QRGenerator.Generate: failed to generate PNG", "data="+data, "size="+fmt.Sprintf("%d", size), "error_type=png_generation_failed")
		return nil, fmt.Errorf("failed to generate PNG: %w", err)
	}

	return pngData, nil
}

// GenerateWithOptions creates a QR code with custom options
func (g *qrCodeGenerator) GenerateWithOptions(data string, size int, options domain.QRCodeOptions) ([]byte, error) {
	if size <= 0 {
		size = options.Size
	}
	if size <= 0 {
		size = g.defaultSize
	}

	var errorLevel qrcode.RecoveryLevel
	switch options.ErrorLevel {
	case domain.QRErrorLevelLow:
		errorLevel = qrcode.Low
	case domain.QRErrorLevelMedium:
		errorLevel = qrcode.Medium
	case domain.QRErrorLevelQuartile:
		errorLevel = qrcode.High
	case domain.QRErrorLevelHigh:
		errorLevel = qrcode.Highest
	default:
		errorLevel = qrcode.Medium
	}

	qrCode, err := qrcode.New(data, errorLevel)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "QRGenerator.GenerateWithOptions: failed to create QR code", "data="+data, "size="+fmt.Sprintf("%d", options.Size), "error_level="+fmt.Sprintf("%d", options.ErrorLevel), "error_type=qr_code_creation_failed")
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	switch options.Format {
	case domain.QRFormatPNG:
		fallthrough
	default:
		pngData, err := qrCode.PNG(size)
		if err != nil {
			logger.ErrorWithCockroachSimple(err, "QRGenerator.GenerateWithOptions: failed to generate PNG", "data="+data, "size="+fmt.Sprintf("%d", size), "format="+fmt.Sprintf("%d", options.Format), "error_type=png_generation_failed")
			return nil, fmt.Errorf("failed to generate PNG: %w", err)
		}
		return pngData, nil
	case domain.QRFormatJPEG:
		return nil, fmt.Errorf("JPEG format not implemented")
	case domain.QRFormatSVG:
		return nil, fmt.Errorf("SVG format not implemented")
	}
}

// generateOptimizedQR creates an optimized QR code for web usage
func (g *qrCodeGenerator) generateOptimizedQR(data string) ([]byte, error) {
	qrCode, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "QRGenerator.generateOptimizedQR: failed to create optimized QR code", "data="+data, "error_type=qr_code_creation_failed")
		return nil, fmt.Errorf("failed to create optimized QR code: %w", err)
	}

	pngData, err := qrCode.PNG(256)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "QRGenerator.generateOptimizedQR: failed to generate optimized PNG", "data="+data, "error_type=png_generation_failed")
		return nil, fmt.Errorf("failed to generate optimized PNG: %w", err)
	}

	return pngData, nil
}

// ValidateQRData checks if the data is suitable for QR code generation
func (g *qrCodeGenerator) ValidateQRData(data string) error {
	if data == "" {
		return fmt.Errorf("QR code data cannot be empty")
	}

	if len(data) > 1000 {
		return fmt.Errorf("QR code data too long: %d characters (max 1000)", len(data))
	}

	return nil
}
