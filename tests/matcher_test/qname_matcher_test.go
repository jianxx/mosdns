package matcher_test

import (
	"context"
	"testing"

	"github.com/IrineSistiana/mosdns/v5/pkg/matcher/domain"
	"github.com/IrineSistiana/mosdns/v5/pkg/query_context"
	"github.com/IrineSistiana/mosdns/v5/plugin/data_provider/domain_set"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

// matchQName replicates the matching logic from qname package
func matchQName(qCtx *query_context.Context, m domain.Matcher[struct{}]) (bool, error) {
	for _, question := range qCtx.Q().Question {
		if _, ok := m.Match(question.Name); ok {
			return true, nil
		}
	}
	return false, nil
}

// simpleMatcher implements a composite domain matcher
type simpleMatcher struct {
	matchers []domain.Matcher[struct{}]
}

func (m *simpleMatcher) Match(ctx context.Context, qCtx *query_context.Context) (bool, error) {
	return matchQName(qCtx, domain_set.MatcherGroup(m.matchers))
}

// TestQNameMatcher validates basic domain matching functionality
func TestQNameMatcher(t *testing.T) {
	// Initialize different matcher types
	fullMatcher := domain.NewFullMatcher[struct{}]()
	subDomainMatcher := domain.NewSubDomainMatcher[struct{}]()

	// Populate test domains
	_ = fullMatcher.Add("example.com", struct{}{})          // Exact match
	_ = subDomainMatcher.Add("test.com", struct{}{})        // Subdomain match
	_ = subDomainMatcher.Add("sub.example.org", struct{}{}) // Subdomain match

	// Create composite matcher
	m := &simpleMatcher{
		matchers: []domain.Matcher[struct{}]{fullMatcher, subDomainMatcher},
	}

	// Test cases matrix
	testCases := []struct {
		domain   string
		expected bool
	}{
		{"example.com.", true},         // Should match exact
		{"sub.example.com.", false},    // Should not match (exact match only)
		{"test.com.", true},            // Should match subdomain
		{"sub.test.com.", true},        // Should match nested subdomain
		{"deep.sub.test.com.", true},   // Should match deep nested subdomain
		{"sub.example.org.", true},     // Should match subdomain pattern
		{"domain.example.org.", false}, // Should not match
	}

	for _, tc := range testCases {
		// Create query context
		qCtx := query_context.NewContext(new(dns.Msg))
		qCtx.Q().Question = []dns.Question{
			{Name: tc.domain, Qtype: dns.TypeA, Qclass: dns.ClassINET},
		}

		// Execute matching
		matched, err := m.Match(context.Background(), qCtx)
		require.NoError(t, err)
		require.Equal(t, tc.expected, matched, "Mismatch for domain %s", tc.domain)
	}
}

// TestQNameMatcherWithDomainSets verifies integration with external domain sets
func TestQNameMatcherWithDomainSets(t *testing.T) {
	// Initialize matchers
	fullMatcher := domain.NewFullMatcher[struct{}]()
	subDomainMatcher := domain.NewSubDomainMatcher[struct{}]()

	// Configure test domains
	_ = fullMatcher.Add("blocked.com", struct{}{})        // Exact match
	_ = fullMatcher.Add("evil.example.net", struct{}{})   // Exact match
	_ = subDomainMatcher.Add("malicious.org", struct{}{}) // Subdomain match

	// Create composite matcher
	m := &simpleMatcher{
		matchers: []domain.Matcher[struct{}]{fullMatcher, subDomainMatcher},
	}

	// Test cases matrix
	testCases := []struct {
		domain   string
		expected bool
	}{
		{"blocked.com.", true},       // Should match exact
		{"sub.blocked.com.", false},  // Should not match (exact only)
		{"malicious.org.", true},     // Should match subdomain
		{"sub.malicious.org.", true}, // Should match nested subdomain
		{"evil.example.net.", true},  // Should match exact
		{"not-in-list.com.", false},  // Should not match
	}

	for _, tc := range testCases {
		// Create query context
		qCtx := query_context.NewContext(new(dns.Msg))
		qCtx.Q().Question = []dns.Question{
			{Name: tc.domain, Qtype: dns.TypeA, Qclass: dns.ClassINET},
		}

		// Execute matching
		matched, err := m.Match(context.Background(), qCtx)
		require.NoError(t, err)
		require.Equal(t, tc.expected, matched, "Mismatch for domain %s", tc.domain)
	}
}
