package main

import "testing"

func TestResponsePreviewType(t *testing.T) {
	cases := map[string]string{
		"image/png":                previewImage,
		"image/svg+xml":            previewImage,
		"text/html; charset=utf-8": previewHTML,
		"application/xhtml+xml":    previewHTML,
		"application/json":         previewNone,
		"":                         previewNone,
	}
	for contentType, want := range cases {
		if got := responsePreviewType(contentType); got != want {
			t.Fatalf("responsePreviewType(%q) = %q, want %q", contentType, got, want)
		}
	}
}
