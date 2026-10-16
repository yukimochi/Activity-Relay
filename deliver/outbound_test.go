package deliver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSenderBlocksLoopback(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	original := HttpClient
	HttpClient = GlobalConfig.OutboundHTTPClient(5 * time.Second)
	t.Cleanup(func() { HttpClient.CloseIdleConnections(); HttpClient = original })
	err := sendActivity(server.URL, RelayActor.PublicKey.ID, []byte(`{}`), GlobalConfig.ActorKey())
	if err == nil || !strings.Contains(err.Error(), "outbound address") || requests.Load() != 0 {
		t.Fatalf("expected policy rejection without delivery: error %v, requests %d", err, requests.Load())
	}
}

func TestSenderDoesNotFollowRedirect(t *testing.T) {
	useLocalDeliveryTransport(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Location", "/target")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	err := sendActivity(server.URL, RelayActor.PublicKey.ID, []byte(`{}`), GlobalConfig.ActorKey())
	if err == nil || !strings.Contains(err.Error(), "307") || requests.Load() != 1 {
		t.Fatalf("expected redirect failure without retry: error %v, requests %d", err, requests.Load())
	}
}

func TestSenderRejectsInvalidURL(t *testing.T) {
	for _, target := range []string{"http://[invalid", "ftp://example.com/inbox"} {
		if err := sendActivity(target, RelayActor.PublicKey.ID, []byte(`{}`), GlobalConfig.ActorKey()); err == nil {
			t.Errorf("invalid inbox %q should fail", target)
		}
	}
}
