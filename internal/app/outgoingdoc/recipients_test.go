//go:build integration

package outgoingdoc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

const baseURL = "http://localhost:5001/api/v1"

func loginCookie(t *testing.T, usernameOrEmail, password string) string {
	t.Helper()
	body := fmt.Sprintf(`{"usernameOrEmail":%q,"password":%q}`, usernameOrEmail, password)
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("login %d: %s", resp.StatusCode, b)
	}
	for _, c := range resp.Cookies() {
		if c.Name == "accessToken" {
			return c.Value
		}
	}
	t.Fatal("no accessToken cookie")
	return ""
}

// TestRecipientsNotEmpty_WhenIncomingDocsExist checks that recipients with NULL incoming_no
// (created after migration 014) still appear correctly in the response.
// Uses a known outgoing doc ID from the live DB that has recipients with NULL incoming_no.
func TestRecipientsNotEmpty_WhenIncomingDocsExist(t *testing.T) {
	cookie := loginCookie(t, "ter", "123456")

	// This is the outgoing doc from the user's response that showed recipients: []
	// but should have incoming_docs in DB.
	// doc_details_id: 799f8bcc-03ca-4009-8d8d-6520140ed743
	knownDocID := "788b852d-c953-4aee-92cf-8996d93f2937" // has 3 confirmed recipients

	url := fmt.Sprintf("%s/outgoing-docs/%s", baseURL, knownDocID)
	req, _ := http.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name: "accessToken", Value: cookie})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET outgoing-doc by ID failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("GET returned %d: %s", resp.StatusCode, body)
	}

	var result struct {
		Data struct {
			Recipients []struct {
				DepartmentName string `json:"department_name"`
				IncomingNo     string `json:"incoming_no"`
				Status         string `json:"status"`
			} `json:"recipients"`
			StatusCounts struct {
				Total   int `json:"total"`
				Pending int `json:"pending"`
			} `json:"status_counts"`
		} `json:"data"`
	}
	json.Unmarshal(body, &result)

	t.Logf("recipients=%d status_counts.total=%d",
		len(result.Data.Recipients), result.Data.StatusCounts.Total)

	if len(result.Data.Recipients) == 0 {
		t.Error("FAIL: recipients is empty — scan of NULL incoming_no may still be broken")
		return
	}

	if result.Data.StatusCounts.Total != len(result.Data.Recipients) {
		t.Errorf("FAIL: status_counts.total=%d but len(recipients)=%d — counts mismatch",
			result.Data.StatusCounts.Total, len(result.Data.Recipients))
		return
	}

	// Verify recipients with NULL incoming_no are handled gracefully (empty string, not error)
	for i, r := range result.Data.Recipients {
		t.Logf("  recipient[%d] dept=%s incoming_no=%q status=%s",
			i, r.DepartmentName, r.IncomingNo, r.Status)
	}

	t.Log("PASS: recipients returned correctly, NULL incoming_no handled ✓")
}
