package mocks

import "service-short-link/internal/domain"

type LinkUseCaseMock struct {
	CreateLinkFunc     func(req *domain.CreateLinkRequest, platform, role, channelCode *string) (*domain.CreateLinkResponse, error)
	RedirectLinkFunc   func(shortCode string, trackingData *domain.TrackingData) (string, error)
	GenerateQRCodeFunc func(shortCode string) ([]byte, error)
}

func (m *LinkUseCaseMock) CreateLink(req *domain.CreateLinkRequest, platform, role, channelCode *string) (*domain.CreateLinkResponse, error) {
	if m.CreateLinkFunc != nil {
		return m.CreateLinkFunc(req, platform, role, channelCode)
	}
	return nil, nil
}

func (m *LinkUseCaseMock) RedirectLink(shortCode string, trackingData *domain.TrackingData) (string, error) {
	if m.RedirectLinkFunc != nil {
		return m.RedirectLinkFunc(shortCode, trackingData)
	}
	return "", nil
}

func (m *LinkUseCaseMock) GenerateQRCode(shortCode string) ([]byte, error) {
	if m.GenerateQRCodeFunc != nil {
		return m.GenerateQRCodeFunc(shortCode)
	}
	return nil, nil
}
