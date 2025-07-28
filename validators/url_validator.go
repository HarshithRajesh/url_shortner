package validators

import (
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"
)

func ValidateURL(inputURL string) error {
	// Basic validation
	if len(inputURL) == 0 {
		return errors.New("URL cannot be empty")
	}

	if len(inputURL) > 2048 {
		return errors.New("URL too long (max 2048 characters)")
	}

	// Parse URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return errors.New("invalid URL format")
	}

	// Scheme validation
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("URL must have a scheme and host")
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return errors.New("only http and https schemes are allowed")
	}

	// Hostname validation
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return errors.New("URL must have a valid hostname")
	}

	// Block suspicious patterns BEFORE DNS lookup
	if containsSuspiciousPatterns(inputURL) {
		return errors.New("URL contains suspicious patterns")
	}

	// Comprehensive SSRF protection
	if isPrivateOrLocalhost(hostname) {
		return errors.New("private/internal URLs are not allowed")
	}

	return nil
}

// ValidateCustomCode validates short URL codes
func ValidateCustomCode(code string) error {
	if len(code) == 0 {
		return nil // Empty is OK, will generate random
	}

	// Length validation
	if len(code) < 3 || len(code) > 50 {
		return errors.New("custom code must be between 3 and 50 characters")
	}

	// Character validation - only alphanumeric, dash, underscore
	validCode := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	if !validCode.MatchString(code) {
		return errors.New("custom code can only contain letters, numbers, hyphens, and underscores")
	}

	// Block reserved words
	reservedWords := []string{
		"api", "admin", "www", "ftp", "mail", "health", "status",
		"docs", "help", "support", "about", "contact", "login",
		"register", "dashboard", "settings", "profile", "user",
	}

	codeToCheck := strings.ToLower(code)
	for _, word := range reservedWords {
		if codeToCheck == word {
			return errors.New("custom code uses a reserved word")
		}
	}

	return nil
}

// Enhanced SSRF protection with more comprehensive checks
func isPrivateOrLocalhost(hostname string) bool {
	hostname = strings.ToLower(hostname)

	// Direct localhost variations (including bypass attempts)
	localhostVariants := []string{
		"localhost", "127.0.0.1", "::1", "0.0.0.0",
		"127.1", "127.0.1", "0x7f000001", "0177.0.0.1",
		"127.0.0.0", "127.255.255.255", "127.1.1.1",
		"0x7f.0x0.0x0.0x1", "0177.0.0.1",
	}

	for _, variant := range localhostVariants {
		if hostname == variant {
			return true
		}
	}

	// Check for 127.x.x.x patterns
	if strings.HasPrefix(hostname, "127.") {
		return true
	}

	// DNS resolution with timeout and private IP check
	ips, err := net.LookupIP(hostname)
	if err != nil {
		// If DNS fails, allow (will fail later anyway)
		return false
	}

	// Check all resolved IPs
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return true
		}
	}

	return false
}

// More comprehensive private IP detection
func isPrivateIP(ip net.IP) bool {
	// Standard private ranges
	privateRanges := []string{
		"10.0.0.0/8",     // Class A private
		"172.16.0.0/12",  // Class B private
		"192.168.0.0/16", // Class C private
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link-local
		"224.0.0.0/4",    // Multicast
		"240.0.0.0/4",    // Reserved
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 private
		"fe80::/10",      // IPv6 link-local
		"ff00::/8",       // IPv6 multicast
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	// Additional checks for edge cases
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsMulticast() {
		return true
	}

	return false
}

// Detect suspicious URL patterns that could be used for attacks
func containsSuspiciousPatterns(inputURL string) bool {
	urlLower := strings.ToLower(inputURL)

	suspiciousPatterns := []string{
		"javascript:", // XSS
		"data:",       // Data URLs
		"file:",       // File protocol
		"ftp:",        // FTP protocol
		"gopher:",     // Gopher protocol
		"ldap:",       // LDAP protocol
		"dict:",       // Dict protocol
		"@",           // URL redirect tricks like http://evil.com@good.com
		"\\",          // Windows path separators
		"%0a", "%0d",  // URL-encoded newlines
		"%00",  // Null bytes
		"../",  // Path traversal
		"..\\", // Windows path traversal
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(urlLower, pattern) {
			return true
		}
	}

	// Check for potential URL encoding bypasses
	if strings.Contains(urlLower, "%") {
		// Basic check for potentially dangerous encoded characters
		dangerousEncoded := []string{
			"%2e%2e", "%2f", "%5c", "%00", "%0a", "%0d",
		}
		for _, encoded := range dangerousEncoded {
			if strings.Contains(urlLower, encoded) {
				return true
			}
		}
	}

	return false
}

// GetValidatedURL - helper function that returns the validated URL
func GetValidatedURL(inputURL string) (string, error) {
	err := ValidateURL(inputURL)
	if err != nil {
		return "", err
	}
	return inputURL, nil
}
