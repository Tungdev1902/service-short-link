package mocks

import "service-short-link/internal/domain"

type QRCodeGeneratorMock struct {
    GenerateFunc           func(data string, size int) ([]byte, error)
    GenerateWithOptionsFunc func(data string, size int, options domain.QRCodeOptions) ([]byte, error)
}

func (m *QRCodeGeneratorMock) Generate(data string, size int) ([]byte, error) {
    if m.GenerateFunc != nil {
        return m.GenerateFunc(data, size)
    }
    return nil, nil
}

func (m *QRCodeGeneratorMock) GenerateWithOptions(data string, size int, options domain.QRCodeOptions) ([]byte, error) {
    if m.GenerateWithOptionsFunc != nil {
        return m.GenerateWithOptionsFunc(data, size, options)
    }
    return nil, nil
}


