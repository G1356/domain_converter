package header_converter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	config := CreateConfig()
	config.LookupServiceURL = "http://test-domain"
	config.DefaultTTL = 120

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(context.Background(), next, config, "test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	plugin, ok := handler.(*DomainConverter)
	if !ok {
		t.Fatal("Expected DomainLookupFilter type")
	}

	if plugin.config.LookupServiceURL != "http://test-domain" {
		t.Errorf("Expected lookupServiceUrl to be 'http://test-domain', got %s", plugin.config.LookupServiceURL)
	}

	if plugin.config.DefaultTTL != 120 {
		t.Errorf("Expected defaultTtl to be 120, got %d", plugin.config.DefaultTTL)
	}
}

func TestNewWithEmptylookupServiceUrl(t *testing.T) {
	config := CreateConfig()
	config.LookupServiceURL = ""

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	_, err := New(context.Background(), next, config, "test")
	if err == nil {
		t.Fatal("Expected error for empty lookupServiceUrl")
	}
}

func TestGetClientIP(t *testing.T) {
	plugin := &DomainConverter{}

	tests := []struct {
		name          string
		xForwardedFor string
		remoteAddr    string
		expectedIP    string
	}{
		{
			name:          "X-Forwarded-For with single IP",
			xForwardedFor: "192.168.1.1",
			remoteAddr:    "10.0.0.1:12345",
			expectedIP:    "192.168.1.1",
		},
		{
			name:          "X-Forwarded-For with multiple IPs",
			xForwardedFor: "192.168.1.1, 10.0.0.1, 172.16.0.1",
			remoteAddr:    "10.0.0.1:12345",
			expectedIP:    "192.168.1.1",
		},
		{
			name:          "No X-Forwarded-For, use RemoteAddr",
			xForwardedFor: "",
			remoteAddr:    "192.168.1.1:12345",
			expectedIP:    "192.168.1.1",
		},
		{
			name:          "No X-Forwarded-For, RemoteAddr without port",
			xForwardedFor: "",
			remoteAddr:    "192.168.1.1",
			expectedIP:    "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			req.RemoteAddr = tt.remoteAddr

			clientIP := plugin.getClientIP(req)
			if clientIP != tt.expectedIP {
				t.Errorf("Expected client IP %s, got %s", tt.expectedIP, clientIP)
			}
		})
	}
}

func TestParseDomainInfo(t *testing.T) {
	plugin := &DomainConverter{}

	tests := []struct {
		name         string
		domainInfo   string
		expectedUUID string
		expectedIPs  []string
	}{
		{
			name:         "Valid domain info with IPs",
			domainInfo:   "uuid123|192.168.1.1,10.0.0.1,172.16.0.1",
			expectedUUID: "uuid123",
			expectedIPs:  []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
		},
		{
			name:         "Valid domain info without IPs",
			domainInfo:   "uuid456|",
			expectedUUID: "uuid456",
			expectedIPs:  []string{},
		},
		{
			name:         "Invalid format",
			domainInfo:   "uuid789",
			expectedUUID: "",
			expectedIPs:  []string{},
		},
		{
			name:         "Empty domain info",
			domainInfo:   "",
			expectedUUID: "",
			expectedIPs:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid, ips := plugin.parseDomainInfo(tt.domainInfo)
			if uuid != tt.expectedUUID {
				t.Errorf("Expected UUID %s, got %s", tt.expectedUUID, uuid)
			}

			if len(ips) != len(tt.expectedIPs) {
				t.Errorf("Expected %d IPs, got %d", len(tt.expectedIPs), len(ips))
				return
			}

			for i, ip := range ips {
				if ip != tt.expectedIPs[i] {
					t.Errorf("Expected IP %s at index %d, got %s", tt.expectedIPs[i], i, ip)
				}
			}
		})
	}
}

func TestIsIPAllowed(t *testing.T) {
	plugin := &DomainConverter{}

	tests := []struct {
		name       string
		clientIP   string
		allowedIPs []string
		expected   bool
	}{
		{
			name:       "IP in allowed list",
			clientIP:   "192.168.1.1",
			allowedIPs: []string{"192.168.1.1", "10.0.0.1"},
			expected:   true,
		},
		{
			name:       "IP not in allowed list",
			clientIP:   "172.16.0.1",
			allowedIPs: []string{"192.168.1.1", "10.0.0.1"},
			expected:   false,
		},
		{
			name:       "Empty allowed list (no restrictions)",
			clientIP:   "192.168.1.1",
			allowedIPs: []string{},
			expected:   true,
		},
		{
			name:       "IP with whitespace in allowed list",
			clientIP:   "192.168.1.1",
			allowedIPs: []string{" 192.168.1.1 ", "10.0.0.1"},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := plugin.isIPAllowed(tt.clientIP, tt.allowedIPs)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParseMaxAge(t *testing.T) {
	plugin := &DomainConverter{
		config: &Config{DefaultTTL: 60},
	}

	tests := []struct {
		name         string
		cacheControl string
		expected     int
	}{
		{
			name:         "Valid max-age",
			cacheControl: "max-age=300",
			expected:     300,
		},
		{
			name:         "Max-age with other directives",
			cacheControl: "public, max-age=600, must-revalidate",
			expected:     600,
		},
		{
			name:         "No max-age directive",
			cacheControl: "public, must-revalidate",
			expected:     60, // default TTL
		},
		{
			name:         "Empty cache control",
			cacheControl: "",
			expected:     60, // default TTL
		},
		{
			name:         "Invalid max-age value",
			cacheControl: "max-age=invalid",
			expected:     60, // default TTL
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := plugin.parseMaxAge(tt.cacheControl)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCacheOperations(t *testing.T) {
	plugin := &DomainConverter{
		cache: make(map[string]*CacheEntry),
	}

	// Test setting cache entry
	key := "example.com"
	value := "uuid123|192.168.1.1"
	expiresAt := time.Now().Add(5 * time.Minute)
	plugin.setCacheEntry(key, value, expiresAt, false)

	// Test getting cache entry
	entry := plugin.getCacheEntry(key)
	if entry == nil {
		t.Fatal("Expected cache entry to exist")
	}

	if entry.Value != value {
		t.Errorf("Expected value %s, got %s", value, entry.Value)
	}

	if entry.IsRedirect != false {
		t.Errorf("Expected IsRedirect to be false, got %v", entry.IsRedirect)
	}

	// Test removing cache entry
	plugin.removeCacheEntry(key)
	entry = plugin.getCacheEntry(key)
	if entry != nil {
		t.Error("Expected cache entry to be removed")
	}
}
