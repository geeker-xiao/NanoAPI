package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"nanoapi/internal/httpclient"
)

var (
	iconSend     = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m22 2-7 20-4-9-9-4Z"/><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M22 2 11 13"/></svg>`))
	iconSave     = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2Z"/><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 21v-8H7v8M7 3v5h8"/></svg>`))
	iconUpload   = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v12m0-12 4 4m-4-4L8 7"/><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 14v5a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-5"/></svg>`))
	iconDownload = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v12m0 0 4-4m-4 4-4-4"/><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 14v5a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-5"/></svg>`))
	iconPlus     = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M12 5v14M5 12h14"/></svg>`))
	iconTrash    = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7h16m-10 4v6m4-6v6M9 7V4h6v3m-9 0 1 13a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-13"/></svg>`))
	iconCopy     = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect x="9" y="9" width="11" height="11" rx="2" fill="none" stroke="currentColor" stroke-width="2"/><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>`))
	iconCancel   = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M6 6l12 12M18 6 6 18"/></svg>`))
)

type KeyValue = httpclient.KeyValue

type HistoryItem struct {
	ID          string
	GroupID     string
	Label       string
	Request     httpclient.Request
	Status      string
	CreatedAt   time.Time
	RequestedAt time.Time
	RespondedAt time.Time
	Duration    time.Duration
}

type RequestGroup struct {
	ID          string
	Name        string
	Description string
}

type App struct {
	mu sync.Mutex

	Method    string
	URL       string
	Params    []KeyValue
	Headers   []KeyValue
	UserAgent string

	RequestTab  int
	ResponseTab int

	BodyMode   string
	Body       string
	FormFields []KeyValue
	FileField  string
	FilePath   string

	TimeoutSeconds  float64
	TimeoutText     string
	FollowRedirects bool
	Stream          bool
	Download        bool
	DownloadPath    string

	ThemeMode        string
	FontFamily       string
	FontSize         float64
	BackgroundMode   string
	BackgroundPath   string
	BackgroundColor  ui.Color
	TabColor         ui.Color
	backgroundBitmap *ui.Bitmap
	HTMLPreview      *mygo.WebView
	ResponseImage    ui.ImageSource
	PreviewKind      string
	PreviewMessage   string

	Running      bool
	Error        string
	Status       string
	ResponseMeta string
	ResponseText string
	HeadersText  string
	Response     *httpclient.Response

	History []HistoryItem
	Groups  []RequestGroup

	SelectedGroup         string
	HistorySearch         string
	GroupDialogOpen       bool
	GroupDraftID          string
	GroupDraftName        string
	GroupDraftDescription string
	GroupDraftError       string
	GroupDeleteOpen       bool
	GroupDeleteID         string

	cancel    context.CancelFunc
	update    func(func())
	lastChunk time.Time
}

func main() {
	app := &App{
		Method:          "GET",
		URL:             "https://httpbin.org/get",
		Params:          []KeyValue{{Enabled: true, Key: "", Value: ""}},
		Headers:         []KeyValue{{Enabled: true, Key: "Accept", Value: "application/json", ValueMode: "Fixed", RandomMode: "UUID v4"}},
		UserAgent:       "NanoAPI/1.0",
		BodyMode:        "none",
		Body:            "{\n  \"message\": \"Hello from NanoAPI\"\n}",
		FormFields:      []KeyValue{{Enabled: true, Key: "", Value: ""}},
		FileField:       "file",
		TimeoutSeconds:  30,
		TimeoutText:     "30",
		FollowRedirects: true,
		SelectedGroup:   "all",
		ThemeMode:       "System",
		FontFamily:      "System",
		FontSize:        14,
		BackgroundMode:  "Solid color",
		BackgroundColor: ui.RGB(245, 247, 250),
		TabColor:        ui.Hex("#149998"),
	}
	app.loadSettings()
	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:   "NanoAPI",
			Width:   1440,
			Height:  900,
			Content: ui.View(app.view),
		})
		_ = win.SetIcon(appIconPNG)
		app.loadBackgroundImage()
		app.HTMLPreview, _ = win.NewWebView(mygo.WebViewOptions{BackgroundColor: "light-dark(#ffffff, #1e1e1e)"})
		app.update = win.Update
		win.SetMinimumSize(980, 680)
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *App) view(c *ui.Context) {
	a.applyTheme(c)
	t := c.Theme()
	root := ui.Box(c).Fill()
	backgroundBitmap := a.backgroundBitmap
	root.Draw(func(p *ui.Painter, r ui.Rect) {
		if a.BackgroundMode == "Image" && backgroundBitmap != nil {
			p.Image(backgroundBitmap, r, ui.Cover)
			return
		}
		p.Fill(r, t.Background, 0)
	})
	root.Children(func() {
		ui.Column(c).Fill().Padding(16).Gap(12).Children(func() {
			a.toolbar(c, t)
			ui.Row(c).Fill().Gap(12).Children(func() {
				a.sidebar(c, t)
				a.workspace(c, t)
			})
		})
		a.groupDialogs(c, t)
	})
}

