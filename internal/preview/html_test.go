package preview

import (
	"strings"
	"testing"
)

func TestHTMLRemovesActiveContent(t *testing.T) {
	got := HTML([]byte(`<!doctype html><html><head><title>Preview</title></head><body><h1 id="title">Hello</h1><script>alert(1)</script><a href="javascript:alert(1)" onclick="alert(2)">link</a><iframe src="https://example.com"></iframe></body></html>`))
	for _, unwanted := range []string{"<script", "javascript:", "onclick", "<iframe"} {
		if strings.Contains(strings.ToLower(got), strings.ToLower(unwanted)) {
			t.Fatalf("HTML contains %q:\n%s", unwanted, got)
		}
	}
	if !strings.Contains(got, "Content-Security-Policy") || !strings.Contains(got, "Hello") {
		t.Fatalf("HTML = %s", got)
	}
}

func TestHTMLFallsBackToEscapedText(t *testing.T) {
	got := HTML([]byte("<p>plain</p>"))
	if !strings.Contains(got, "<p>plain</p>") {
		t.Fatalf("HTML = %s", got)
	}
}
