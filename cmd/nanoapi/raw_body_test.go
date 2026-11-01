package main

import (
	"strings"
	"testing"
)

func TestRawBodySampleCoversEveryKind(t *testing.T) {
	for _, kind := range []string{"json", "xml", "html", "javascript", "text"} {
		sample := rawBodySample(kind)
		if strings.TrimSpace(sample) == "" {
			t.Fatalf("kind %q has no sample", kind)
		}
		if !isUntouchedRawBody(sample) {
			t.Fatalf("sample for %q is not recognised as a sample", kind)
		}
	}
	if rawBodySample("nonsense") != rawBodySample("json") {
		t.Fatal("unknown kind should fall back to the JSON sample")
	}
}

func TestSyncRawBodySampleSwapsWithKind(t *testing.T) {
	tab := &Tab{BodyRawType: "json", Body: rawBodySample("json")}
	tab.rawKindApplied = "json"

	tab.BodyRawType = "xml"
	tab.syncRawBodySample()
	if tab.Body != rawBodySample("xml") {
		t.Fatalf("expected the XML sample, got %q", tab.Body)
	}

	tab.BodyRawType = "javascript"
	tab.syncRawBodySample()
	if !strings.Contains(tab.Body, "console.log") {
		t.Fatalf("expected the JavaScript sample, got %q", tab.Body)
	}
}

func TestSyncRawBodySampleKeepsEditedBody(t *testing.T) {
	tab := &Tab{BodyRawType: "json", Body: "{\"orderId\": 42}"}
	tab.rawKindApplied = "json"

	tab.BodyRawType = "xml"
	tab.syncRawBodySample()
	if tab.Body != "{\"orderId\": 42}" {
		t.Fatalf("an edited body must survive a kind switch, got %q", tab.Body)
	}
}

func TestSyncRawBodySampleKeepsEmptyBody(t *testing.T) {
	tab := &Tab{BodyRawType: "json", Body: ""}
	tab.rawKindApplied = "json"

	// A stable kind must not refill an editor the user cleared.
	tab.syncRawBodySample()
	if tab.Body != "" {
		t.Fatalf("a cleared body must stay cleared, got %q", tab.Body)
	}
}

func TestSyncRawBodySampleDefaultsEmptyKind(t *testing.T) {
	tab := &Tab{BodyRawType: "", Body: ""}
	tab.syncRawBodySample()
	if tab.BodyRawType != "json" {
		t.Fatalf("expected the kind to default to json, got %q", tab.BodyRawType)
	}
	if tab.Body != rawBodySample("json") {
		t.Fatalf("expected the JSON sample, got %q", tab.Body)
	}
}
