package services

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"

	"service-short-link/internal/domain"
	"service-short-link/pkg/logger"
)

type urlValidator struct {
	httpPattern  *regexp.Regexp
	httpsPattern *regexp.Regexp
	
	blacklistedDomains []string
	maliciousPatterns  []*regexp.Regexp
}

// NewURLValidator creates a new URL validator
func NewURLValidator() domain.URLValidator {
	return &urlValidator{
		httpPattern:  regexp.MustCompile(`^https?://`),
		httpsPattern: regexp.MustCompile(`^https://`),
		blacklistedDomains: []string{
			"localhost",
			"127.0.0.1",
			"0.0.0.0",
			"::1",
			"10.",
			"172.",
			"192.168.",
		},
		maliciousPatterns: []*regexp.Regexp{
			regexp.MustCompile(`javascript:`),
			regexp.MustCompile(`data:`),
			regexp.MustCompile(`vbscript:`),
			regexp.MustCompile(`file:`),
			regexp.MustCompile(`ftp:`),
		},
	}
}

// IsValid checks if a URL is valid and safe
func (v *urlValidator) IsValid(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	decodedURL, err := url.QueryUnescape(rawURL)
	if err != nil {
		decodedURL = rawURL
	}

	parsedURL, err := url.Parse(decodedURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	if parsedURL.Host == "" {
		return false
	}

	lowerURL := strings.ToLower(rawURL)
	lowerDecodedURL := strings.ToLower(decodedURL)
	
	for _, pattern := range v.maliciousPatterns {
		if pattern.MatchString(lowerURL) || pattern.MatchString(lowerDecodedURL) {
			return false
		}
	}

	return v.isValidHost(parsedURL.Host)
}

// Normalize cleans and standardizes a URL
func (v *urlValidator) Normalize(rawURL string) (string, error) {
	if rawURL == "" {
		return "", domain.ErrInvalidURL
	}

	decodedURL, err := url.QueryUnescape(rawURL)
	if err != nil {
		decodedURL = rawURL
	}

	if !v.httpPattern.MatchString(decodedURL) {
		decodedURL = "http://" + decodedURL
	}

	parsedURL, err := url.Parse(decodedURL)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "URLValidator.Normalize: failed to parse URL", "raw_url="+rawURL, "decoded_url="+decodedURL, "error_type=url_parse_failed")
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	parsedURL.Host = strings.ToLower(parsedURL.Host)

	if parsedURL.Scheme == "http" && strings.HasSuffix(parsedURL.Host, ":80") {
		parsedURL.Host = strings.TrimSuffix(parsedURL.Host, ":80")
	}
	if parsedURL.Scheme == "https" && strings.HasSuffix(parsedURL.Host, ":443") {
		parsedURL.Host = strings.TrimSuffix(parsedURL.Host, ":443")
	}

	if parsedURL.Path == "/" {
		parsedURL.Path = ""
	}

	normalized := parsedURL.String()
	if !v.IsValid(normalized) {
		return "", domain.ErrInvalidURL
	}

	return normalized, nil
}

// IsSafeURL checks if a URL is safe to redirect to
func (v *urlValidator) IsSafeURL(rawURL string) bool {
	if !v.IsValid(rawURL) {
		return false
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	host := strings.ToLower(parsedURL.Host)
	
	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
	}

	for _, blacklisted := range v.blacklistedDomains {
		if strings.Contains(host, blacklisted) {
			return false
		}
	}

	if v.isPrivateIP(host) {
		return false
	}

	return parsedURL.Scheme == "https" || parsedURL.Scheme == "http"
}

// isValidHost checks if a host is valid
func (v *urlValidator) isValidHost(host string) bool {
	if host == "" {
		return false
	}

	hostname := host
	if strings.Contains(host, ":") {
		var err error
		hostname, _, err = net.SplitHostPort(host)
		if err != nil {
			return false
		}
	}

	if ip := net.ParseIP(hostname); ip != nil {
		return !v.isPrivateIP(hostname)
	}

	return v.isValidDomain(hostname)
}

// isValidDomain checks if a domain name is valid
func (v *urlValidator) isValidDomain(domain string) bool {
	if domain == "" || len(domain) > 253 {
		return false
	}

	if !strings.Contains(domain, ".") {
		return false
	}

	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		
		if !regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`).MatchString(label) {
			return false
		}
	}

	return true
}

// isPrivateIP checks if an IP address is private/local
func (v *urlValidator) isPrivateIP(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12", 
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fe80::/10",
		"fc00::/7",
	}

	for _, cidr := range privateRanges {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipnet.Contains(ip) {
			return true
		}
	}

	return false
}