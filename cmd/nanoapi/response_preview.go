package main

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"nanoapi/internal/httpclient"
	"nanoapi/internal/preview"
)

const (
	previewNone  = ""
	previewImage = "image"
	previewHTML  = "html"
)

func responsePreviewType(contentType string) string {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return previewImage
	case strings.Contains(contentType, "text/html"), strings.Contains(contentType, "application/xhtml+xml"):
		return previewHTML
	default:
		return previewNone
	}
}

func (a *App) setResponsePreview(requestURL string, response *httpclient.Response) {
	a.ResponseImage = nil
	a.PreviewKind = previewNone
	a.PreviewMessage = ""
	a.PreviewKind = responsePreviewType(response.ContentType)
	if a.PreviewKind == previewNone {
		return
	}
	if response.Truncated {
		a.PreviewMessage = "Preview unavailable because the response exceeds the in-memory preview limit."
		return
	}
	switch a.PreviewKind {
	case previewImage:
		var image ui.ImageSource
		var err error
		if strings.Contains(strings.ToLower(response.ContentType), "svg+xml") {
			image, err = ui.ParseSVG(response.Body)
		} else {
			image, err = ui.DecodeBitmap(response.Body)
		}
		if err != nil {
			a.PreviewMessage = "The image response could not be decoded: " + err.Error()
			return
		}
		a.ResponseImage = image
	case previewHTML:
		if a.HTMLPreview == nil {
			a.PreviewMessage = "HTML preview is unavailable in this window."
			return
		}
		a.HTMLPreview.Page().LoadHTML(preview.HTML(response.Body), requestURL)
	}
}

func (a *App) responsePreviewView(c *ui.Context, t *ui.Theme) {
	if a.Response == nil {
		ui.Text(c, "Send a request to see a visual preview.").TextColor(t.TextMuted)
		return
	}
	if a.PreviewMessage != "" {
		ui.Text(c, a.PreviewMessage).TextColor(t.Danger).Selectable()
		return
	}
	switch a.PreviewKind {
	case previewImage:
		if a.ResponseImage == nil {
			ui.Text(c, "Image preview is unavailable.").TextColor(t.TextMuted)
			return
		}
		image := a.ResponseImage
		ui.Box(c).Grow(1).FillHeight().Background(t.Surface).Border(1, t.Border).Radius(6).Draw(func(p *ui.Painter, r ui.Rect) {
			p.Image(image, r, ui.Contain)
		})
	case previewHTML:
		if a.HTMLPreview == nil {
			ui.Text(c, "HTML preview is unavailable in this window.").TextColor(t.TextMuted)
			return
		}
		ui.WebView(c, a.HTMLPreview).Grow(1).FillHeight().Border(1, t.Border).Radius(6)
	default:
		ui.Text(c, "No visual preview is available for this Content-Type. Use Body to inspect the response.").TextColor(t.TextMuted)
	}
}
