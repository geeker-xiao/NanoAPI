package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// KeyValue is one enabled request parameter or header.
type KeyValue struct {
	Enabled    bool   `json:"enabled"`
	Key        string `json:"key"`
	Value      string `json:"value,omitempty"`
	ValueMode  string `json:"valueMode,omitempty"`
	RandomMode string `json:"randomMode,omitempty"`
	Pattern    string `json:"pattern,omitempty"`
}

// BodyMode
type BodyMode string

const (
	BodyNone      BodyMode = "none"
	BodyRaw       BodyMode = "raw"
	BodyForm      BodyMode = "form"
	BodyMultipart BodyMode = "multipart"
)

// Request describes an HTTP request in the UI-independent execution layer.
type Request struct {
	Method          string        `json:"method"`
	URL             string        `json:"url"`
	Params          []KeyValue    `json:"params,omitempty"`
	Headers         []KeyValue    `json:"headers,omitempty"`
	UserAgent       string        `json:"userAgent,omitempty"`
	BodyMode        BodyMode      `json:"bodyMode,omitempty"`
	Body            string        `json:"body,omitempty"`
	FormFields      []KeyValue    `json:"formFields,omitempty"`
	FileField       string        `json:"fileField,omitempty"`
	FilePath        string        `json:"filePath,omitempty"`
	Timeout         time.Duration `json:"timeout,omitempty"`
	FollowRedirects bool          `json:"followRedirects"`
	Stream          bool          `json:"stream"`
	DownloadPath    string        `json:"downloadPath,omitempty"`
}

// Response is the result of a request. Body is bounded for in-memory display;
// DownloadPath and SavedTo describe a complete response written to disk.
type Response struct {
	Status      string
	StatusCode  int
	Headers     []KeyValue
	Body        []byte
	Preview     string
	ContentType string
	BytesRead   int64
	StartedAt   time.Time
	CompletedAt time.Time
	Duration    time.Duration
	Truncated   bool
	SavedTo     string
}

const (
	maxInMemoryBody = 8 << 20
	maxPreviewBody  = 512 << 10
)

var (
	ErrInvalidURL = errors.New("request URL is invalid")
	ErrNoMethod   = errors.New("request method is empty")
)

// OnChunk receives response bytes as they arrive. It is optional and is used
// by the UI for streaming responses.
type OnChunk func([]byte)

// Do executes req and reports transport errors. HTTP status codes are returned
// as Response values rather than errors.
func Do(ctx context.Context, req Request, onChunk OnChunk) (*Response, error) {
	if strings.TrimSpace(req.Method) == "" {
		return nil, ErrNoMethod
	}
	target, err := buildURL(req.URL, req.Params)
	if err != nil {
		return nil, err
	}

	body, contentType, bodyLength, bodyCleanup, err := buildBody(req)
	if bodyCleanup != nil {
		defer bodyCleanup()
	}
	if err != nil {
		return nil, err
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	request, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, err
	}
	if bodyLength >= 0 {
		request.ContentLength = bodyLength
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if strings.TrimSpace(req.UserAgent) != "" {
		request.Header.Set("User-Agent", req.UserAgent)
	}
	for _, header := range req.Headers {
		if !header.Enabled || strings.TrimSpace(header.Key) == "" {
			continue
		}
		key := strings.TrimSpace(header.Key)
		if strings.EqualFold(key, "Content-Type") {
			request.Header.Set(key, ResolveValue(header, time.Now()))
			continue
		}
		request.Header.Add(key, ResolveValue(header, time.Now()))
	}

	transport := http.DefaultTransport
	base := http.DefaultTransport
	if req.Stream && req.Timeout > 0 {
		clone := http.DefaultTransport.(*http.Transport).Clone()
		clone.ResponseHeaderTimeout = req.Timeout
		base = clone
	}
	if !req.FollowRedirects {
		transport = &redirectBlockingTransport{next: base}
	}
	client := &http.Client{Transport: transport}
	if req.Timeout > 0 && !req.Stream {
		client.Timeout = req.Timeout
	}

	start := time.Now()
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &Response{
		Status:      resp.Status,
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Headers:     headersFrom(resp.Header),
	}
	var output io.Writer
	var file *os.File
	if req.DownloadPath != "" {
		if err := os.MkdirAll(filepath.Dir(req.DownloadPath), 0o755); err != nil {
			return nil, fmt.Errorf("create download directory: %w", err)
		}
		file, err = os.Create(req.DownloadPath)
		if err != nil {
			return nil, fmt.Errorf("create download file: %w", err)
		}
		defer file.Close()
		output = file
	}

	var captured []byte
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			result.BytesRead += int64(n)
			if onChunk != nil && req.Stream {
				onChunk(chunk)
			}
			if output != nil {
				if _, err := output.Write(chunk); err != nil {
					return nil, fmt.Errorf("write download file: %w", err)
				}
			}
			if len(captured) < maxInMemoryBody {
				remaining := maxInMemoryBody - len(captured)
				if n > remaining {
					captured = append(captured, chunk[:remaining]...)
					result.Truncated = true
				} else {
					captured = append(captured, chunk...)
				}
			} else {
				result.Truncated = true
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, readErr
		}
	}
	result.CompletedAt = time.Now()
	result.StartedAt = start
	result.Duration = result.CompletedAt.Sub(start)
	result.Body = captured
	result.Preview = preview(captured, result.ContentType, result.Truncated)
	if file != nil {
		result.SavedTo = req.DownloadPath
	}
	return result, nil
}

