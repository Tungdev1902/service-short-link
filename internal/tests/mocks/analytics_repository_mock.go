package mocks

import "service-short-link/internal/domain"

type AnalyticsRepositoryMock struct {
	CreateFunc func(analytics *domain.Analytics) error
}

func (m *AnalyticsRepositoryMock) Create(analytics *domain.Analytics) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(analytics)
	}
	return nil
}
