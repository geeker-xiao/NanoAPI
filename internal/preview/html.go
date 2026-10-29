package preview

import (
	"bytes"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const contentSecurityPolicy = "default-src 'none'; img-src data: http: https:; style-src 'unsafe-inline'; font-src data:; media-src data:"

var droppedElements = map[string]bool{
	"base": true, "embed": true, "frame": true, "frameset": true,
	"iframe": true, "link": true, "object": true, "script": true,
}

// HTML parses, sanitizes, and wraps response HTML for a restricted preview.
func HTML(input []byte) string {
	document, err := html.Parse(bytes.NewReader(input))
	if err != nil {
		return fallback(string(input))
	}
	sanitize(document)
	ensureHead(document)
	data, err := render(document)
	if err != nil {
		return fallback(string(input))
	}
	return string(data)
}

func sanitize(node *html.Node) {
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		if child.Type == html.ElementNode && droppedElements[strings.ToLower(child.Data)] {
			node.RemoveChild(child)
			child = next
			continue
		}
		if child.Type == html.ElementNode {
			sanitizeAttributes(child)
		}
		sanitize(child)
		child = next
	}
}

func sanitizeAttributes(element *html.Node) {
	attributes := element.Attr[:0]
	for _, attribute := range element.Attr {
		key := strings.ToLower(attribute.Key)
		value := strings.TrimSpace(attribute.Val)
		if strings.HasPrefix(key, "on") || isBlockedURL(value) {
			continue
		}
		if strings.ToLower(element.Data) == "meta" && key == "http-equiv" && strings.EqualFold(value, "refresh") {
			continue
		}
		attributes = append(attributes, attribute)
	}
	element.Attr = attributes
}

func isBlockedURL(value string) bool {
	decoded, err := url.PathUnescape(value)
	if err != nil {
		decoded = value
	}
	decoded = strings.ReplaceAll(decoded, "\x00", "")
	decoded = strings.TrimSpace(strings.ToLower(decoded))
	return strings.HasPrefix(decoded, "javascript:") || strings.HasPrefix(decoded, "vbscript:")
}

func ensureHead(document *html.Node) {
	head := findElement(document, "head")
	if head == nil {
		return
	}
	for child := head.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && strings.EqualFold(child.Data, "meta") {
			for _, attribute := range child.Attr {
				if strings.EqualFold(attribute.Key, "http-equiv") && strings.EqualFold(attribute.Val, "Content-Security-Policy") {
					return
				}
			}
		}
	}
	meta := &html.Node{Type: html.ElementNode, Data: "meta", Attr: []html.Attribute{
		{Key: "http-equiv", Val: "Content-Security-Policy"},
		{Key: "content", Val: contentSecurityPolicy},
	}}
	head.InsertBefore(meta, head.FirstChild)
}

func findElement(node *html.Node, name string) *html.Node {
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, name) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findElement(child, name); found != nil {
			return found
		}
	}
	return nil
}

func render(document *html.Node) ([]byte, error) {
	var buffer bytes.Buffer
	if err := html.Render(&buffer, document); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func fallback(input string) string {
	var builder strings.Builder
	builder.WriteString(`<!doctype html><html><head><meta http-equiv="Content-Security-Policy" content="`)
	builder.WriteString(contentSecurityPolicy)
	builder.WriteString(`"></head><body><pre>`)
	builder.WriteString(html.EscapeString(input))
	builder.WriteString("</pre></body></html>")
	return builder.String()
}
