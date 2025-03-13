package integration_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/IrineSistiana/mosdns/v5/coremain"
	"github.com/IrineSistiana/mosdns/v5/mlog"
	_ "github.com/IrineSistiana/mosdns/v5/plugin" // Import all plugins to ensure they're registered
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

// TestLoadBalance tests the load balancing functionality across multiple upstream servers
func TestLoadBalance(t *testing.T) {
	// Set up a configuration with multiple upstream servers
	testPort := 15359
	cfg := &coremain.Config{
		Log: mlog.LogConfig{ // Changed from pointer to value
			Level: "error", // Set to error level to avoid excessive logging
		},
		Plugins: []coremain.PluginConfig{
			// Forward handler with multiple upstream servers
			{
				Tag:  "forward_handler",
				Type: "forward",
				Args: map[string]interface{}{
					"upstreams": []map[string]interface{}{
						{
							"addr": "tls://223.5.5.5:853", // AliDNS
							"tag":  "AliDNS",
						},
						{
							"addr": "tls://120.53.53.53:853", // DNSPod
							"tag":  "DNSPod",
						},
						{
							"addr": "9.9.9.9:53", // Quad9 DNS
							"tag":  "quad9",
						},
					},
					"concurrent": 3, // Set number of concurrent queries
				},
			},
			// Define the main processing sequence
			{
				Tag:  "main_sequence",
				Type: "sequence",
				Args: []map[string]interface{}{
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

	// Test concurrent queries
	t.Run("ConcurrentQueries", func(t *testing.T) {
		// Perform multiple concurrent DNS queries
		concurrency := 50
		domains := []string{
			"example.com.",
			"google.com.",
			"github.com.",
			"microsoft.com.",
			"apple.com.",
			"amazon.com.",
			"facebook.com.",
			"twitter.com.",
			"netflix.com.",
			"linkedin.com.",
		}

		var wg sync.WaitGroup
		wg.Add(concurrency)

		successCount := 0
		var mu sync.Mutex

		for i := 0; i < concurrency; i++ {
			go func(idx int) {
				defer wg.Done()

				domain := domains[idx%len(domains)]

				// Create DNS query
				m := new(dns.Msg)
				m.SetQuestion(domain, dns.TypeA)

				// Send query
				client := dns.Client{Net: "udp"}
				resp, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))

				// Verify results
				if err == nil && resp != nil && resp.Rcode == dns.RcodeSuccess {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}(i)
		}

		// Wait for all queries to complete
		wg.Wait()

		// Verify that most queries succeeded
		successRate := float64(successCount) / float64(concurrency)
		t.Logf("Success rate: %.2f%%", successRate*100)
		require.True(t, successRate > 0.8, "Load balancing query success rate should be above 80%%")
	})

	// Test failover functionality
	t.Run("Failover", func(t *testing.T) {
		// Create DNS query
		m := new(dns.Msg)
		m.SetQuestion("example.com.", dns.TypeA)

		// Send query
		client := dns.Client{Net: "udp"}
		resp, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, dns.RcodeSuccess, resp.Rcode)

		// Even if some upstream servers are unavailable, the query should still succeed
		// This test mainly verifies that the load balancing system works correctly
		// In practice, we cannot easily simulate upstream server failures in the test
	})

	// Test query performance
	t.Run("QueryPerformance", func(t *testing.T) {
		startTime := time.Now()

		// Perform multiple queries
		queryCount := 10
		for i := 0; i < queryCount; i++ {
			// Create DNS query
			m := new(dns.Msg)
			m.SetQuestion(fmt.Sprintf("test%d.example.com.", i), dns.TypeA)

			// Send query
			client := dns.Client{Net: "udp"}
			_, _, err := client.Exchange(m, fmt.Sprintf("127.0.0.1:%d", testPort))
			require.NoError(t, err)
		}

		// Calculate average query time
		totalTime := time.Since(startTime)
		avgTime := totalTime / time.Duration(queryCount)

		t.Logf("Average query time: %v", avgTime)

		// No hard assertions, just log performance data
	})

	// Wait for server to close gracefully
	sc.SendCloseSignal(nil)
	err = sc.WaitClosed()
	require.NoError(t, err)
}