func buildURL(raw string, params []KeyValue) (string, error) {
	raw = strings.ReplaceAll(raw, "\r", "")
	raw = strings.ReplaceAll(raw, "\n", "")
	raw = strings.ReplaceAll(raw, "\t", "")
	if raw == "" {
		return "", ErrInvalidURL
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidURL
	}
	query := parsed.Query()
	for _, item := range params {
		if !item.Enabled || strings.TrimSpace(item.Key) == "" {
			continue
		}
		query.Add(strings.TrimSpace(item.Key), item.Value)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func buildBody(req Request) (io.Reader, string, int64, func(), error) {
	switch req.BodyMode {
	case BodyNone, "":
		return nil, "", 0, nil, nil
	case BodyRaw:
		contentType := "text/plain; charset=utf-8"
		if strings.HasPrefix(strings.TrimSpace(req.Body), "{") || strings.HasPrefix(strings.TrimSpace(req.Body), "[") {
			contentType = "application/json; charset=utf-8"
		}
		return strings.NewReader(req.Body), contentType, int64(len(req.Body)), nil, nil
	case BodyForm:
		values := url.Values{}
		for _, field := range req.FormFields {
			if field.Enabled && strings.TrimSpace(field.Key) != "" {
				values.Add(strings.TrimSpace(field.Key), field.Value)
			}
		}
		data := values.Encode()
		return strings.NewReader(data), "application/x-www-form-urlencoded", int64(len(data)), nil, nil
	case BodyMultipart:
		if strings.TrimSpace(req.FilePath) == "" {
			values := url.Values{}
			for _, field := range req.FormFields {
				if field.Enabled && strings.TrimSpace(field.Key) != "" {
					values.Add(strings.TrimSpace(field.Key), field.Value)
				}
			}
			data := values.Encode()
			return strings.NewReader(data), "application/x-www-form-urlencoded", int64(len(data)), nil, nil
		}
		file, err := os.Open(req.FilePath)
		if err != nil {
			return nil, "", 0, nil, fmt.Errorf("open upload file: %w", err)
		}
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		go func() {
			defer pw.Close()
			defer file.Close()
			for _, field := range req.FormFields {
				if field.Enabled && strings.TrimSpace(field.Key) != "" {
					_ = writer.WriteField(strings.TrimSpace(field.Key), field.Value)
				}
			}
			name := filepath.Base(req.FilePath)
			part, err := writer.CreateFormFile(strings.TrimSpace(orDefault(req.FileField, "file")), name)
			if err == nil {
				_, err = io.Copy(part, file)
			}
			closeErr := writer.Close()
			if err == nil {
				err = closeErr
			}
			if err != nil {
				_ = pw.CloseWithError(err)
			}
		}()
		return pr, writer.FormDataContentType(), -1, nil, nil
	default:
		return nil, "", 0, nil, fmt.Errorf("unsupported body mode %q", req.BodyMode)
	}
}

type redirectBlockingTransport struct{ next http.RoundTripper }

func (t redirectBlockingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		if resp.Header.Get("Location") != "" {
			resp.Body.Close()
			return nil, fmt.Errorf("redirect blocked with status %s", resp.Status)
		}
	}
	return resp, nil
}

func headersFrom(header http.Header) []KeyValue {
	keys := make([]string, 0, len(header))
	for key := range header {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]KeyValue, 0, len(header))
	for _, key := range keys {
		for _, value := range header[key] {
			result = append(result, KeyValue{Enabled: true, Key: key, Value: value})
		}
	}
	return result
}

func preview(body []byte, contentType string, truncated bool) string {
	if len(body) == 0 {
		if truncated {
			return "(response body is larger than the in-memory preview limit)"
		}
		return ""
	}
	if strings.Contains(strings.ToLower(contentType), "json") {
		return prettyJSON(body, truncated)
	}
	text := string(body)
	if !isPrintable(text) {
		return fmt.Sprintf("(binary response: %d bytes; save it to a file to inspect it)", len(body))
	}
	if truncated {
		text += "\n\n... preview truncated ..."
	}
	return text
}

func orDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
