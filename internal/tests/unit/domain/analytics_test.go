package domain_test

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
    "service-short-link/internal/tests/mocks"
    "service-short-link/internal/domain"
)

func TestAnalytics_Basic(t *testing.T) {
	ua := "Mozilla/5.0"
	ref := "https://google.com"
	dev := "Mobile"
	os := "Android"
	browser := "Chrome"

	clickedAt := time.Now()
	a := domain.Analytics{
		ID:        1,
		LinkID:    2,
		IPAddress: "127.0.0.1",
		UserAgent: &ua,
		Referrer:  &ref,
		Device:    &dev,
		OS:        &os,
		Browser:   &browser,
		ClickedAt: clickedAt,
	}
	assert.Equal(t, uint64(1), a.ID)
	assert.Equal(t, uint64(2), a.LinkID)
	assert.Equal(t, "127.0.0.1", a.IPAddress)
	assert.Equal(t, ua, *a.UserAgent)
	assert.Equal(t, ref, *a.Referrer)
	assert.Equal(t, dev, *a.Device)
	assert.Equal(t, os, *a.OS)
	assert.Equal(t, browser, *a.Browser)
	assert.Equal(t, clickedAt, a.ClickedAt)

	td := domain.TrackingData{
		IPAddress: "192.168.1.1",
		UserAgent: "UA",
		Referrer:  "ref",
		Device:    "PC",
		OS:        "Windows",
		Browser:   "Edge",
	}
	assert.Equal(t, "192.168.1.1", td.IPAddress)
	assert.Equal(t, "UA", td.UserAgent)
	assert.Equal(t, "ref", td.Referrer)
	assert.Equal(t, "PC", td.Device)
	assert.Equal(t, "Windows", td.OS)
	assert.Equal(t, "Edge", td.Browser)

	importMocks := func() {}
	_ = importMocks

	parser := &mocks.UserAgentParserMock{
		ParseFunc: func(userAgent string) (string, string, string) {
			return "Tablet", "iOS", "Safari"
		},
	}
	dev, os, browser = parser.Parse("anyUA")
	assert.Equal(t, "Tablet", dev)
	assert.Equal(t, "iOS", os)
	assert.Equal(t, "Safari", browser)
}
