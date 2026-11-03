package main

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"nanoapi/internal/httpclient"
	"nanoapi/internal/preview"
	"nanoapi/locales"
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
	tab := a.tab()
	tab.ResponseImage = nil
	tab.PreviewKind = previewNone
	tab.PreviewMessage = ""
	tab.PreviewKind = responsePreviewType(response.ContentType)
	if tab.PreviewKind == previewNone {
		return
	}
	if response.Truncated {
		tab.PreviewMessage = locales.PreviewTooLarge()
		return
	}
	switch tab.PreviewKind {
	case previewImage:
		var image ui.ImageSource
		var err error
		if strings.Contains(strings.ToLower(response.ContentType), "svg+xml") {
			image, err = ui.ParseSVG(response.Body)
		} else {
			image, err = ui.DecodeBitmap(response.Body)
		}
		if err != nil {
			tab.PreviewMessage = locales.ImageDecodeFailed(err.Error())
			return
		}
		tab.ResponseImage = image
	case previewHTML:
		if a.HTMLPreview == nil {
			tab.PreviewMessage = locales.HtmlPreviewUnavailable()
			return
		}
		a.HTMLPreview.Page().LoadHTML(preview.HTML(response.Body), requestURL)
	}
}

func (a *App) responsePreviewView(c *ui.Context, t *ui.Theme) {
	tab := a.tab()
	if tab.Response == nil {
		ui.Text(c, locales.SendToSeePreview()).TextColor(t.TextMuted)
		return
	}
	if tab.PreviewMessage != "" {
		ui.Text(c, tab.PreviewMessage).TextColor(t.Danger).Selectable()
		return
	}
	switch tab.PreviewKind {
	case previewImage:
		if tab.ResponseImage == nil {
			ui.Text(c, locales.ImagePreviewUnavailable()).TextColor(t.TextMuted)
			return
		}
		image := tab.ResponseImage
		ui.Box(c).Grow(1).FillHeight().Background(t.Surface).Border(1, t.Border).Radius(6).Draw(func(p *ui.Painter, r ui.Rect) {
			p.Image(image, r, ui.Contain)
		})
	case previewHTML:
		if a.HTMLPreview == nil {
			ui.Text(c, locales.HtmlPreviewUnavailable()).TextColor(t.TextMuted)
			return
		}
		ui.WebView(c, a.HTMLPreview).Grow(1).FillHeight().Border(1, t.Border).Radius(6)
	default:
		ui.Text(c, locales.NoVisualPreview()).TextColor(t.TextMuted)
	}
}
