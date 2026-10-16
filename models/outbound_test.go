package models

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

func TestOutboundCIDRParsing(t *testing.T) {
	prefixesWithSpaces, err := parseOutboundBlockedCIDRs(" 8.8.8.9/24, 2606:4700::/32 ")
	if err != nil || len(prefixesWithSpaces) != 2 || prefixesWithSpaces[0].String() != "8.8.8.0/24" {
		t.Fatalf("CIDR trimming and masking = %v, %v", prefixesWithSpaces, err)
	}
	for _, value := range []string{"", "  "} {
		prefixes, err := parseOutboundBlockedCIDRs(value)
		if err != nil || len(prefixes) != 0 {
			t.Fatalf("parse(%q) = %v, %v", value, prefixes, err)
		}
	}
	for _, value := range []string{"invalid", "8.8.8.8", "8.8.8.8/33", "8.8.8.8/32,", ",", "::ffff:8.8.8.8/80"} {
		if _, err := parseOutboundBlockedCIDRs(value); err == nil {
			t.Errorf("parse(%q) should fail", value)
		}
	}
	prefixes, err := parseOutboundBlockedCIDRs("::ffff:8.8.8.8/128")
	if err != nil || len(prefixes) != 1 || prefixes[0].String() != "8.8.8.8/32" {
		t.Fatalf("IPv4-mapped CIDR normalization = %v, %v", prefixes, err)
	}
}

func TestOutboundCIDRsFromYAMLAndEnvironment(t *testing.T) {
	config := viper.New()
	config.SetConfigType("yaml")
	if err := config.ReadConfig(strings.NewReader("OUTBOUND_BLOCKED_CIDRS: '8.8.8.8/32, 2606:4700::/32'")); err != nil {
		t.Fatal(err)
	}
	if err := config.BindEnv("OUTBOUND_BLOCKED_CIDRS"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OUTBOUND_BLOCKED_CIDRS", "")
	prefixes, err := parseOutboundBlockedCIDRs(config.GetString("OUTBOUND_BLOCKED_CIDRS"))
	if err != nil || len(prefixes) != 2 {
		t.Fatalf("YAML CIDRs = %v, %v", prefixes, err)
	}
	t.Setenv("OUTBOUND_BLOCKED_CIDRS", "1.1.1.1/32")
	prefixes, err = parseOutboundBlockedCIDRs(config.GetString("OUTBOUND_BLOCKED_CIDRS"))
	if err != nil || len(prefixes) != 1 || prefixes[0].String() != "1.1.1.1/32" {
		t.Fatalf("environment should override YAML: %v, %v", prefixes, err)
	}
}

func TestOutboundAddressPolicy(t *testing.T) {
	blocked := append(append([]netip.Prefix(nil), defaultOutboundBlockedCIDRs...),
		netip.MustParsePrefix("8.8.8.0/24"), netip.MustParsePrefix("2606:4700::/32"))
	for _, host := range []string{
		"0.0.0.0", "0.1.2.3", "10.0.0.1", "100.64.0.1", "127.0.0.1", "169.254.169.254",
		"172.16.1.1", "192.168.1.1", "224.0.0.1", "255.255.255.255",
		"::", "::1", "fc00::1", "fe80::1", "ff02::1", "fec0::1", "2002:7f00:1::1",
		"64:ff9b::7f00:1", "64:ff9b:1::1", "::ffff:127.0.0.1", "8.8.8.8", "::ffff:8.8.8.8", "2606:4700::1111",
	} {
		t.Run(host, func(t *testing.T) {
			if err := validateOutboundAddress("tcp", net.JoinHostPort(host, "443"), blocked); err == nil {
				t.Fatal("expected address to be blocked")
			}
		})
	}
	for _, host := range []string{
		"1.1.1.1", "::ffff:1.1.1.1", "2001:4860:4860::8888",
		"192.0.0.1", "192.0.2.1", "192.88.99.1", "198.18.0.1", "198.51.100.1", "203.0.113.1", "240.0.0.1",
		"::192.0.2.1", "100::1", "2001::1", "2001:db8::1", "3fff::1", "5f00::1",
	} {
		if err := validateOutboundAddress("tcp", net.JoinHostPort(host, "443"), blocked); err != nil {
			t.Errorf("public address %s rejected: %v", host, err)
		}
	}
	if err := validateOutboundAddress("tcp4", "invalid", blocked); err == nil {
		t.Fatal("malformed address should fail")
	}
}

func TestOutboundTransportConfiguration(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://proxy.invalid:8080")
	t.Setenv("HTTPS_PROXY", "http://proxy.invalid:8080")
	config := &RelayConfig{outboundTransport: newOutboundHTTPTransport(nil)}
	fetch := config.OutboundHTTPClient(10 * time.Second)
	sender := config.OutboundHTTPClient(5 * time.Second)
	if fetch.Transport.(outboundRoundTripper).transport != sender.Transport.(outboundRoundTripper).transport {
		t.Fatal("fetch and delivery must share the configured transport")
	}
	if config.outboundTransport.Proxy != nil || config.outboundTransport.DialContext == nil ||
		config.outboundTransport.DialTLS != nil || config.outboundTransport.DialTLSContext != nil {
		t.Fatal("unexpected outbound transport configuration")
	}
	if fetch.Timeout != 10*time.Second || sender.Timeout != 5*time.Second {
		t.Fatal("client timeouts were not preserved")
	}
	fetch.CloseIdleConnections()
}

func TestOutboundBlocksResolvedLoopback(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	client := (&RelayConfig{}).OutboundHTTPClient(time.Second)
	defer client.CloseIdleConnections()
	for _, scheme := range []string{"http", "https"} {
		for _, host := range []string{"127.0.0.1", "localhost"} {
			_, err := client.Get(scheme + "://" + net.JoinHostPort(host, port))
			if err == nil || !strings.Contains(err.Error(), "outbound address") {
				t.Errorf("%s %s: expected policy rejection, got %v", scheme, host, err)
			}
		}
	}
}

func TestOutboundRedirectIsNotFollowed(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Location", "http://127.0.0.1/target")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()
	transport := newOutboundHTTPTransport(nil)
	// Only the test replaces dialing, so the redirect policy can be exercised locally.
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}
	client := (&RelayConfig{outboundTransport: transport}).OutboundHTTPClient(time.Second)
	defer client.CloseIdleConnections()
	response, err := client.Get("http://actor.example/actor")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusFound || requests.Load() != 1 {
		t.Fatalf("status %d, requests %d", response.StatusCode, requests.Load())
	}
}

