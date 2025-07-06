package utils

import (
	"strings"

	"gonotes/internal/model"
)

// parseUserAgent parses user agent string to extract device information
func parseUserAgent(userAgent string) *model.DeviceInfo {
	if userAgent == "" {
		return &model.DeviceInfo{
			Browser:  "Unknown",
			OS:       "Unknown",
			Device:   "Unknown",
			IsMobile: false,
		}
	}

	ua := strings.ToLower(userAgent)

	deviceInfo := &model.DeviceInfo{
		Browser:  parseBrowser(ua),
		OS:       parseOS(ua),
		Device:   parseDevice(ua),
		IsMobile: isMobileDevice(ua),
	}

	return deviceInfo
}

// parseBrowser extracts browser information from user agent
func parseBrowser(ua string) string {
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		return "Chrome"
	}
	if strings.Contains(ua, "firefox") {
		return "Firefox"
	}
	if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return "Safari"
	}
	if strings.Contains(ua, "edg") {
		return "Edge"
	}
	if strings.Contains(ua, "opera") || strings.Contains(ua, "opr") {
		return "Opera"
	}
	if strings.Contains(ua, "curl") {
		return "cURL"
	}
	if strings.Contains(ua, "postman") {
		return "Postman"
	}
	if strings.Contains(ua, "insomnia") {
		return "Insomnia"
	}
	return "Unknown"
}

// parseOS extracts operating system information from user agent
func parseOS(ua string) string {
	if strings.Contains(ua, "windows") {
		if strings.Contains(ua, "windows nt 10") {
			return "Windows 10"
		}
		if strings.Contains(ua, "windows nt 11") {
			return "Windows 11"
		}
		return "Windows"
	}
	if strings.Contains(ua, "mac os x") || strings.Contains(ua, "macos") {
		return "macOS"
	}
	if strings.Contains(ua, "linux") {
		if strings.Contains(ua, "ubuntu") {
			return "Ubuntu"
		}
		if strings.Contains(ua, "debian") {
			return "Debian"
		}
		if strings.Contains(ua, "centos") {
			return "CentOS"
		}
		return "Linux"
	}
	if strings.Contains(ua, "android") {
		return "Android"
	}
	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		return "iOS"
	}
	return "Unknown"
}

// parseDevice extracts device type information from user agent
func parseDevice(ua string) string {
	if strings.Contains(ua, "iphone") {
		return "iPhone"
	}
	if strings.Contains(ua, "ipad") {
		return "iPad"
	}
	if strings.Contains(ua, "android") {
		if strings.Contains(ua, "mobile") {
			return "Android Phone"
		}
		return "Android Tablet"
	}
	if strings.Contains(ua, "curl") {
		return "Command Line"
	}
	if strings.Contains(ua, "postman") {
		return "API Client"
	}
	if isMobileDevice(ua) {
		return "Mobile Device"
	}
	return "Desktop"
}

// isMobileDevice checks if the user agent indicates a mobile device
func isMobileDevice(ua string) bool {
	mobileIndicators := []string{
		"mobile", "android", "iphone", "ipad", "ipod",
		"blackberry", "windows phone", "palm", "symbian",
	}

	for _, indicator := range mobileIndicators {
		if strings.Contains(ua, indicator) {
			return true
		}
	}

	return false
}

// ParseUserAgent is the exported function for parsing user agent
func ParseUserAgent(userAgent string) *model.DeviceInfo {
	return parseUserAgent(userAgent)
}
