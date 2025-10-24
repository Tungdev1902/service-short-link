package infrastructure_test

import (
	"testing"

	"service-short-link/internal/infrastructure/services"

	"github.com/stretchr/testify/assert"
)

func TestUserAgentParser_Parse(t *testing.T) {
	p := services.NewUserAgentParser()

	// iPhone Safari
	d, o, b := p.Parse("Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
	assert.Equal(t, "Mobile", d)
	assert.Contains(t, o, "iOS")
	assert.Contains(t, b, "Safari")

	// Windows Chrome
	d, o, b = p.Parse("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	assert.Equal(t, "Desktop", d)
	assert.Contains(t, o, "Windows")
	assert.Contains(t, b, "Chrome")

	// Android Chrome
	d, o, b = p.Parse("Mozilla/5.0 (Linux; Android 11; SM-T870) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36")
	assert.Equal(t, "Tablet", d)
	assert.Contains(t, o, "Android 11")
	assert.Contains(t, b, "Chrome")

	// iPad Safari
	d, o, b = p.Parse("Mozilla/5.0 (iPad; CPU OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	assert.Equal(t, "Tablet", d)
	assert.Contains(t, o, "iOS (iPad)")
	assert.Contains(t, b, "Safari")

	// Edge
	d, o, b = p.Parse("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59")
	assert.Equal(t, "Desktop", d)
	assert.Contains(t, o, "Windows")
	assert.Contains(t, b, "Edge 91.0.864.59")

	// Opera
	d, o, b = p.Parse("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 OPR/77.0.4054.172")
	assert.Equal(t, "Desktop", d)
	assert.Contains(t, o, "Windows")
	assert.Contains(t, b, "Opera 77.0.4054.172")

	// Firefox
	d, o, b = p.Parse("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0")
	assert.Equal(t, "Desktop", d)
	assert.Contains(t, o, "Windows")
	assert.Contains(t, b, "Firefox 89.0")

	// Internet Explorer
	d, o, b = p.Parse("Mozilla/5.0 (Windows NT 10.0; Trident/7.0; rv:11.0) like Gecko")
	assert.Equal(t, "Desktop", d)
	assert.Contains(t, o, "Windows")
	assert.Contains(t, b, "Internet Explorer 11.0")

	// Unknown
	d, o, b = p.Parse("")
	assert.Equal(t, "Unknown", d)
	assert.Equal(t, "Unknown", o)
	assert.Equal(t, "Unknown", b)
}
