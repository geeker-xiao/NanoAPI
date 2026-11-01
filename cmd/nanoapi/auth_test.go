package main

import (
	"encoding/base64"
	"testing"
)

func TestAuthBearerHeader(t *testing.T) {
	auth := Auth{Mode: "Bearer token", Token: "abc123"}
	row, ok := auth.header()
	if !ok {
		t.Fatal("expected a bearer header")
	}
	if row.Key != "Authorization" || row.Value != "Bearer abc123" || !row.Enabled {
		t.Fatalf("unexpected row: %+v", row)
	}
	if _, ok := auth.param(); ok {
		t.Fatal("a bearer token is not a query parameter")
	}
}

func TestAuthBearerNeedsAToken(t *testing.T) {
	if _, ok := (Auth{Mode: "Bearer token"}).header(); ok {
		t.Fatal("an empty token should not send a header")
	}
}

func TestAuthBasicHeader(t *testing.T) {
	auth := Auth{Mode: "Basic auth", Username: "user", Password: "pass"}
	row, ok := auth.header()
	if !ok {
		t.Fatal("expected a basic header")
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
	if row.Value != want {
		t.Fatalf("got %q, want %q", row.Value, want)
	}
}

func TestAuthAPIKeyTravelsWhereItIsTold(t *testing.T) {
	header := Auth{Mode: "API key", KeyName: "X-API-Key", KeyValue: "secret", KeyIn: "Header"}
	row, ok := header.header()
	if !ok || row.Key != "X-API-Key" || row.Value != "secret" {
		t.Fatalf("header expected the key row, got %+v", row)
	}
	if _, ok := header.param(); ok {
		t.Fatal("a header key should not also be a parameter")
	}
	query := Auth{Mode: "API key", KeyName: "api_key", KeyValue: "secret", KeyIn: "Query"}
	row, ok = query.param()
	if !ok || row.Key != "api_key" || row.Value != "secret" {
		t.Fatalf("query expected the key row, got %+v", row)
	}
	if _, ok := query.header(); ok {
		t.Fatal("a query key should not also be a header")
	}
}

func TestWithAuthKeepsHandTypedRows(t *testing.T) {
	tab := newTab()
	tab.Headers = []KeyValue{{Enabled: true, Key: "Authorization", Value: "Bearer mine"}}
	tab.Auth = Auth{Mode: "Bearer token", Token: "theirs"}
	_, headers := withAuth(tab)
	if len(headers) != 1 || headers[0].Value != "Bearer mine" {
		t.Fatalf("expected the hand-typed row to win, got %+v", headers)
	}
}

func TestWithAuthAddsRows(t *testing.T) {
	tab := newTab()
	tab.Params = []KeyValue{{Enabled: true, Key: "page", Value: "1"}}
	tab.Auth = Auth{Mode: "API key", KeyName: "api_key", KeyValue: "secret", KeyIn: "Query"}
	params, headers := withAuth(tab)
	if len(params) != 2 {
		t.Fatalf("expected the key to join the parameters, got %+v", params)
	}
	if len(headers) != 1 {
		t.Fatalf("expected the header list untouched, got %+v", headers)
	}
}

func TestAuthEditorShowsFieldsPerMode(t *testing.T) {
	// The modes drive which fields appear, so a typo would silently hide one.
	for _, mode := range authModes {
		auth := Auth{Mode: mode}
		if mode == "Bearer token" {
			auth.Token = "t"
		}
		if got := auth.summary(); got == "" {
			t.Fatalf("no summary for mode %q", mode)
		}
	}
	if _, ok := (Auth{Mode: "None"}).header(); ok {
		t.Fatal("no credential should add a header")
	}
}

func TestPutKVMatchesKeysWithoutCase(t *testing.T) {
	rows := putKV([]KeyValue{{Key: "authorization"}}, KeyValue{Key: "Authorization"})
	if len(rows) != 1 {
		t.Fatalf("expected the row to be recognised, got %+v", rows)
	}
}

func TestTabRoundTripsThroughSavedRequest(t *testing.T) {
	tab := newTab()
	tab.Method = "POST"
	tab.URL = "https://api.example.com/v1/items"
	tab.Params = []KeyValue{{Enabled: true, Key: "page", Value: "2"}}
	tab.Headers = []KeyValue{{Enabled: true, Key: "X-Trace", Value: "abc"}}
	tab.Auth = Auth{Mode: "Bearer token", Token: "tok"}
	tab.UserAgent = "NanoAPI/9"
	tab.BodyMode = "raw"
	tab.BodyRawType = "xml"
	tab.Body = "<a/>"
	tab.TimeoutSeconds = 12

	restored := newTabFromSaved(savedRequestFromTab(tab))
	if restored.Method != "POST" || restored.URL != tab.URL {
		t.Fatalf("request identity lost: %+v", restored)
	}
	if len(restored.Params) != 1 || restored.Params[0].Value != "2" {
		t.Fatalf("params lost: %+v", restored.Params)
	}
	if len(restored.Headers) != 1 || restored.Headers[0].Key != "X-Trace" {
		t.Fatalf("headers lost: %+v", restored.Headers)
	}
	if restored.Auth.Token != "tok" || restored.UserAgent != "NanoAPI/9" {
		t.Fatalf("credential or agent lost: %+v", restored)
	}
	if restored.Body != "<a/>" || restored.BodyRawType != "xml" {
		t.Fatalf("body lost: %+v", restored)
	}
	if restored.TimeoutSeconds != 12 {
		t.Fatalf("timeout lost: %+v", restored)
	}
	// Editing the copy must not reach back into the saved request.
	restored.Params[0].Value = "3"
}

func TestTabSnapshotKeepsSavedIDAndTabIDApart(t *testing.T) {
	tab := newTab()
	tab.SavedID = "req-1"
	snapshot := tabSnapshot(tab)
	if snapshot.ID != "req-1" {
		t.Fatalf("saved id lost: %q", snapshot.ID)
	}
	if snapshot.TabID != tab.ID {
		t.Fatalf("tab id lost: %q", snapshot.TabID)
	}
	reopened := newTabFromSaved(snapshot)
	if reopened.SavedID != "req-1" {
		t.Fatalf("reopening should keep the link, got %q", reopened.SavedID)
	}
}

func TestNormalizeTabFillsMissingPieces(t *testing.T) {
	tab := &Tab{}
	normalizeTab(tab)
	if tab.Method != "GET" || tab.UserAgent != "NanoAPI/1.0" {
		t.Fatalf("request defaults missing: %+v", tab)
	}
	if tab.Auth.Mode != "None" || tab.Auth.KeyIn != "Header" {
		t.Fatalf("auth defaults missing: %+v", tab.Auth)
	}
	if tab.TimeoutSeconds != 30 || tab.FileField != "file" {
		t.Fatalf("option defaults missing: %+v", tab)
	}
	if len(tab.Params) == 0 || len(tab.FormFields) == 0 {
		t.Fatal("expected one row to start with")
	}
}
