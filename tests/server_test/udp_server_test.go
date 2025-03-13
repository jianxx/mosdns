package server_test

import (
	"fmt"
	"testing"

	"github.com/IrineSistiana/mosdns/v5/coremain"
	"github.com/IrineSistiana/mosdns/v5/mlog"
	_ "github.com/IrineSistiana/mosdns/v5/plugin" // Import all plugins to ensure they're registered
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

// TestUDPServer tests the basic functionality of the UDP server
func TestUDPServer(t *testing.T) {
	// Set up a simple configuration for testing
	testPort := 15353
	cfg := &coremain.Config{
		Log: mlog.LogConfig{
			Level: "error", // Set to error level to avoid excessive logging
		},
		Plugins: []coremain.PluginConfig{
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
			{
				Tag:  "udp_server",
				Type: "udp_server",
				Args: map[string]interface{}{
					"entry":  "forward_handler",
					"listen": fmt.Sprintf("127.0.0.1:%d", testPort),
				},
			},
		},
	}

	// Start the test server
	server, err := coremain.NewMosdns(cfg)
	require.NoError(t, err)

	// Get the SafeClose handler for shutdown
	sc := server.GetSafeClose()

	// Set up a defer to close the server when the test is done
	defer server.CloseWithErr(nil)

	// Create a DNS query
	m := new(dns.Msg)
	m.SetQuestion("baidu.com.", dns.TypeA)

	// Send the query via UDP
	client := dns.Client{Net: "udp"}
	resp, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))

	// Verify the results
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, dns.RcodeSuccess, resp.Rcode)
	require.True(t, resp.Response)

	// Wait for server to close gracefully
	sc.SendCloseSignal(nil)
	err = sc.WaitClosed()
	require.NoError(t, err)
}

// TestUDPServerConcurrency tests the concurrent processing capability of the UDP server
func TestUDPServerConcurrency(t *testing.T) {
	// Set up a simple configuration for testing
	testPort := 15354
	cfg := &coremain.Config{
		Log: mlog.LogConfig{
			Level: "error", // Set to error level to avoid excessive logging
		},
		Plugins: []coremain.PluginConfig{
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
			{
				Tag:  "udp_server",
				Type: "udp_server",
				Args: map[string]interface{}{
					"entry":  "forward_handler",
					"listen": fmt.Sprintf("127.0.0.1:%d", testPort),
				},
			},
		},
	}

	// Start the test server
	server, err := coremain.NewMosdns(cfg)
	require.NoError(t, err)

	// Get the SafeClose handler for shutdown
	sc := server.GetSafeClose()

	// Set up a defer to close the server when the test is done
	defer server.CloseWithErr(nil)

	// Perform multiple concurrent DNS queries
	concurrency := 10
	domains := []string{
		"example.com.",
		"google.com.",
		"github.com.",
		"microsoft.com.",
		"apple.com.",
	}

	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			domain := domains[idx%len(domains)]

			// Create a DNS query
			m := new(dns.Msg)
			m.SetQuestion(domain, dns.TypeA)

			// Send the query via UDP
			client := dns.Client{Net: "udp"}
			resp, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))

			// Verify the results
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, dns.RcodeSuccess, resp.Rcode)
			require.True(t, resp.Response)

			done <- true
		}(i)
	}

	// Wait for all queries to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// Wait for server to close gracefully
	sc.SendCloseSignal(nil)
	err = sc.WaitClosed()
	require.NoError(t, err)
}
