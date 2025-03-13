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

// TestTCPServer tests the basic functionality of the TCP server
func TestTCPServer(t *testing.T) {
	// Set up a simple configuration for testing
	testPort := 15355
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
				Tag:  "tcp_server",
				Type: "tcp_server",
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

	// Send the query via TCP
	client := dns.Client{Net: "tcp"}
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

// TestTCPServerLargeResponse tests the TCP server's ability to handle large responses
func TestTCPServerLargeResponse(t *testing.T) {
	// Set up a simple configuration for testing
	testPort := 15356
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
				Tag:  "tcp_server",
				Type: "tcp_server",
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

	// Create a DNS query that may generate a large number of records
	m := new(dns.Msg)
	m.SetQuestion("baidu.com.", dns.TypeANY) // Request all records with ANY type

	// Send the query via TCP
	client := dns.Client{Net: "tcp"}
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