func (a *App) toolbar(c *ui.Context, t *ui.Theme) {
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, "NanoAPI").FontSize(20).Bold()
		ui.Text(c, "Native HTTP client").TextColor(t.TextMuted)
		ui.Spacer(c)
		a.requestSummary(c, t)
	})
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		a.methodSelect(c)
		ui.TextArea(c, &a.URL).Placeholder("https://api.example.com/v1/resource").Lines(2, 4).Font("monospace").Grow(1)
		if ui.PrimaryButton(c, "").Label(actionLabel(a.Running)).Tooltip(actionTooltip(a.Running)).Children(func() {
			if a.Running {
				ui.Icon(c, iconCancel)
				ui.Text(c, "Cancel").SingleLine()
			} else {
				ui.Icon(c, iconSend)
				ui.Text(c, "Send").SingleLine()
			}
		}).Clicked() {
			if a.Running {
				a.cancelRequest()
			} else {
				a.send()
			}
		}
	})
	ui.Row(c).FillWidth().Gap(10).Children(func() {
		ui.Text(c, "Timeout").TextColor(t.TextMuted)
		a.timeoutStepper(c)
		ui.Checkbox(c, &a.FollowRedirects, "Follow redirects")
		ui.Checkbox(c, &a.Stream, "Stream response")
		ui.Checkbox(c, &a.Download, "Download to file")
		if a.Download {
			if ui.Button(c, "").Label("Choose destination").Tooltip("Choose where to save the response").Children(func() {
				ui.Icon(c, iconDownload)
				ui.Text(c, "Destination").SingleLine()
			}).Clicked() {
				go a.chooseDownloadPath()
			}
		}
	})
}

func (a *App) sidebar(c *ui.Context, t *ui.Theme) {
	ui.Column(c).Width(280).FillHeight().Gap(8).Padding(12).Background(t.Surface).Radius(10).Children(func() {
		items := a.filteredHistory()
		ui.Row(c).FillWidth().Children(func() {
			ui.Text(c, "History").Bold()
			ui.Spacer(c)
			ui.Textf(c, "%d", len(items)).TextColor(t.TextMuted)
		})
		ui.SearchField(c, &a.HistorySearch).Label("Search request history")
		a.groupSelect(c)
		a.groupTools(c, t)
		a.historyTools(c, t)
		if len(items) == 0 {
			message := "Requests you send will appear here."
			if strings.TrimSpace(a.HistorySearch) != "" {
				message = "No requests match this search."
			}
			ui.Text(c, message).FontSize(12).TextColor(t.TextMuted)
		}
		ui.Scroll(c).Grow(1).Children(func() {
			for i := range items {
				item := items[i]
				key := item.ID
				if key == "" {
					key = fmt.Sprintf("fallback-%d", i)
				}
				button := ui.Button(c.Key(key), "").Label(item.Label).Tooltip(item.Label).FillWidth().Padding(8, 10).Children(func() {
					ui.Row(c).FillWidth().Gap(6).Children(func() {
						ui.Text(c, item.Request.Method).Bold().TextColor(methodColor(item.Request.Method)).SingleLine()
						ui.Text(c, item.Status).TextColor(t.TextMuted).SingleLine()
					})
					ui.Text(c, compactURL(item.Request.URL)).FontSize(11).TextColor(t.TextMuted).SingleLine()
					ui.Textf(c, "%s | %s", formatHistoryTime(item.RequestedAt), formatDuration(item.Duration)).FontSize(10).TextColor(t.TextMuted).SingleLine()
					if item.GroupID != "" {
						if group, ok := a.groupByID(item.GroupID); ok {
							ui.Text(c, group.Name).FontSize(10).TextColor(t.TextMuted).SingleLine()
						}
					}
				})
				if button.Clicked() {
					a.loadHistory(item)
				}
			}
		})
	})
}