func TestRemoteFetchUsesConfiguredClient(t *testing.T) {
	original := httpClient
	t.Cleanup(func() { httpClient = original })
	config := &RelayConfig{outboundBlockedCIDRs: []netip.Prefix{netip.MustParsePrefix("8.8.8.8/32")}}
	httpClient = config.OutboundHTTPClient(time.Second)
	defer httpClient.CloseIdleConnections()
	actorCache := cache.New(time.Minute, time.Minute)
	if _, err := NewActivityPubActorFromRemoteActor("http://8.8.8.8/actor", "test", actorCache, "", nil); err == nil || !strings.Contains(err.Error(), "blocked by") {
		t.Fatalf("actor fetch must enforce configured CIDRs: %v", err)
	}
	if _, err := NewActivityPubActivityFromRemoteActivity("http://8.8.8.8/activity", "test"); err == nil || !strings.Contains(err.Error(), "blocked by") {
		t.Fatalf("activity fetch must enforce configured CIDRs: %v", err)
	}
	for _, target := range []string{"http://[invalid", "ftp://example.com/actor"} {
		if _, err := NewActivityPubActorFromRemoteActor(target, "test", actorCache, "", nil); err == nil {
			t.Errorf("invalid actor URL %q should fail", target)
		}
		if _, err := NewActivityPubActivityFromRemoteActivity(target, "test"); err == nil {
			t.Errorf("invalid activity URL %q should fail", target)
		}
	}
}

func TestRemoteFetchPreservesSigning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("Signature") == "" || r.Header.Get("Accept") != "application/activity+json" {
			t.Error("expected signed ActivityPub GET")
		}
		io.WriteString(w, `{"id":"https://actor.example/actor"}`)
	}))
	defer server.Close()
	transport := newOutboundHTTPTransport(nil)
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}
	original := httpClient
	httpClient = (&RelayConfig{outboundTransport: transport}).OutboundHTTPClient(time.Second)
	t.Cleanup(func() { httpClient = original; transport.CloseIdleConnections() })
	actor, err := NewActivityPubActorFromRemoteActor("http://actor.example/actor", "test", cache.New(time.Minute, time.Minute), "https://relay.example/actor#key", globalConfig.ActorKey())
	if err != nil || actor.ID != "https://actor.example/actor" {
		t.Fatalf("actor = %v, error = %v", actor, err)
	}
}
