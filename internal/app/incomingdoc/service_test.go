//go:build integration

package incomingdoc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

const baseURL = "http://localhost:5001/api/v1"

func loginCookie(t *testing.T) string {
	t.Helper()
	resp, err := http.Post(baseURL+"/auth/login", "application/json",
		bytes.NewBufferString(`{"usernameOrEmail":"ter","password":"123456"}`))
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

func doRequest(t *testing.T, method, url, body, cookie string) (int, map[string]any) {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}
	req, _ := http.NewRequest(method, url, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.AddCookie(&http.Cookie{Name: "accessToken", Value: cookie})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(raw, &result)
	return resp.StatusCode, result
}

// TestIncomingNoNullOnCreate — after create outgoing doc, the incoming_docs
// for each recipient must have incoming_no = null (not yet received)
func TestIncomingNoNullOnCreate(t *testing.T) {
	cookie := loginCookie(t)

	// Fetch the latest outgoing doc created by this user
	status, res := doRequest(t, "GET", baseURL+"/outgoing-docs?page=1&limit=1", "", cookie)
	if status != 200 {
		t.Fatalf("GET outgoing-docs returned %d: %v", status, res)
	}

	data, _ := res["data"].(map[string]any)
	items, _ := data["items"].([]any)
	if len(items) == 0 {
		t.Skip("no outgoing docs to inspect — create one first")
	}

	doc := items[0].(map[string]any)
	recipients, _ := doc["recipients"].([]any)
	if len(recipients) == 0 {
		t.Skip("doc has no recipients")
	}

	t.Logf("checking %d recipients", len(recipients))
	for i, r := range recipients {
		rec := r.(map[string]any)
		incomingNo, hasKey := rec["incoming_no"]
		if hasKey && incomingNo != nil && incomingNo != "" {
			t.Errorf("recipient[%d] %s: expected incoming_no=null before receive, got %q",
				i, rec["department_name"], incomingNo)
		} else {
			t.Logf("PASS recipient[%d] %s: incoming_no is null/empty ✓", i, rec["department_name"])
		}
	}
}

// TestIncomingNoAssignedOnReceive — after calling /receive, incoming_no must be set
func TestIncomingNoAssignedOnReceive(t *testing.T) {
	cookie := loginCookie(t)

	// Find a pending incoming doc for our department
	status, res := doRequest(t, "GET", baseURL+"/incoming-docs?page=1&limit=10", "", cookie)
	if status != 200 {
		t.Fatalf("GET incoming-docs returned %d: %v", status, res)
	}

	data, _ := res["data"].(map[string]any)
	items, _ := data["items"].([]any)

	var pendingID string
	for _, item := range items {
		doc := item.(map[string]any)
		if doc["status"] == "pending" {
			pendingID = fmt.Sprintf("%v", doc["id"])
			break
		}
	}

	if pendingID == "" {
		t.Skip("no pending incoming doc found — create one first")
	}

	t.Logf("receiving incoming doc id=%s", pendingID)
	url := fmt.Sprintf("%s/incoming-docs/%s/receive", baseURL, pendingID)
	status, res = doRequest(t, "POST", url, `{"remark":"unit test receive"}`, cookie)

	if status != 200 {
		t.Fatalf("POST /receive returned %d: %v", status, res)
	}

	data2, _ := res["data"].(map[string]any)
	incomingNo, _ := data2["incoming_no"].(string)

	if incomingNo == "" {
		t.Error("FAIL: incoming_no is still empty after /receive — should have been generated")
	} else {
		t.Logf("PASS: incoming_no=%q assigned on receive ✓", incomingNo)
	}

	// Also verify status changed
	if data2["status"] != "received" {
		t.Errorf("expected status=received, got %v", data2["status"])
	}
}