func (a *App) workspace(c *ui.Context, t *ui.Theme) {
	ui.Column(c).Grow(1).FillHeight().Gap(10).Children(func() {
		coloredTabs(c, &a.RequestTab, []string{"Params", "Headers", "Body", "Upload", "Appearance", "Options"}, a.TabColor)
		switch a.RequestTab {
		case 0:
			a.keyValueEditor(c, t, "Query parameters", &a.Params)
		case 1:
			a.headersEditor(c, t)
		case 2:
			a.bodyEditor(c, t)
		case 3:
			a.uploadEditor(c, t)
		case 4:
			a.appearanceEditor(c, t)
		case 5:
			a.optionsEditor(c, t)
		}
		ui.Divider(c)
		a.responsePanel(c, t)
	})
}

func (a *App) keyValueEditor(c *ui.Context, t *ui.Theme, title string, items *[]KeyValue) {
	ui.Row(c).FillWidth().Children(func() {
		ui.Text(c, title).Bold()
		ui.Spacer(c)
		if ui.Button(c, "").Label("Add row").Tooltip("Add a key/value row").Children(func() {
			ui.Icon(c, iconPlus)
			ui.Text(c, "Add").SingleLine()
		}).Clicked() {
			*items = append(*items, KeyValue{Enabled: true})
		}
	})
	if len(*items) == 0 {
		*items = append(*items, KeyValue{Enabled: true})
	}
	deleteIndex := -1
	ui.Column(c).FillWidth().Gap(6).Children(func() {
		for i := range *items {
			item := &(*items)[i]
			ui.Row(c.Key(i)).FillWidth().Gap(6).Children(func() {
				ui.Checkbox(c, &item.Enabled, "").Label("Enable row").Tooltip("Enable this row")
				ui.TextInput(c, &item.Key).Placeholder("key").Width(190)
				ui.TextInput(c, &item.Value).Placeholder("value").Grow(1)
				if ui.Button(c, "").Label("Delete row").Tooltip("Delete this row").Children(func() {
					ui.Icon(c, iconTrash)
				}).Clicked() {
					deleteIndex = i
				}
			})
		}
	})
	if deleteIndex >= 0 {
		*items = append((*items)[:deleteIndex], (*items)[deleteIndex+1:]...)
	}
}

func (a *App) bodyEditor(c *ui.Context, t *ui.Theme) {
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, "Body").Bold()
		ui.Select(c, &a.BodyMode, []string{"none", "raw", "form", "multipart"}).Width(130)
		ui.Text(c, bodyModeDescription(a.BodyMode)).TextColor(t.TextMuted)
	})
	switch a.BodyMode {
	case "raw":
		ui.Field(c, "JSON or text", func() {
			ui.TextArea(c, &a.Body).Font("monospace").Grow(1).FillHeight().Placeholder("Request body")
		}).Description("Content-Type is inferred as JSON for objects and arrays.")
	case "form":
		a.keyValueEditor(c, t, "Form fields", &a.FormFields)
	case "multipart":
		a.keyValueEditor(c, t, "Multipart fields", &a.FormFields)
		ui.Text(c, "Use the Upload tab to choose the file sent with this form.").TextColor(t.TextMuted)
	default:
		ui.Text(c, "No request body will be sent.").TextColor(t.TextMuted)
	}
}

func (a *App) uploadEditor(c *ui.Context, t *ui.Theme) {
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, "File upload").Bold()
		ui.Text(c, "Used by multipart and raw file requests.").TextColor(t.TextMuted)
	})
	ui.Form(c, func() {
		ui.Field(c, "File field", func() { ui.TextInput(c, &a.FileField).Placeholder("file") })
		ui.Field(c, "File", func() {
			ui.Row(c).FillWidth().Gap(8).Children(func() {
				ui.Text(c, shortPath(a.FilePath)).Grow(1).SingleLine()
				if ui.Button(c, "").Label("Choose file").Tooltip("Choose a file to upload").Children(func() {
					ui.Icon(c, iconUpload)
					ui.Text(c, "Choose").SingleLine()
				}).Clicked() {
					go a.chooseUploadFile()
				}
			})
		}).Description("The filename and MIME type are sent automatically.")
	})
}

