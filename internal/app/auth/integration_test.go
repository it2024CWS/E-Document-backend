package auth_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

const baseURL = "http://localhost:5001/api/v1"

// loginAndGetCookie logs in as ter/123456 and returns the accessToken cookie value
func loginAndGetCookie(t *testing.T) string {
	t.Helper()

	body := `{"usernameOrEmail":"ter","password":"123456"}`
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("login returned %d: %s", resp.StatusCode, raw)
	}

	for _, c := range resp.Cookies() {
		if c.Name == "accessToken" {
			return c.Value
		}
	}
	t.Fatal("accessToken cookie not found in login response")
	return ""
}

// jwtClaims decodes the payload of a JWT without verifying signature
func jwtClaims(t *testing.T, token string) map[string]interface{} {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid JWT: expected 3 parts, got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode JWT payload: %v", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("failed to unmarshal JWT claims: %v", err)
	}
	return claims
}

// TestIntegration_LoginRoleNameInJWT verifies that after login the JWT carries role_name
func TestIntegration_LoginRoleNameInJWT(t *testing.T) {
	token := loginAndGetCookie(t)

	claims := jwtClaims(t, token)

	roleName, _ := claims["role_name"].(string)
	if roleName == "" {
		t.Error("FAIL: role_name is empty in JWT — FindByUsername does not JOIN user_roles")
		return
	}
	t.Logf("PASS: role_name=%q found in JWT", roleName)
}

// TestIntegration_ReceiveDocument_SecretaryAllowed calls the receive endpoint
// expects 200 (received) or 400 (already received) or 404 (doc not found) — NOT 403
func TestIntegration_ReceiveDocument_SecretaryAllowed(t *testing.T) {
	token := loginAndGetCookie(t)

	incomingDocID := "73abc836-5cc2-4203-863b-5b2be6035833"
	url := fmt.Sprintf("%s/incoming-docs/%s/receive", baseURL, incomingDocID)

	req, _ := http.NewRequest(http.MethodPost, url,
		bytes.NewBufferString(`{"remark":"test receive"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "accessToken", Value: token})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("receive request failed: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 403 {
		t.Errorf("FAIL: got 403 Forbidden — role_name missing from JWT\nBody: %s", raw)
	} else {
		t.Logf("PASS: status=%d (not 403)\nBody: %s", resp.StatusCode, raw)
	}
}
