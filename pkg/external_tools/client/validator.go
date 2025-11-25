package client

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	interfaceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
	hostnamePattern      = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*[a-zA-Z0-9]$`)
)

func ValidateInterface(iface string) error {
	if iface == "" {
		return nil
	}
	if iface == "any" {
		return nil
	}
	if len(iface) > 15 {
		return fmt.Errorf("interface name too long: %s", iface)
	}
	if !interfaceNamePattern.MatchString(iface) {
		return fmt.Errorf("invalid interface name: %s", iface)
	}
	return nil
}

func ValidateIP(ipStr string) error {
	if ipStr == "" {
		return nil
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipStr)
	}
	return nil
}

func ValidateHostname(host string) error {
	if host == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	if net.ParseIP(host) != nil {
		return nil
	}

	if len(host) > 253 {
		return fmt.Errorf("hostname too long: %s", host)
	}
	if !hostnamePattern.MatchString(host) {
		return fmt.Errorf("invalid hostname: %s", host)
	}
	return nil
}

func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be 1-65535)", port)
	}
	return nil
}

func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must have a scheme (http:// or https://)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}
	return nil
}

func ValidateTableName(table string) error {
	if table == "" {
		return nil
	}

	if _, err := strconv.Atoi(table); err == nil {
		return nil
	}
	validTables := map[string]bool{
		"filter":  true,
		"nat":     true,
		"mangle":  true,
		"raw":     true,
		"main":    true,
		"local":   true,
		"default": true,
	}

	if !validTables[table] {
		return fmt.Errorf("invalid table name: %s", table)
	}
	return nil
}

func ValidateBPFFilter(filter string) error {
	if filter == "" {
		return nil
	}
	if len(filter) > 1024 {
		return fmt.Errorf("BPF filter too long (max 1024 characters)")
	}

	dangerous := []string{";", "|", "&", "`", "$", "$("}
	for _, pattern := range dangerous {
		if strings.Contains(filter, pattern) {
			return fmt.Errorf("BPF filter contains potentially dangerous characters")
		}
	}
	return nil
}

func ValidateIntRange(value, min, max int, name string) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d, got %d", name, min, max, value)
	}
	return nil
}

func ValidateSysctlPattern(pattern string) error {
	if pattern == "" {
		return nil
	}

	if !strings.HasPrefix(pattern, "net.") && pattern != "net" {
		return fmt.Errorf("sysctl pattern must start with 'net.'")
	}

	dangerous := []string{";", "|", "&", "`", "$", "(", ")", "<", ">"}
	for _, char := range dangerous {
		if strings.Contains(pattern, char) {
			return fmt.Errorf("sysctl pattern contains invalid characters")
		}
	}
	return nil
}
