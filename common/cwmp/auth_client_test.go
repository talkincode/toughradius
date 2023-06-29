package cwmp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConnectionRequestAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/auth" {
			if r.Header.Get("Authorization") == "" {
				w.Header().Set("WWW-Authenticate", `Digest realm="myRealm", nonce="randomNonce", opaque="randomOpaque", qop="auth"`)
				http.Error(w, "Unauthorized.", http.StatusUnauthorized)
				return
			}

			// 验证Authorization头
			auth := r.Header.Get("Authorization")
			if strings.Contains(auth, `username="wronguser"`) {
				http.Error(w, "Unauthorized.", http.StatusUnauthorized)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}
	}))
	defer ts.Close()

	authorized, err := ConnectionRequestAuth("username", "password", ts.URL+"/auth")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !authorized {
		t.Errorf("Expected authorized, but was not")
	}

	authorized, err = ConnectionRequestAuth("wronguser", "wrongpass", ts.URL+"/auth")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if authorized {
		t.Errorf("Expected unauthorized, but was authorized")
	}
}
