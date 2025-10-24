package mocks

type ShortCodeGeneratorMock struct {
	GenerateWithLengthFunc func(length int) string
	IsValidFunc            func(shortCode string) bool
}

func (m *ShortCodeGeneratorMock) GenerateWithLength(length int) string {
	if m.GenerateWithLengthFunc != nil {
		return m.GenerateWithLengthFunc(length)
	}
	return ""
}

func (m *ShortCodeGeneratorMock) IsValid(shortCode string) bool {
	if m.IsValidFunc != nil {
		return m.IsValidFunc(shortCode)
	}
	return false
}
