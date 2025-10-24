package mocks

import "service-short-link/internal/domain"

type LinkRepositoryMock struct {
	CreateFunc              func(link *domain.Link) error
	GetByShortCodeFunc      func(shortCode string) (*domain.Link, error)
	GetForRedirectFunc      func(shortCode string) (*domain.LinkForRedirect, error)
	ExistsByShortCodeFunc   func(shortCode string) (bool, error)
	IncrementClickCountFunc func(linkID uint64) error
	UpdateLastAccessedFunc  func(linkID uint64) error
}

func (m *LinkRepositoryMock) Create(link *domain.Link) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(link)
	}
	return nil
}
func (m *LinkRepositoryMock) GetByShortCode(shortCode string) (*domain.Link, error) {
	if m.GetByShortCodeFunc != nil {
		return m.GetByShortCodeFunc(shortCode)
	}
	return nil, nil
}
func (m *LinkRepositoryMock) GetForRedirect(shortCode string) (*domain.LinkForRedirect, error) {
	if m.GetForRedirectFunc != nil {
		return m.GetForRedirectFunc(shortCode)
	}
	return nil, nil
}
func (m *LinkRepositoryMock) ExistsByShortCode(shortCode string) (bool, error) {
	if m.ExistsByShortCodeFunc != nil {
		return m.ExistsByShortCodeFunc(shortCode)
	}
	return false, nil
}
func (m *LinkRepositoryMock) IncrementClickCount(linkID uint64) error {
	if m.IncrementClickCountFunc != nil {
		return m.IncrementClickCountFunc(linkID)
	}
	return nil
}
func (m *LinkRepositoryMock) UpdateLastAccessed(linkID uint64) error {
	if m.UpdateLastAccessedFunc != nil {
		return m.UpdateLastAccessedFunc(linkID)
	}
	return nil
}
