package mocks

type UserAgentParserMock struct {
	ParseFunc func(userAgent string) (device, os, browser string)
}

func (m *UserAgentParserMock) Parse(userAgent string) (device, os, browser string) {
	if m.ParseFunc != nil {
		return m.ParseFunc(userAgent)
	}
	return "", "", ""
}