func (a *App) optionsEditor(c *ui.Context, t *ui.Theme) {
	ui.Row(c).FillWidth().Gap(16).Children(func() {
		ui.Checkbox(c, &a.FollowRedirects, "Follow redirects")
		ui.Checkbox(c, &a.Stream, "Stream response chunks")
		ui.Checkbox(c, &a.Download, "Save full response to disk")
	})
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, "Download destination").Bold()
		ui.Text(c, shortPath(a.DownloadPath)).Grow(1).TextColor(t.TextMuted).SingleLine()
		if ui.Button(c, "").Label("Choose destination").Tooltip("Choose the response download path").Children(func() {
			ui.Icon(c, iconDownload)
			ui.Text(c, "Choose").SingleLine()
		}).Clicked() {
			go a.chooseDownloadPath()
		}
	})
}

func (a *App) responsePanel(c *ui.Context, t *ui.Theme) {
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, "Response").Bold()
		if a.Status != "" {
			color := t.TextMuted
			if strings.HasPrefix(a.Status, "2") {
				color = t.Success
			} else if strings.HasPrefix(a.Status, "4") || strings.HasPrefix(a.Status, "5") {
				color = t.Danger
			}
			ui.Text(c, a.Status).Bold().TextColor(color)
		}
		ui.Text(c, a.ResponseMeta).TextColor(t.TextMuted)
		ui.Spacer(c)
		if ui.Button(c, "").Label("Copy response").Tooltip("Copy response body").Disabled(a.Response == nil).Children(func() {
			ui.Icon(c, iconCopy)
			ui.Text(c, "Copy").SingleLine()
		}).Clicked() {
			c.WriteClipboard(a.ResponseText)
			c.Toast("Response copied")
		}
		if ui.Button(c, "").Label("Save response").Tooltip("Save response body to a file").Disabled(a.Response == nil || a.Response.SavedTo != "").Children(func() {
			ui.Icon(c, iconSave)
			ui.Text(c, "Save").SingleLine()
		}).Clicked() {
			go a.saveResponse()
		}
	})
	if a.Error != "" {
		ui.Text(c, a.Error).TextColor(t.Danger).Selectable()
	}
	coloredTabs(c, &a.ResponseTab, []string{"Preview", "Body", "Headers", "Metadata"}, a.TabColor)
	switch a.ResponseTab {
	case 0:
		a.responsePreviewView(c, t)
	case 1:
		ui.TextArea(c, &a.ResponseText).ReadOnly(true).Font("monospace").Grow(1).FillHeight().Placeholder("Send a request to see its response")
	case 2:
		ui.TextArea(c, &a.HeadersText).ReadOnly(true).Font("monospace").Grow(1).FillHeight().Placeholder("Response headers")
	case 3:
		ui.Scroll(c).Grow(1).FillHeight().Children(func() {
			ui.Text(c, a.metadataText()).Font("monospace").Selectable()
		})
	}
}

func (a *App) send() {
	if a.Download && strings.TrimSpace(a.DownloadPath) == "" {
		a.Error = "Choose a download destination before sending."
		return
	}
	a.parseTimeout()
	req := a.snapshotRequest()
	requestStartedAt := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.Running = true
	a.Error = ""
	a.Status = ""
	a.ResponseMeta = ""
	a.ResponseText = ""
	a.HeadersText = ""
	a.Response = nil
	a.ResponseImage = nil
	a.PreviewKind = previewNone
	a.PreviewMessage = ""
	a.lastChunk = time.Time{}

	var pendingMu sync.Mutex
	var pending strings.Builder
	onChunk := func(chunk []byte) {
		if !req.Stream {
			return
		}
		pendingMu.Lock()
		pending.Write(chunk)
		flush := time.Since(a.lastChunk) >= 75*time.Millisecond
		data := ""
		if flush {
			a.lastChunk = time.Now()
			data = pending.String()
			pending.Reset()
		}
		pendingMu.Unlock()
		if flush {
			a.publish(func() { a.appendChunk(data) })
		}
	}

	go func() {
		response, err := httpclient.Do(ctx, req, onChunk)
		respondedAt := time.Now()
		pendingMu.Lock()
		finalChunk := pending.String()
		pending.Reset()
		pendingMu.Unlock()
		a.publish(func() {
			if finalChunk != "" {
				a.appendChunk(finalChunk)
			}
			a.Running = false
			a.cancel = nil
			if err != nil {
				if errors.Is(err, context.Canceled) {
					a.Error = "Request canceled"
				} else {
					a.Error = err.Error()
				}
				a.addHistory(req, "ERROR", requestStartedAt, respondedAt, respondedAt.Sub(requestStartedAt))
				return
			}
			a.Response = response
			a.setResponsePreview(req.URL, response)
			a.Status = fmt.Sprintf("%d %s", response.StatusCode, strings.TrimPrefix(response.Status, fmt.Sprintf("%d ", response.StatusCode)))
			a.ResponseMeta = fmt.Sprintf("%s | %s -> %s | %s", formatBytes(response.BytesRead), formatClock(requestStartedAt), formatClock(respondedAt), formatDuration(respondedAt.Sub(requestStartedAt)))
			if !req.Stream {
				a.ResponseText = response.Preview
			}
			a.HeadersText = headersText(response.Headers)
			a.ResponseTab = 0
			if a.PreviewKind == previewNone {
				a.ResponseTab = 1
			}
			a.addHistory(req, a.Status, requestStartedAt, respondedAt, respondedAt.Sub(requestStartedAt))
		})
	}()
}

