package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestDoMergesParamsAndSendsHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "hello world" || r.URL.Query().Get("page") != "2" {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		if r.Header.Get("X-Test") != "yes" {
			t.Errorf("header = %q", r.Header.Get("X-Test"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	response, err := Do(context.Background(), Request{
		Method:  "GET",
		URL:     server.URL,
		Params:  []KeyValue{{Enabled: true, Key: "q", Value: "hello world"}, {Enabled: true, Key: "page", Value: "2"}},
		Headers: []KeyValue{{Enabled: true, Key: "X-Test", Value: "yes"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.StartedAt.IsZero() || response.CompletedAt.Before(response.StartedAt) || response.Duration <= 0 {
		t.Fatalf("timing = %#v", response)
	}
	if response.StatusCode != http.StatusOK || response.Preview != "{\n  \"ok\": true\n}" {
		t.Fatalf("response = %#v", response)
	}
}

func TestDoSetsUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "custom-agent/2.0" {
			t.Fatalf("User-Agent = %q", got)
		}
	}))
	defer server.Close()

	_, err := Do(context.Background(), Request{Method: "GET", URL: server.URL, UserAgent: "custom-agent/2.0"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestDoGeneratesRandomHeader(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Header.Get("X-Request-ID")
		if !pattern.MatchString(value) {
			t.Fatalf("X-Request-ID = %q", value)
		}
	}))
	defer server.Close()

	_, err := Do(context.Background(), Request{
		Method: "GET", URL: server.URL,
		Headers: []KeyValue{{Enabled: true, Key: "X-Request-ID", ValueMode: ValueModeRandom, RandomMode: RandomUUIDv4}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestDoUsesExplicitContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Values("Content-Type"); len(got) != 1 || got[0] != "application/vnd.custom+json" {
			t.Fatalf("Content-Type = %#v", got)
		}
	}))
	defer server.Close()

	_, err := Do(context.Background(), Request{
		Method: "POST", URL: server.URL, BodyMode: BodyRaw, Body: "{}",
		Headers: []KeyValue{{Enabled: true, Key: "Content-Type", Value: "application/vnd.custom+json"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoCanBlockRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("target"))
	}))
	defer server.Close()

	_, err := Do(context.Background(), Request{Method: "GET", URL: server.URL + "/redirect", FollowRedirects: false}, nil)
	if err == nil || !strings.Contains(err.Error(), "redirect blocked") {
		t.Fatalf("err = %v", err)
	}
}
func TestDoUploadsMultipartAndStreams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("note") != "hello" {
			t.Errorf("note = %q", r.FormValue("note"))
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if header.Filename != "upload.txt" {
			t.Errorf("filename = %q", header.Filename)
		}
		_, _ = w.Write([]byte("stream-"))
		_, _ = w.Write([]byte("response"))
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "upload.txt")
	if err := os.WriteFile(path, []byte("file body"), 0o600); err != nil {
		t.Fatal(err)
	}
	var chunks []string
	response, err := Do(context.Background(), Request{
		Method: "POST", URL: server.URL, BodyMode: BodyMultipart,
		FormFields: []KeyValue{{Enabled: true, Key: "note", Value: "hello"}},
		FileField:  "file", FilePath: path, Stream: true,
	}, func(chunk []byte) { chunks = append(chunks, string(chunk)) })
	if err != nil {
		t.Fatal(err)
	}
	if response.Preview != "stream-response" || len(chunks) == 0 {
		t.Fatalf("response = %#v, chunks = %#v", response, chunks)
	}
}

func TestDoSavesDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("download me"))
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "nested", "response.txt")
	response, err := Do(context.Background(), Request{Method: "GET", URL: server.URL, DownloadPath: path}, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "download me" {
		t.Fatalf("saved %q, err %v", data, err)
	}
	if response.SavedTo != path {
		t.Fatalf("SavedTo = %q", response.SavedTo)
	}
}

func TestBuildURLIgnoresEditorLineBreaks(t *testing.T) {
	got, err := buildURL("https://example.com/v1/items\r\n?page=1", []KeyValue{{Enabled: true, Key: "q", Value: "hello world"}})
	if err != nil {
		t.Fatal(err)
	}
	want := "https://example.com/v1/items?page=1&q=hello+world"
	if got != want {
		t.Fatalf("url = %q, want %q", got, want)
	}
}

func TestDoRejectsInvalidURL(t *testing.T) {
	_, err := Do(context.Background(), Request{Method: "GET", URL: "://bad"}, nil)
	if err == nil {
		t.Fatal("expected invalid URL error")
	}
}

func TestDoHonorsCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Do(ctx, Request{Method: "GET", URL: server.URL}, nil)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("err = %v", err)
	}
}
