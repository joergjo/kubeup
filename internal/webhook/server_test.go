package webhook

import (
	"net/http"
	"strings"
	"testing"
)

func TestServerRequiresExplicitAuthorization(t *testing.T) {
	_, err := NewServer(http.NotFoundHandler(), WithPath("/events"), WithPort(8000))
	if err == nil || !strings.Contains(err.Error(), "authorization required") {
		t.Fatalf("expected authorization error, got %v", err)
	}
}