func (a *App) chooseUploadFile() {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Title: "Choose a file to upload"})
	a.publish(func() {
		if err != nil {
			a.Error = err.Error()
			return
		}
		if len(paths) > 0 {
			a.FilePath = paths[0]
			a.Error = ""
		}
	})
}
func (a *App) chooseDownloadPath() {
	path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{Title: "Save response"})
	a.publish(func() {
		if err != nil {
			a.Error = err.Error()
			return
		}
		if path != "" {
			a.DownloadPath = path
			a.Error = ""
		}
	})
}

func (a *App) cancelRequest() {
	if a.cancel != nil {
		a.cancel()
	}
}

func (a *App) saveResponse() {
	response := a.Response
	if response == nil {
		return
	}
	if response.Truncated {
		a.publish(func() {
			a.Error = "The response preview is truncated. Enable Download to file for the complete response."
		})
		return
	}
	defaultName := "response.txt"
	if strings.Contains(strings.ToLower(response.ContentType), "json") {
		defaultName = "response.json"
	}
	path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{Title: "Save response", DefaultPath: defaultName})
	if err != nil {
		a.publish(func() { a.Error = err.Error() })
		return
	}
	if path == "" {
		return
	}
	err = os.WriteFile(path, response.Body, 0o600)
	a.publish(func() {
		if err != nil {
			a.Error = err.Error()
			return
		}
		a.Response.SavedTo = path
		a.ResponseMeta = "Saved to " + shortPath(path)
		a.Error = ""
	})
}

func (a *App) publish(fn func()) {
	if a.update != nil {
		a.update(fn)
		return
	}
	fn()
}

func (a *App) appendChunk(chunk string) {
	a.ResponseText += chunk
	if len(a.ResponseText) > 1024*1024 {
		a.ResponseText = a.ResponseText[len(a.ResponseText)-1024*1024:]
		a.ResponseText = "... stream preview truncated ...\n" + a.ResponseText
	}
}

func (a *App) snapshotRequest() httpclient.Request {
	bodyMode := httpclient.BodyMode(a.BodyMode)
	downloadPath := ""
	if a.Download {
		downloadPath = a.DownloadPath
	}
	return httpclient.Request{
		Method:          a.Method,
		URL:             compactURL(a.URL),
		Params:          cloneKV(a.Params),
		Headers:         cloneKV(a.Headers),
		UserAgent:       a.UserAgent,
		BodyMode:        bodyMode,
		Body:            a.Body,
		FormFields:      cloneKV(a.FormFields),
		FileField:       a.FileField,
		FilePath:        a.FilePath,
		Timeout:         time.Duration(a.TimeoutSeconds) * time.Second,
		FollowRedirects: a.FollowRedirects,
		Stream:          a.Stream,
		DownloadPath:    downloadPath,
	}
}

