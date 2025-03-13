# MosDNS Functional Test Cases

This directory contains test cases for core MosDNS functionalities, organized by functional domains.

## Test Directory Structure

- `server_test/`: Functional tests for DNS server protocols (UDP, TCP, DoH, DoQ)
- `matcher_test/`: Validation of DNS request matchers (domain, IP, query type, etc.)
- `executable_test/`: Verification of executable plugins (forwarding, caching, etc.)
- `integration_test/`: End-to-end scenarios testing component interactions

## Functional Coverage

### Server Functionality

- UDP Server: Validate basic DNS query/response functionality
- TCP Server: Test DNS over TCP protocol implementation
- DoH Server: Verify DNS over HTTPS (DoH) compliance
- DoQ Server: Validate DNS over QUIC (DoQ) implementation

### Matchers

- Domain Matcher (qname): Verify domain matching rules
- IP Matcher (client_ip, resp_ip): Test client/response IP matching
- Query Type Matcher (qtype): Validate DNS record type matching
- Random Matcher: Verify probabilistic matching logic

### Executable Plugins

- Forward: Validate DNS request forwarding
- Cache: Test response caching mechanisms
- Redirect: Verify DNS redirection functionality
- Black Hole: Validate request blocking implementation
- ECS Handler: Test EDNS Client Subnet processing

### Integration Tests

- Basic DNS Server: Validate end-to-end DNS service
- Forwarding with Caching: Test combined forwarding and caching
- Domain Filtering: Verify domain-based filtering workflows
- Load Balancing: Validate multi-upstream server distribution
