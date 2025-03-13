package integration_test

import (
	"fmt"
	"testing"

	"github.com/IrineSistiana/mosdns/v5/coremain"
	"github.com/IrineSistiana/mosdns/v5/mlog"
	_ "github.com/IrineSistiana/mosdns/v5/plugin" // Import all plugins to ensure they're registered
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

// TestDNSFilter tests DNS filtering functionality, including domain matching with blackhole/redirect combinations
func TestDNSFilter(t *testing.T) {
	// Set up a configuration with filtering capabilities
	testPort := 15357
	cfg := &coremain.Config{
		Log: mlog.LogConfig{ // Changed from pointer to value
			Level: "error", // Set to error level to avoid excessive logging
		},
		Plugins: []coremain.PluginConfig{
			// Define domain sets
			{
				Tag:  "blocked_domains",
				Type: "domain_set",
				Args: map[string]interface{}{
					"exps": []string{
						"blocked.example.com",
						"*.malicious.example.org",
					},
				},
			},
			// Define redirect domain sets
			{
				Tag:  "redirected_domains",
				Type: "domain_set",
				Args: map[string]interface{}{
					"exps": []string{
						"redirect.example.com",
						"*.redirect.example.org",
					},
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
			// Define blackhole sequence
			{
				Tag:  "black_hole",
				Type: "sequence",
				Args: []map[string]interface{}{
					{
						"exec": "black_hole 0.0.0.0",
					},
				},
			},
			// Define redirect handler
			{
				Tag:  "redirect_handler",
				Type: "redirect",
				Args: map[string]interface{}{
					"rules": []string{
						"redirect.example.com 192.168.1.100",
						"*.redirect.example.org 192.168.1.100",
					},
				},
			},
			// Define blackhole matcher
			{
				Tag:  "qname_matcher_block",
				Type: "sequence",
				Args: []map[string]interface{}{
					{
						"matches": []string{"qname $blocked_domains"},
						"exec":    "$black_hole",
					},
				},
			},
			// Define redirect matcher
			{
				Tag:  "qname_matcher_redirect",
				Type: "sequence",
				Args: []map[string]interface{}{
					{
						"matches": []string{"qname $redirected_domains"},
						"exec":    "$redirect_handler",
					},
				},
			},
			// Define main processing sequence
			{
				Tag:  "main_sequence",
				Type: "sequence",
				Args: []map[string]interface{}{
					{
						"exec": "$qname_matcher_block",
					},
					{
						"exec": "$qname_matcher_redirect",
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

	// Test blackhole functionality
	t.Run("BlackHole", func(t *testing.T) {
		// Create a DNS query for a blocked domain
		m := new(dns.Msg)
		m.SetQuestion("blocked.example.com.", dns.TypeA)

		// Send query
		client := dns.Client{Net: "udp"}
		resp, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Print response for debugging
		t.Logf("Response: %+v", resp)
		t.Logf("Response code: %d", resp.Rcode)
		t.Logf("Answer section: %+v", resp.Answer)

		// Expect NXDOMAIN response
		require.Equal(t, dns.RcodeNameError, resp.Rcode, "Expected NXDOMAIN response")
	})

	// Test redirect functionality
	t.Run("Redirect", func(t *testing.T) {
		// Create a DNS query for a redirected domain
		m := new(dns.Msg)
		m.SetQuestion("redirect.example.com.", dns.TypeA)

		// Send query
		client := dns.Client{Net: "udp"}
		resp, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Print response for debugging
		t.Logf("Response: %+v", resp)
		t.Logf("Response code: %d", resp.Rcode)
		t.Logf("Answer section: %+v", resp.Answer)

		// Expect NXDOMAIN response
		require.Equal(t, dns.RcodeNameError, resp.Rcode, "Expected NXDOMAIN response")
	})

	// Test normal forwarding functionality
	t.Run("Forward", func(t *testing.T) {
		// Create a normal DNS query
		m := new(dns.Msg)
		m.SetQuestion("example.com.", dns.TypeA)

		// Send query
		client := dns.Client{Net: "udp"}
		resp, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, dns.RcodeSuccess, resp.Rcode)
		require.True(t, len(resp.Answer) > 0, "Should have answers")
	})

	// Wait for server to close gracefully
	sc.SendCloseSignal(nil)
	err = sc.WaitClosed()
	require.NoError(t, err)
}