func (a *App) loadHistory(item HistoryItem) {
	reqGroupID := item.GroupID
	req := item.Request
	a.Method = req.Method
	a.URL = req.URL
	a.SelectedGroup = "ungrouped"
	for _, group := range a.Groups {
		if group.ID == reqGroupID {
			a.SelectedGroup = group.ID
			break
		}
	}
	a.Params = cloneKV(req.Params)
	a.Headers = cloneKV(req.Headers)
	a.UserAgent = req.UserAgent
	a.BodyMode = string(req.BodyMode)
	a.Body = req.Body
	a.FormFields = cloneKV(req.FormFields)
	a.FileField = req.FileField
	a.FilePath = req.FilePath
	a.TimeoutSeconds = req.Timeout.Seconds()
	a.TimeoutText = formatTimeout(a.TimeoutSeconds)
	a.FollowRedirects = req.FollowRedirects
	a.Stream = req.Stream
	a.Download = req.DownloadPath != ""
	a.DownloadPath = req.DownloadPath
	a.Error = ""
}

func (a *App) addHistory(req httpclient.Request, status string, requestedAt, respondedAt time.Time, duration time.Duration) {
	groupID := ""
	if a.SelectedGroup != "all" && a.SelectedGroup != "ungrouped" {
		groupID = a.SelectedGroup
	}
	label := req.Method + " " + shortURL(req.URL)
	a.History = append([]HistoryItem{{ID: fmt.Sprintf("req-%d", time.Now().UnixNano()), GroupID: groupID, Label: label, Request: req, Status: status, CreatedAt: time.Now().UTC(), RequestedAt: requestedAt, RespondedAt: respondedAt, Duration: duration}}, a.History...)
	if len(a.History) > 200 {
		a.History = a.History[:200]
	}
}

func (a *App) metadataText() string {
	if a.Response == nil {
		return "No response yet."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Status: %s\n", a.Response.Status)
	fmt.Fprintf(&b, "Content-Type: %s\n", a.Response.ContentType)
	fmt.Fprintf(&b, "Bytes read: %d\n", a.Response.BytesRead)
	fmt.Fprintf(&b, "Requested at: %s\n", formatFullTime(a.Response.StartedAt))
	fmt.Fprintf(&b, "Responded at: %s\n", formatFullTime(a.Response.CompletedAt))
	fmt.Fprintf(&b, "Duration: %s\n", formatDuration(a.Response.Duration))
	fmt.Fprintf(&b, "Saved to: %s\n", a.Response.SavedTo)
	fmt.Fprintf(&b, "Preview truncated: %t\n", a.Response.Truncated)
	return b.String()
}

func cloneKV(items []KeyValue) []KeyValue {
	if items == nil {
		return nil
	}
	return append([]KeyValue(nil), items...)
}

func headersText(headers []KeyValue) string {
	var b strings.Builder
	for _, h := range headers {
		fmt.Fprintf(&b, "%s: %s\n", h.Key, h.Value)
	}
	return b.String()
}

func bodyModeDescription(mode string) string {
	switch mode {
	case "raw":
		return "JSON / text"
	case "form":
		return "application/x-www-form-urlencoded"
	case "multipart":
		return "multipart/form-data"
	default:
		return "empty body"
	}
}

func actionLabel(running bool) string {
	if running {
		return "Cancel request"
	}
	return "Send request"
}

func actionTooltip(running bool) string {
	if running {
		return "Cancel the current request"
	}
	return "Send the request"
}

func compactURL(raw string) string {
	raw = strings.ReplaceAll(raw, "\r", "")
	raw = strings.ReplaceAll(raw, "\n", "")
	return strings.ReplaceAll(raw, "\t", "")
}

func shortURL(raw string) string {
	raw = compactURL(raw)
	raw = strings.TrimPrefix(strings.TrimPrefix(raw, "https://"), "http://")
	if len(raw) > 46 {
		return raw[:43] + "..."
	}
	return raw
}

func shortPath(path string) string {
	if path == "" {
		return "Not selected"
	}
	if len(path) > 54 {
		return "..." + path[len(path)-51:]
	}
	return path
}

func formatHistoryTime(value time.Time) string {
	if value.IsZero() {
		return "unknown"
	}
	return value.Local().Format("2006-01-02 15:04:05")
}

func formatClock(value time.Time) string {
	if value.IsZero() {
		return "unknown"
	}
	return value.Local().Format("15:04:05.000")
}

func formatFullTime(value time.Time) string {
	if value.IsZero() {
		return "unknown"
	}
	return value.Local().Format(time.RFC3339Nano)
}

func formatDuration(value time.Duration) string {
	return value.Round(time.Millisecond).String()
}

func formatBytes(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
