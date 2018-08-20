package logger

import (
	"net"
	"net/http"
	"strings"
)

// GetRemoteAddr gets the origin/client ip for a request.
// X-FORWARDED-FOR is checked. If multiple IPs are included the first one is returned
// X-REAL-IP is checked. If multiple IPs are included the first one is returned
// Finally r.RemoteAddr is used
// Only benevolent services will allow access to the real IP.
func GetRemoteAddr(r *http.Request) string {
	if r == nil {
		return ""
	}
	tryHeader := func(key string) (string, bool) {
		if headerVal := r.Header.Get(key); len(headerVal) > 0 {
			if !strings.ContainsRune(headerVal, ',') {
				return headerVal, true
			}
			return strings.SplitN(headerVal, ",", 2)[0], true
		}
		return "", false
	}

	for _, header := range []string{"X-FORWARDED-FOR", "X-REAL-IP"} {
		if headerVal, ok := tryHeader(header); ok {
			return headerVal
		}
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// GetProto gets the request proto.
// X-FORWARDED-PROTO is checked first, then the original request proto is used.
func GetProto(r *http.Request) string {
	if r == nil {
		return ""
	}
	tryHeader := func(key string) (string, bool) {
		if headerVal := r.Header.Get(key); len(headerVal) > 0 {
			if !strings.ContainsRune(headerVal, ',') {
				return headerVal, true
			}
			return strings.SplitN(headerVal, ",", 2)[0], true
		}
		return "", false
	}

	for _, header := range []string{"X-FORWARDED-PROTO"} {
		if headerVal, ok := tryHeader(header); ok {
			return headerVal
		}
	}

	return r.Proto
}

// GetHost returns the request host, omiting the port if specified.
func GetHost(r *http.Request) string {
	if r == nil {
		return ""
	}
	tryHeader := func(key string) (string, bool) {
		if headerVal := r.Header.Get(key); len(headerVal) > 0 {
			if !strings.ContainsRune(headerVal, ',') {
				return headerVal, true
			}
			return strings.SplitN(headerVal, ",", 2)[0], true
		}
		return "", false
	}

	for _, header := range []string{"X-FORWARDED-HOST"} {
		if headerVal, ok := tryHeader(header); ok {
			return headerVal
		}
	}
	if r.URL != nil && len(r.URL.Host) > 0 {
		return r.URL.Host
	}
	if strings.Contains(r.Host, ":") {
		return strings.SplitN(r.Host, ":", 2)[0]
	}
	return r.Host
}

// GetUserAgent gets a user agent from a request.
func GetUserAgent(r *http.Request) string {
	return r.UserAgent()
}
