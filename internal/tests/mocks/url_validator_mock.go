package mocks

type URLValidatorMock struct {
	IsValidFunc   func(url string) bool
	NormalizeFunc func(url string) (string, error)
	IsSafeURLFunc func(url string) bool
}

func (m *URLValidatorMock) IsValid(url string) bool {
	if m.IsValidFunc != nil {
		return m.IsValidFunc(url)
	}
	return false
}

func (m *URLValidatorMock) Normalize(url string) (string, error) {
	if m.NormalizeFunc != nil {
		return m.NormalizeFunc(url)
	}
	return url, nil
}

func (m *URLValidatorMock) IsSafeURL(url string) bool {
	if m.IsSafeURLFunc != nil {
		return m.IsSafeURLFunc(url)
	}
	return true
}
