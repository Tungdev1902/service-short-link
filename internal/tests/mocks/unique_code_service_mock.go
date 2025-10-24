package mocks

import "service-short-link/internal/domain"

type UniqueCodeServiceMock struct {
	GenerateUniqueCodeFunc func(generator domain.ShortCodeGenerator, repo domain.LinkRepository, length int) (string, error)
}

func (m *UniqueCodeServiceMock) GenerateUniqueCode(generator domain.ShortCodeGenerator, repo domain.LinkRepository, length int) (string, error) {
	if m.GenerateUniqueCodeFunc != nil {
		return m.GenerateUniqueCodeFunc(generator, repo, length)
	}
	return "abcde", nil
}
