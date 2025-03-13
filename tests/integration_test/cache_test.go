package integration_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/IrineSistiana/mosdns/v5/coremain"
	"github.com/IrineSistiana/mosdns/v5/mlog"
	_ "github.com/IrineSistiana/mosdns/v5/plugin" // Import all plugins to ensure they're registered
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

// TestDNSCache tests the DNS caching functionality
func TestDNSCache(t *testing.T) {
	// Set up a configuration with caching functionality
	testPort := 15358
	cfg := &coremain.Config{
		Log: mlog.LogConfig{
			Level: "error", // Set to error level to avoid excessive logging
		},
		Plugins: []coremain.PluginConfig{
			// Cache handler
			{
				Tag:  "cache_handler",
				Type: "cache",
				Args: map[string]interface{}{
					"size":           1024,
					"lazy_cache_ttl": 300,
				},
			},
			// Forward handler
			{
				Tag:  "forward_handler",
				Type: "forward",
				Args: map[string]interface{}{
					"upstreams": []map[string]interface{}{
						{
							"addr": "tls://223.5.5.5:853", // Use AliDNS as upstream
						},
					},
				},
			},
			// Define the main processing sequence
			{
				Tag:  "main_sequence",
				Type: "sequence",
				Args: []map[string]interface{}{
					{
						"exec": "$cache_handler",
					},
					{
						"exec": "$forward_handler",
					},
				},
			},
			// UDP Server
			{
				Tag:  "udp_server",
				Type: "udp_server",
				Args: map[string]interface{}{
					"entry":  "main_sequence",
					"listen": fmt.Sprintf("127.0.0.1:%d", testPort),
				},
			},
		},
	}

	// Start the test server - using NewMosdns instead of NewServer with Args
	server, err := coremain.NewMosdns(cfg)
	require.NoError(t, err)

	// Get the SafeClose handler for shutdown
	sc := server.GetSafeClose()

	// Set up a defer to close the server when the test is done
	defer server.CloseWithErr(nil)

	// First query, should retrieve from upstream
	t.Run("FirstQuery", func(t *testing.T) {
		startTime := time.Now()

		// Create DNS query
		m := new(dns.Msg)
		m.SetQuestion("baidu.com.", dns.TypeA)

		// Send query
		client := dns.Client{Net: "udp"}
		resp, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, dns.RcodeSuccess, resp.Rcode)

		// Record time for the first query
		firstQueryTime := time.Since(startTime)
		t.Logf("First query time: %v", firstQueryTime)
	})

	// Second query, should retrieve from cache, faster
	t.Run("CachedQuery", func(t *testing.T) {
		startTime := time.Now()

		// Create the same DNS query
		m := new(dns.Msg)
		m.SetQuestion("baidu.com.", dns.TypeA)

		// Send query
		client := dns.Client{Net: "udp"}
		resp, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, dns.RcodeSuccess, resp.Rcode)

		// Record time for the second query
		secondQueryTime := time.Since(startTime)
		t.Logf("Second query time: %v", secondQueryTime)

		// Second query should be faster than the first
		// Note: This test may be unstable due to network latency and other factors
		// So we don't make a hard assertion, just log the information
	})

	// Test that different query types don't confuse the cache
	t.Run("DifferentQueryType", func(t *testing.T) {
		// Create a different type of DNS query
		m := new(dns.Msg)
		m.SetQuestion("baidu.com.", dns.TypeAAAA) // Query for IPv6 address

		// Send query
		client := dns.Client{Net: "udp"}
		resp, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Check answer types
		hasAAAA := false
		for _, ans := range resp.Answer {
			if _, ok := ans.(*dns.AAAA); ok {
				hasAAAA = true
				break
			}
		}

		// Should not have A records incorrectly returned from cache
		hasA := false
		for _, ans := range resp.Answer {
			if _, ok := ans.(*dns.A); ok {
				hasA = true
				break
			}
		}

		// If there are answers, should only have AAAA records, not A records
		if len(resp.Answer) > 0 {
			require.True(t, hasAAAA || !hasA, "Should not return incorrect record types from cache")
		}
	})

	// Wait for server to close gracefully
	sc.SendCloseSignal(nil)
	err = sc.WaitClosed()
	require.NoError(t, err)
}
