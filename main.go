package header_converter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config the plugin configuration.
type Config struct {
	LookupServiceURL string `json:"lookupServiceUrl,omitempty"`
	DefaultTTL       int    `json:"defaultTtl,omitempty"`
	DomainIDHeader   string `json:"domainIdHeader,omitempty"`
	URLPath          string `json:"urlPath,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		LookupServiceURL: "http://domain",
		DefaultTTL:       60,
		DomainIDHeader:   "x-domain-id",
	}
}

// CacheEntry represents a cached domain lookup result
type CacheEntry struct {
	Value      string
	ExpiresAt  time.Time
	IsRedirect bool
}

// DomainLookupFilter plugin struct
type DomainLookupFilter struct {
	config *Config
	name   string
	next   http.Handler
	cache  map[string]*CacheEntry
	mutex  sync.RWMutex
}

// New created a new plugin instance.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.LookupServiceURL == "" {
		return nil, fmt.Errorf("lookupServiceUrl is required")
	}

	if config.DefaultTTL <= 0 {
		config.DefaultTTL = 60
	}

	return &DomainLookupFilter{
		config: config,
		name:   name,
		next:   next,
		cache:  make(map[string]*CacheEntry),
		mutex:  sync.RWMutex{},
	}, nil
}

func (a *DomainLookupFilter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	hostname := req.Host
	if hostname == "" {
		a.next.ServeHTTP(rw, req)
		return
	}

	// Check cache first
	if entry := a.getCacheEntry(hostname); entry != nil {
		if time.Now().Before(entry.ExpiresAt) {
			if entry.Value == "NOT FOUND" {
				http.Error(rw, "Unauthorized (404)", http.StatusNotFound)
				return
			}

			if entry.IsRedirect {
				http.Redirect(rw, req, entry.Value, http.StatusFound)
				return
			}

			// Parse cached domain info and validate client IP
			clientIP := a.getClientIP(req)
			uuid, allowedIPs := a.parseDomainInfo(entry.Value)

			if !a.isIPAllowed(clientIP, allowedIPs) {
				http.Error(rw, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Set UUID as x-domain-id header
			req.Header.Set(a.config.DomainIDHeader, uuid)
			a.next.ServeHTTP(rw, req)
			return
		} else {
			// Cache expired, remove entry
			a.removeCacheEntry(hostname)
		}
	}

	// Cache miss, make HTTP call to lookup service
	domainInfo, statusCode, cacheControl, err := a.lookupDomain(hostname)
	if err != nil {
		// On error, continue without blocking
		a.next.ServeHTTP(rw, req)
		return
	}

	switch statusCode {
	case 200:
		// Parse domain info and validate client IP
		clientIP := a.getClientIP(req)
		uuid, allowedIPs := a.parseDomainInfo(domainInfo)

		if !a.isIPAllowed(clientIP, allowedIPs) {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Cache the result
		maxAge := a.parseMaxAge(cacheControl)
		if maxAge > 0 {
			a.setCacheEntry(hostname, domainInfo, time.Now().Add(time.Duration(maxAge)*time.Second), false)
		}

		// Set UUID as x-agency-id header
		req.Header.Set("x-agency-id", uuid)
		a.next.ServeHTTP(rw, req)

	case 201:
		// Redirect case
		maxAge := a.parseMaxAge(cacheControl)
		if maxAge > 0 {
			a.setCacheEntry(hostname, domainInfo, time.Now().Add(time.Duration(maxAge)*time.Second), true)
		}
		http.Redirect(rw, req, domainInfo, http.StatusFound)

	case 404:
		// Cache NOT FOUND result
		a.setCacheEntry(hostname, "NOT FOUND", time.Now().Add(time.Duration(a.config.DefaultTTL)*time.Second), false)
		http.Error(rw, "Page not found", http.StatusUnauthorized)

	default:
		http.Error(rw, "Unexpected error occurred", http.StatusInternalServerError)
	}
}

// getClientIP extracts the client IP from X-Forwarded-For header or RemoteAddr
func (a *DomainLookupFilter) getClientIP(req *http.Request) string {
	xForwardedFor := req.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// Take the first IP from the comma-separated list
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Fallback to RemoteAddr
	remoteAddr := req.RemoteAddr
	if colonIndex := strings.LastIndex(remoteAddr, ":"); colonIndex != -1 {
		return remoteAddr[:colonIndex]
	}
	return remoteAddr
}

// parseDomainInfo parses the domain info string format: "uuid|ip1,ip2,ip3"
func (a *DomainLookupFilter) parseDomainInfo(domainInfo string) (string, []string) {
	parts := strings.Split(domainInfo, "|")
	if len(parts) < 2 {
		return "", []string{}
	}

	uuid := parts[0]
	allowedIPsString := parts[1]

	if allowedIPsString == "" {
		return uuid, []string{}
	}

	allowedIPs := strings.Split(allowedIPsString, ",")
	for i, ip := range allowedIPs {
		allowedIPs[i] = strings.TrimSpace(ip)
	}

	return uuid, allowedIPs
}

// isIPAllowed checks if the client IP is in the allowed IPs list
func (a *DomainLookupFilter) isIPAllowed(clientIP string, allowedIPs []string) bool {
	if len(allowedIPs) == 0 {
		return true // No restrictions
	}

	for _, allowedIP := range allowedIPs {
		if strings.TrimSpace(allowedIP) == clientIP {
			return true
		}
	}
	return false
}

// lookupDomain makes an HTTP call to the admin lookup service
func (a *DomainLookupFilter) lookupDomain(hostname string) (string, int, string, error) {
	// lookupURL := fmt.Sprintf("%s/api/admin-domain/domain-to-agency-id?domain=%s", a.config.LookupServiceURL, hostname)
	lookupURL := fmt.Sprintf("%s%s?domain=%s", a.config.LookupServiceURL, a.config.URLPath, hostname)

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	resp, err := client.Get(lookupURL)
	if err != nil {
		return "", 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, "", err
	}

	cacheControl := resp.Header.Get("Cache-Control")
	return string(body), resp.StatusCode, cacheControl, nil
}

// parseMaxAge extracts max-age value from Cache-Control header
func (a *DomainLookupFilter) parseMaxAge(cacheControl string) int {
	if cacheControl == "" {
		return a.config.DefaultTTL
	}

	directives := strings.Split(cacheControl, ",")
	for _, directive := range directives {
		directive = strings.TrimSpace(directive)
		if strings.HasPrefix(directive, "max-age=") {
			maxAgeStr := directive[8:]
			if maxAge, err := strconv.Atoi(maxAgeStr); err == nil {
				return maxAge
			}
		}
	}
	return a.config.DefaultTTL
}

// Cache management methods
func (a *DomainLookupFilter) getCacheEntry(key string) *CacheEntry {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.cache[key]
}

func (a *DomainLookupFilter) setCacheEntry(key, value string, expiresAt time.Time, isRedirect bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.cache[key] = &CacheEntry{
		Value:      value,
		ExpiresAt:  expiresAt,
		IsRedirect: isRedirect,
	}
}

func (a *DomainLookupFilter) removeCacheEntry(key string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	delete(a.cache, key)
}
