package executable_test

import (
	"context"
	"testing"
	"time"

	"github.com/IrineSistiana/mosdns/v5/pkg/query_context"
	fastforward "github.com/IrineSistiana/mosdns/v5/plugin/executable/forward"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

// TestForward tests the DNS forwarding functionality
func TestForward(t *testing.T) {
	// Create forward plugin configuration
	args := &fastforward.Args{
		Upstreams: []fastforward.UpstreamConfig{
			{
				Addr: "tls://223.5.5.5:853", // Use AliDNS as upstream
			},
		},
		Concurrent: 2, // Set number of concurrent queries
	}

	// Initialize forward plugin
	forward, err := fastforward.NewForward(args, fastforward.Opts{})
	require.NoError(t, err)
	defer forward.Close()

	// Create DNS query
	m := new(dns.Msg)
	m.SetQuestion("baidu.com.", dns.TypeA)

	// Create query context
	qCtx := query_context.NewContext(m)

	// Execute forwarding
	err = forward.Exec(context.Background(), qCtx)
	require.NoError(t, err)

	// Verify response
	resp := qCtx.R()
	require.NotNil(t, resp)
	require.Equal(t, dns.RcodeSuccess, resp.Rcode)
	require.True(t, resp.Response)
}

// TestForwardMultipleUpstreams tests the forwarding functionality with multiple upstream servers
func TestForwardMultipleUpstreams(t *testing.T) {
	// Create forward plugin configuration
	args := &fastforward.Args{
		Upstreams: []fastforward.UpstreamConfig{
			{
				Addr: "tls://223.5.5.5:853", // AliDNS
			},
			{
				Addr: "tls://120.53.53.53:853", // DNSPod
			},
		},
		Concurrent: 2, // Set number of concurrent queries
	}

	// Initialize forward plugin
	forward, err := fastforward.NewForward(args, fastforward.Opts{})
	require.NoError(t, err)
	defer forward.Close()

	// Create multiple DNS queries
	domains := []string{
		"example.com.",
		"google.com.",
		"github.com.",
		"amazon.com.",
		"microsoft.com.",
	}

	for _, domain := range domains {
		// Create DNS query
		m := new(dns.Msg)
		m.SetQuestion(domain, dns.TypeA)

		// Create query context
		qCtx := query_context.NewContext(m)

		// Execute forwarding
		err = forward.Exec(context.Background(), qCtx)
		require.NoError(t, err)

		// Verify response
		resp := qCtx.R()
		require.NotNil(t, resp)
		require.Equal(t, dns.RcodeSuccess, resp.Rcode)
		require.True(t, resp.Response)
	}
}

// TestForwardTimeout tests the DNS forwarding timeout handling
func TestForwardTimeout(t *testing.T) {
	// Create forward plugin configuration with a non-existent upstream server
	args := &fastforward.Args{
		Upstreams: []fastforward.UpstreamConfig{
			{
				Addr: "192.0.2.1:53", // Using TEST-NET-1 address, typically unreachable
			},
		},
		Concurrent: 1,
	}

	// Initialize forward plugin
	forward, err := fastforward.NewForward(args, fastforward.Opts{})
	require.NoError(t, err)
	defer forward.Close()

	// Create DNS query
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeA)

	// Create query context
	qCtx := query_context.NewContext(m)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Execute forwarding, should timeout
	err = forward.Exec(ctx, qCtx)

	// May return an error or no response
	if err == nil {
		// If no error is returned, ensure the response is empty or contains an error code
		resp := qCtx.R()
		if resp != nil {
			require.NotEqual(t, dns.RcodeSuccess, resp.Rcode)
		}
	}
}
