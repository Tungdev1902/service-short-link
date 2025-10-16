package services

import (
	"regexp"
	"strings"

	"service-short-link/internal/domain"
)

type userAgentParser struct {
	mobilePattern  *regexp.Regexp
	tabletPattern  *regexp.Regexp
	desktopPattern *regexp.Regexp

	iosPattern     *regexp.Regexp
	androidPattern *regexp.Regexp
	windowsPattern *regexp.Regexp
	macPattern     *regexp.Regexp
	linuxPattern   *regexp.Regexp

	chromePattern  *regexp.Regexp
	firefoxPattern *regexp.Regexp
	safariPattern  *regexp.Regexp
	edgePattern    *regexp.Regexp
	operaPattern   *regexp.Regexp
	iePattern      *regexp.Regexp
}

// NewUserAgentParser creates a new user agent parser
func NewUserAgentParser() domain.UserAgentParser {
	return &userAgentParser{
		mobilePattern:  regexp.MustCompile(`(?i)mobile|android.*mobile|blackberry|iphone|ipod|opera mini|iemobile`),
		tabletPattern:  regexp.MustCompile(`(?i)tablet|ipad`),
		desktopPattern: regexp.MustCompile(`(?i)windows|macintosh|linux|x11`),

		iosPattern:     regexp.MustCompile(`(?i)iPhone|iPad|iPod`),
		androidPattern: regexp.MustCompile(`(?i)Android`),
		windowsPattern: regexp.MustCompile(`(?i)Windows`),
		macPattern:     regexp.MustCompile(`(?i)Macintosh|Mac OS X`),
		linuxPattern:   regexp.MustCompile(`(?i)Linux|X11`),

		chromePattern:  regexp.MustCompile(`(?i)Chrome/([0-9.]+)`),
		firefoxPattern: regexp.MustCompile(`(?i)Firefox/([0-9.]+)`),
		safariPattern:  regexp.MustCompile(`(?i)Safari/([0-9.]+)`),
		edgePattern:    regexp.MustCompile(`(?i)Edge/([0-9.]+)|Edg/([0-9.]+)`),
		operaPattern:   regexp.MustCompile(`(?i)Opera/([0-9.]+)|OPR/([0-9.]+)`),
		iePattern:      regexp.MustCompile(`(?i)MSIE ([0-9.]+)|Trident.*rv:([0-9.]+)`),
	}
}

// Parse extracts device, OS, and browser information from user agent string
func (p *userAgentParser) Parse(userAgent string) (device, os, browser string) {
	if userAgent == "" {
		return "Unknown", "Unknown", "Unknown"
	}

	userAgent = strings.TrimSpace(userAgent)

	device = p.parseDevice(userAgent)
	os = p.parseOS(userAgent)
	browser = p.parseBrowser(userAgent)

	return device, os, browser
}

// parseDevice determines the device type from user agent
func (p *userAgentParser) parseDevice(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	// Special handling for Android devices
	if strings.Contains(userAgent, "android") {
		if !strings.Contains(userAgent, "mobile") {
			return "Tablet"
		}
		return "Mobile"
	}

	// Check for tablet (iPad and other tablets)
	if p.tabletPattern.MatchString(userAgent) {
		return "Tablet"
	}

	// Check for mobile
	if p.mobilePattern.MatchString(userAgent) {
		return "Mobile"
	}

	// Check for desktop indicators
	if p.desktopPattern.MatchString(userAgent) {
		return "Desktop"
	}

	return "Desktop"
}

// parseOS determines the operating system from user agent
func (p *userAgentParser) parseOS(userAgent string) string {
	// Check for iOS first
	if p.iosPattern.MatchString(userAgent) {
		if strings.Contains(strings.ToLower(userAgent), "iphone") {
			return "iOS (iPhone)"
		} else if strings.Contains(strings.ToLower(userAgent), "ipad") {
			return "iOS (iPad)"
		} else if strings.Contains(strings.ToLower(userAgent), "ipod") {
			return "iOS (iPod)"
		}
		return "iOS"
	}

	// Check for Android
	if p.androidPattern.MatchString(userAgent) {
		androidVersionPattern := regexp.MustCompile(`(?i)Android ([0-9.]+)`)
		if matches := androidVersionPattern.FindStringSubmatch(userAgent); len(matches) > 1 {
			return "Android " + matches[1]
		}
		return "Android"
	}

	// Check for Windows
	if p.windowsPattern.MatchString(userAgent) {
		if strings.Contains(userAgent, "Windows NT 10.0") {
			return "Windows 10"
		} else if strings.Contains(userAgent, "Windows NT 6.3") {
			return "Windows 8.1"
		} else if strings.Contains(userAgent, "Windows NT 6.2") {
			return "Windows 8"
		} else if strings.Contains(userAgent, "Windows NT 6.1") {
			return "Windows 7"
		} else if strings.Contains(userAgent, "Windows NT 6.0") {
			return "Windows Vista"
		} else if strings.Contains(userAgent, "Windows NT 5.1") {
			return "Windows XP"
		}
		return "Windows"
	}

	if p.macPattern.MatchString(userAgent) {
		macVersionPattern := regexp.MustCompile(`(?i)Mac OS X ([0-9_]+)`)
		if matches := macVersionPattern.FindStringSubmatch(userAgent); len(matches) > 1 {
			version := strings.ReplaceAll(matches[1], "_", ".")
			return "macOS " + version
		}
		return "macOS"
	}

	if p.linuxPattern.MatchString(userAgent) {
		return "Linux"
	}

	return "Unknown"
}

// parseBrowser determines the browser from user agent
func (p *userAgentParser) parseBrowser(userAgent string) string {
	// Check for Edge first (as it also contains Chrome)
	if p.edgePattern.MatchString(userAgent) {
		if matches := p.edgePattern.FindStringSubmatch(userAgent); len(matches) > 1 {
			version := matches[1]
			if version == "" && len(matches) > 2 {
				version = matches[2]
			}
			return "Edge " + version
		}
		return "Edge"
	}

	// Check for Opera (as it may also contain Chrome)
	if p.operaPattern.MatchString(userAgent) {
		if matches := p.operaPattern.FindStringSubmatch(userAgent); len(matches) > 1 {
			version := matches[1]
			if version == "" && len(matches) > 2 {
				version = matches[2]
			}
			return "Opera " + version
		}
		return "Opera"
	}

	// Check for Chrome (before Safari, as Chrome contains Safari)
	if p.chromePattern.MatchString(userAgent) && !strings.Contains(userAgent, "Edge") {
		if matches := p.chromePattern.FindStringSubmatch(userAgent); len(matches) > 1 {
			return "Chrome " + matches[1]
		}
		return "Chrome"
	}

	// Check for Firefox
	if p.firefoxPattern.MatchString(userAgent) {
		if matches := p.firefoxPattern.FindStringSubmatch(userAgent); len(matches) > 1 {
			return "Firefox " + matches[1]
		}
		return "Firefox"
	}

	// Check for Safari (after Chrome check)
	if p.safariPattern.MatchString(userAgent) && !strings.Contains(userAgent, "Chrome") {
		if matches := p.safariPattern.FindStringSubmatch(userAgent); len(matches) > 1 {
			return "Safari " + matches[1]
		}
		return "Safari"
	}

	// Check for Internet Explorer
	if p.iePattern.MatchString(userAgent) {
		if matches := p.iePattern.FindStringSubmatch(userAgent); len(matches) > 1 {
			version := matches[1]
			if version == "" && len(matches) > 2 {
				version = matches[2]
			}
			return "Internet Explorer " + version
		}
		return "Internet Explorer"
	}

	return "Unknown"
}