package models

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"syscall"
	"time"
)

var defaultOutboundBlockedCIDRs = mustParseCIDRs([]string{
	// These local/shared ranges are not excluded by IsGlobalUnicast or IsPrivate.
	"0.0.0.0/8",
	"100.64.0.0/10",
	"fec0::/10",
	// Preserve NAT64/6to4 restrictions pending a separate compatibility decision.
	"64:ff9b::/96",
	"64:ff9b:1::/48",
	"2002::/16",
})

func parseOutboundBlockedCIDRs(value string) ([]netip.Prefix, error) {
	var prefixes []netip.Prefix
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			if strings.TrimSpace(value) == "" {
				return nil, nil
			}
			return nil, fmt.Errorf("OUTBOUND_BLOCKED_CIDRS contains an empty entry")
		}
		prefix, err := netip.ParsePrefix(item)
		if err != nil {
			return nil, fmt.Errorf("invalid OUTBOUND_BLOCKED_CIDRS entry %q: %w", item, err)
		}
		if prefix.Addr().Is4In6() {
			if prefix.Bits() < 96 {
				return nil, fmt.Errorf("OUTBOUND_BLOCKED_CIDRS entry %q has an ambiguous IPv4-mapped prefix", item)
			}
			prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

func mustParseCIDRs(values []string) []netip.Prefix {
	prefixes, err := parseOutboundBlockedCIDRs(strings.Join(values, ","))
	if err != nil {
		panic(err)
	}
	return prefixes
}

func newOutboundHTTPTransport(configuredBlockedCIDRs []netip.Prefix) *http.Transport {
	blockedCIDRs := append(append([]netip.Prefix(nil), defaultOutboundBlockedCIDRs...), configuredBlockedCIDRs...)
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		// Control receives the resolved IP for each connection attempt, before connecting.
		Control: func(network, address string, _ syscall.RawConn) error {
			return validateOutboundAddress(network, address, blockedCIDRs)
		},
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = dialer.DialContext
	transport.DialTLSContext = nil
	return transport
}

func validateOutboundAddress(network, address string, blockedCIDRs []netip.Prefix) error {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return fmt.Errorf("outbound network %q is not allowed", network)
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid outbound address %q: %w", address, err)
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return fmt.Errorf("outbound address %q is not an IP address: %w", host, err)
	}
	if ip.Zone() != "" {
		return fmt.Errorf("outbound address %q has an IPv6 zone", host)
	}
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return fmt.Errorf("outbound address %s is not globally routable", ip)
	}
	for _, prefix := range blockedCIDRs {
		if prefix.Contains(ip) {
			return fmt.Errorf("outbound address %s is blocked by %s", ip, prefix)
		}
	}
	return nil
}

type outboundRoundTripper struct {
	transport http.RoundTripper
}

func (transport outboundRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL == nil || (request.URL.Scheme != "http" && request.URL.Scheme != "https") || request.URL.Hostname() == "" {
		return nil, fmt.Errorf("outbound URL must use HTTP or HTTPS and include a host")
	}
	return transport.transport.RoundTrip(request)
}

func (transport outboundRoundTripper) CloseIdleConnections() {
	if closer, ok := transport.transport.(interface{ CloseIdleConnections() }); ok {
		closer.CloseIdleConnections()
	}
}

func (relayConfig *RelayConfig) OutboundHTTPClient(timeout time.Duration) *http.Client {
	transport := relayConfig.outboundTransport
	if transport == nil {
		transport = newOutboundHTTPTransport(relayConfig.outboundBlockedCIDRs)
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: outboundRoundTripper{transport: transport},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
