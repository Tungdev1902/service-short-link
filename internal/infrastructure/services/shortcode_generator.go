package services

import (
    "crypto/rand"
    "fmt"
    "math/big"
    "regexp"
    "strings"
    "time"

    "service-short-link/internal/domain"
    "service-short-link/pkg/logger"
)

const (
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    defaultLength = 7
    defaultMinLength = 5
    defaultMaxLength = 10
	maxRetries = 5
)

type shortCodeGenerator struct {
	charset  string
	length   int
    pattern  *regexp.Regexp
    minLength int
    maxLength int
}

// NewShortCodeGenerator creates a new short code generator
func NewShortCodeGenerator() domain.ShortCodeGenerator {
    return NewShortCodeGeneratorWithBounds(defaultMinLength, defaultMaxLength)
}

// NewShortCodeGeneratorWithBounds creates a generator with configurable min/max
func NewShortCodeGeneratorWithBounds(minLen, maxLen int) domain.ShortCodeGenerator {
    if minLen < 1 {
        minLen = defaultMinLength
    }
    if maxLen < minLen {
        maxLen = minLen
    }
    pattern := regexp.MustCompile(fmt.Sprintf(`^[a-zA-Z0-9]{%d,%d}$`, minLen, maxLen))
    return &shortCodeGenerator{
        charset:   base62Chars,
        length:    defaultLength,
        pattern:   pattern,
        minLength: minLen,
        maxLength: maxLen,
    }
}

// Generate creates a new short code with default length
func (g *shortCodeGenerator) Generate() string {
	return g.GenerateWithLength(g.length)
}

// GenerateWithLength creates a new short code with specified length
func (g *shortCodeGenerator) GenerateWithLength(length int) string {
    if length < g.minLength {
        length = g.minLength
    }
    if length > g.maxLength {
        length = g.maxLength
    }

	timestamp := time.Now().UnixNano()
	timestampCode := g.encodeBase62(timestamp)
	randomSuffix := g.generateRandomString(length - len(timestampCode))
	code := timestampCode + randomSuffix
	
	if len(code) > length {
		code = code[:length]
	} else if len(code) < length {
		code += g.generateRandomString(length - len(code))
	}

	return code
}

// IsValid checks if a short code is valid
func (g *shortCodeGenerator) IsValid(shortCode string) bool {
	if shortCode == "" {
		return false
	}
	
	if !g.pattern.MatchString(shortCode) {
		return false
	}
	
	reservedWords := []string{
		"admin", "api", "swagger", "health", "ready", "live",
	}
	
	lowerCode := strings.ToLower(shortCode)
	for _, reserved := range reservedWords {
		if lowerCode == reserved {
			return false
		}
	}
	
	return true
}

// encodeBase62 converts a number to base62 string
func (g *shortCodeGenerator) encodeBase62(num int64) string {
	if num == 0 {
		return string(g.charset[0])
	}

	result := ""
	base := int64(len(g.charset))
	
	for num > 0 {
		remainder := num % base
		result = string(g.charset[remainder]) + result
		num = num / base
	}
	
	return result
}

// generateRandomString creates a random string of specified length
func (g *shortCodeGenerator) generateRandomString(length int) string {
	if length <= 0 {
		return ""
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(g.charset)))
	
	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			result[i] = g.charset[time.Now().UnixNano()%int64(len(g.charset))]
		} else {
			result[i] = g.charset[randomIndex.Int64()]
		}
	}
	
	return string(result)
}

func GenerateUniqueCode(generator domain.ShortCodeGenerator, repo domain.LinkRepository, length int) (string, error) {
    currentLength := length
    for i := 0; i < maxRetries; i++ {
        code := generator.GenerateWithLength(currentLength)

        exists, err := repo.ExistsByShortCode(code)
        if err != nil {
            logger.ErrorWithCockroachSimple(err, "ShortCodeGenerator.GenerateUnique: failed to check code existence", "code="+code, "attempt="+fmt.Sprintf("%d", i+1), "error_type=database_query_failed")
            return "", fmt.Errorf("failed to check code existence: %w", err)
        }

        if !exists {
            return code, nil
        }

        if currentLength < 10 {
            currentLength++
        }

        time.Sleep(time.Millisecond * time.Duration(i+1))
    }

    return "", fmt.Errorf("failed to generate unique code after %d attempts", maxRetries)
}