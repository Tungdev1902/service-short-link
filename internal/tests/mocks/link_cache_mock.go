package mocks

import (
	"service-short-link/internal/domain"
	"time"
)

type LinkCacheMock struct {
	SetFunc       func(key string, link *domain.CachedLink, ttl time.Duration) error
	GetFunc       func(key string) (*domain.CachedLink, error)
	SetQRCodeFunc func(shortCode string, qrData []byte, ttl time.Duration) error
	GetQRCodeFunc func(shortCode string) ([]byte, error)
}

func (m *LinkCacheMock) Set(key string, link *domain.CachedLink, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, link, ttl)
	}
	return nil
}
func (m *LinkCacheMock) Get(key string) (*domain.CachedLink, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	return nil, nil
}
func (m *LinkCacheMock) SetQRCode(shortCode string, qrData []byte, ttl time.Duration) error {
	if m.SetQRCodeFunc != nil {
		return m.SetQRCodeFunc(shortCode, qrData, ttl)
	}
	return nil
}
func (m *LinkCacheMock) GetQRCode(shortCode string) ([]byte, error) {
	if m.GetQRCodeFunc != nil {
		return m.GetQRCodeFunc(shortCode)
	}
	return nil, nil
}
