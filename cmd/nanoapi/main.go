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

	// Tabs hold the open requests; ActiveTab is the one shown on the right.
	Tabs      []*Tab
	ActiveTab int
	// ShowAppearance swaps the request pane for the appearance settings,
	// which belong to the whole app rather than to one request.
	ShowAppearance bool

	// SavedRequests are the requests filed under a name on this device.
	SavedRequests []SavedRequest
	// TabBarDisplay picks how the request tabs are laid out.
	TabBarDisplay string

	ThemeMode        string
	FontFamily       string
	FontSize         float64
	BackgroundMode   string
	BackgroundPath   string
	BackgroundColor  ui.Color
	TabColor         ui.Color
	backgroundBitmap *ui.Bitmap
	HTMLPreview      *mygo.WebView

	// Error carries problems outside a single request, such as settings or
	// history import failures. A request's own failure lives on its tab.
	Error string

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

	// The dialog that files the active request under a name.
	SaveDialogOpen   bool
	SaveDraftName    string
	SaveDraftError   string
	SavedSectionOpen bool
	SavedDeleteOpen  bool
	SavedDeleteID    string

	update func(func())
}

func main() {
	first := newTab()
	first.URL = "https://httpbin.org/get"
	app := &App{
		Tabs:             []*Tab{first},
		ActiveTab:        0,
		SelectedGroup:    "all",
		ThemeMode:        "System",
		FontFamily:       "System",
		FontSize:         14,
		BackgroundMode:   "Solid color",
		BackgroundColor:  styleBackground,
		TabColor:         styleAccent,
		TabBarDisplay:    tabBarMultiLine,
		SavedSectionOpen: true,
	}
	app.loadSettings()
	// Restored tabs replace the starter one, so quitting comes back to what
	// was being worked on.
	app.loadRequests()
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
		update := win.Update
		startRequestAutosave(update, app.queueSaveRequests)
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
			a.appHeader(c, t)
			// History on the left, the open request on the right.
			ui.Row(c).Fill().Gap(12).Children(func() {
				a.sidebar(c, t)
				a.workspace(c, t)
			})
		})
		a.saveRequestDialogs(c, t)
		a.groupDialogs(c, t)
	})
}

// saveRequestDialogs holds the dialog that files a request under a name, and
// the confirmation shown before a saved request is dropped.
func (a *App) saveRequestDialogs(c *ui.Context, t *ui.Theme) {
	ui.Modal(c, &a.SaveDialogOpen, func() {
		tab := a.tab()
		ui.Column(c).Width(440).Padding(20).Gap(12).Background(t.Surface).Border(1, t.Border).Radius(12).Children(func() {
			ui.Text(c, "Save request").FontSize(20).Bold()
			ui.Text(c, "Kept on this device with its URL, headers, query parameters and credential.").TextColor(t.TextMuted)
			ui.Field(c, "Name", func() {
				ui.TextInput(c, &a.SaveDraftName).Placeholder("Request name").Grow(1).AutoFocus()
			})
			ui.Textf(c, "%s %s", tab.Method, shortURL(tab.URL)).FontSize(12).TextColor(t.TextMuted).SingleLine()
			if a.SaveDraftError != "" {
				ui.Text(c, a.SaveDraftError).TextColor(t.Danger)
			}
			ui.Row(c).FillWidth().Gap(8).Justify(ui.End).Children(func() {
				if ui.Button(c, "Cancel").Clicked() {
					a.SaveDialogOpen = false
					a.SaveDraftError = ""
				}
				if ui.PrimaryButton(c, "Save").Clicked() {
					a.saveRequestDraft()
				}
			})
		})
	})

	ui.Modal(c, &a.SavedDeleteOpen, func() {
		saved, ok := a.savedRequestByID(a.SavedDeleteID)
		name := saved.Name
		if !ok {
			name = "this request"
		}
		ui.Column(c).Width(400).Padding(20).Gap(12).Background(t.Surface).Border(1, t.Border).Radius(12).Children(func() {
			ui.Text(c, "Delete saved request?").FontSize(20).Bold()
			ui.Textf(c, "%s will no longer be listed. Requests already open stay open.", name).TextColor(t.TextMuted)
			ui.Row(c).FillWidth().Gap(8).Justify(ui.End).Children(func() {
				if ui.Button(c, "Cancel").Clicked() {
					a.SavedDeleteOpen = false
					a.SavedDeleteID = ""
				}
				if ui.Button(c, "Delete").TextColor(t.Danger).Clicked() {
					a.deleteSavedRequest(a.SavedDeleteID)
					a.SavedDeleteOpen = false
					a.SavedDeleteID = ""
				}
			})
		})
	})
}

// saveTooltip explains what Save does, and when it will update an entry.
func saveTooltip(saved bool) string {
	if saved {
		return "Update the saved copy of this request, or rename it"
	}
	return "File this request under a name"
}

// appHeader is the app-wide bar: it holds nothing that belongs to one request,
// so switching tabs only ever changes the pane on the right.
func (a *App) appHeader(c *ui.Context, t *ui.Theme) {
	ui.Row(c).FillWidth().Gap(t.Space(2)).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "NanoAPI").FontSize(20).Bold()
		ui.Text(c, "Native HTTP client").TextColor(t.TextMuted)
		ui.Spacer(c)
		if a.Error != "" {
			ui.Text(c, a.Error).TextColor(t.Danger).SingleLine()
		}
		if ui.Button(c, "").Label("Appearance").Tooltip("Edit how the app looks").Children(func() {
			ui.Text(c, "Appearance").SingleLine()
		}).Clicked() {
			a.ShowAppearance = !a.ShowAppearance
		}
	})
}

// requestBar is the part of a request that sits above its detail tabs: the
// method, the URL, the send button and the options for this request only.
func (a *App) requestBar(c *ui.Context, t *ui.Theme) {
	tab := a.tab()
	// Stretch keeps the send box as tall as the method and URL box beside it.
	ui.Row(c).FillWidth().Gap(t.Space(2)).AlignItems(ui.Stretch).Children(func() {
		// The method and the URL live in one box, so they read as one field.
		ui.Row(c).Grow(1).Gap(t.Space(1)).Padding(t.Space(1)).Background(t.Surface).Border(1, t.Border).Radius(t.Radius).AlignItems(ui.Start).Children(func() {
			a.methodSelect(c)
			ui.Box(c).Width(1).Height(t.Rem(1.8)).Background(t.Border)
			a.urlField(c, t)
		})
		if ui.PrimaryButton(c, "").Label(actionLabel(tab.Running)).Tooltip(actionTooltip(tab.Running)).MinWidth(t.Rem(6.5)).Padding(t.Space(2), t.Space(3)).Radius(t.Radius).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			if tab.Running {
				ui.Icon(c, iconCancel)
				ui.Text(c, "Cancel").SingleLine()
			} else {
				ui.Icon(c, iconSend)
				ui.Text(c, "Send").SingleLine()
			}
		}).Clicked() {
			if tab.Running {
				a.cancelRequest()
			} else {
				a.send()
			}
		}
		// Save files the request under a name so it can be run again later.
		if ui.Button(c, "").Label("Save request").Tooltip(saveTooltip(tab.SavedID != "")).Padding(t.Space(2), t.Space(2.5)).Radius(t.Radius).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			ui.Icon(c, iconSave)
			ui.Text(c, "Save").SingleLine()
		}).Clicked() {
			a.openSaveDialog()
		}
	})
	ui.Row(c).FillWidth().Gap(t.Space(2.5)).Children(func() {
		ui.Text(c, "Timeout").TextColor(t.TextMuted)
		a.timeoutStepper(c)
		ui.Checkbox(c, &tab.FollowRedirects, "Follow redirects")
		ui.Checkbox(c, &tab.Stream, "Stream response")
		ui.Checkbox(c, &tab.Download, "Download to file")
		if tab.Download {
			if ui.Button(c, "").Label("Choose destination").Tooltip("Choose where to save the response").Children(func() {
				ui.Icon(c, iconDownload)
				ui.Text(c, "Destination").SingleLine()
			}).Clicked() {
				go a.chooseDownloadPath()
			}
		}
	})
}

// urlField is the URL half of the combined box. It is a single-line input, so
// the method, URL and Send bar always stay on one row and the URL never spreads
// across multiple lines, however long it gets.
func (a *App) urlField(c *ui.Context, t *ui.Theme) {
	tab := a.tab()
	// Borderless: the surrounding box already draws the frame, and a second
	// one would cut the URL off from the method beside it.
	ui.TextInput(c, &tab.URL).Placeholder("https://api.example.com/v1/resource").Font("monospace").Grow(1).Border(0, t.Border)
}

func (a *App) sidebar(c *ui.Context, t *ui.Theme) {
	ui.Column(c).Width(280).FillHeight().Gap(8).Padding(12).Background(t.Surface).Radius(14).Children(func() {
		a.savedSection(c, t)
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
						ui.Text(c, item.Request.Method).Bold().TextColor(methodColor(t, item.Request.Method)).SingleLine()
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

// requestTabs are the detail sections of one request. Appearance is not among
// them: it belongs to the whole app and lives behind the header button.
var requestTabs = []string{"Params", "Headers", "Auth", "Body", "Upload", "Options"}

func (a *App) workspace(c *ui.Context, t *ui.Theme) {
	tab := a.tab()
	ui.Column(c).Grow(1).FillHeight().Gap(10).Children(func() {
		a.tabBar(c, t)
		if a.ShowAppearance {
			if ui.Button(c, "").Label("Back to the request").Tooltip("Close the appearance settings").Children(func() {
				ui.Text(c, "Back to the request").SingleLine()
			}).Clicked() {
				a.ShowAppearance = false
			}
			a.appearanceEditor(c, t)
			return
		}
		a.requestBar(c, t)
		ui.Divider(c)
		coloredTabs(c, &tab.RequestTab, requestTabs, a.TabColor)
		switch tab.RequestTab {
		case 0:
			a.keyValueEditor(c, t, "Query parameters", &tab.Params)
		case 1:
			a.headersEditor(c, t)
		case 2:
			a.authEditor(c, t)
		case 3:
			a.bodyEditor(c, t)
		case 4:
			a.uploadEditor(c, t)
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

// bodyModes are the modes offered for a request body, in the order shown.
var bodyModes = []string{"none", "raw", "form", bodyModeFormData}

// bodyModeFormData is the current name of the multipart mode; "multipart" is
// still accepted when reading older requests.
const bodyModeFormData = "form-data"

// formDataEditor lists form-data rows, each of which carries text or a file.
func (a *App) formDataEditor(c *ui.Context, t *ui.Theme, title string, items *[]KeyValue) {
	ui.Row(c).FillWidth().Children(func() {
		ui.Text(c, title).Bold()
		ui.Spacer(c)
		if ui.Button(c, "").Label("Add row").Tooltip("Add a form-data row").Children(func() {
			ui.Icon(c, iconPlus)
			ui.Text(c, "Add").SingleLine()
		}).Clicked() {
			*items = append(*items, KeyValue{Enabled: true, FieldType: "text"})
		}
	})
	if len(*items) == 0 {
		*items = append(*items, KeyValue{Enabled: true, FieldType: "text"})
	}
	deleteIndex := -1
	ui.Column(c).FillWidth().Gap(6).Children(func() {
		for i := range *items {
			item := &(*items)[i]
			if strings.TrimSpace(item.FieldType) == "" {
				item.FieldType = "text"
			}
			isFile := strings.EqualFold(item.FieldType, "file")
			ui.Row(c.Key(i)).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Checkbox(c, &item.Enabled, "").Label("Enable row").Tooltip("Enable this row")
				ui.TextInput(c, &item.Key).Placeholder("key").Width(190)
				ui.Select(c, &item.FieldType, httpclient.FieldTypes).Width(90)
				if isFile {
					label := shortPath(item.Value)
					if strings.TrimSpace(label) == "" {
						label = "No file chosen"
					}
					ui.Text(c, label).Grow(1).SingleLine().TextColor(t.TextMuted)
					if ui.Button(c, "").Label("Choose file").Tooltip("Choose the file sent as this field").Children(func() {
						ui.Icon(c, iconUpload)
						ui.Text(c, "Choose").SingleLine()
					}).Clicked() {
						go a.chooseFormFile(i)
					}
				} else {
					ui.TextInput(c, &item.Value).Placeholder("value").Grow(1)
				}
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
	tab := a.tab()
	ui.Row(c).FillWidth().Gap(t.Space(2)).Children(func() {
		ui.Text(c, "Body").Bold()
		ui.Select(c, &tab.BodyMode, bodyModes).Width(150)
		// The kind of a raw body, right after the mode that turns it on.
		if tab.BodyMode == "raw" {
			if tab.BodyRawType == "" {
				tab.BodyRawType = "json"
			}
			ui.Select(c, &tab.BodyRawType, httpclient.RawBodyTypes).Width(120)
		}
		if description := bodyModeDescription(tab.BodyMode); description != "" {
			ui.Text(c, description).TextColor(t.TextMuted)
		}
	})
	switch tab.BodyMode {
	case "raw":
		// Keep the sample in the editor in step with the picked kind.
		tab.syncRawBodySample()
		ui.Field(c, rawBodyLabel(tab.BodyRawType), func() {
			ui.TextArea(c, &tab.Body).Font("monospace").Grow(1).FillHeight().Placeholder("Request body")
		}).Description("Sent as " + httpclient.RawContentType(tab.BodyRawType) + ".")
	case "form":
		a.keyValueEditor(c, t, "Form fields", &tab.FormFields)
	case bodyModeFormData, "multipart":
		a.formDataEditor(c, t, "Form-data fields", &tab.FormFields)
	default:
		ui.Text(c, "No request body will be sent.").TextColor(t.TextMuted)
	}
}

func (a *App) uploadEditor(c *ui.Context, t *ui.Theme) {
	tab := a.tab()
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, "File upload").Bold()
		ui.Text(c, "Also sent with form-data requests saved before rows had a kind.").TextColor(t.TextMuted)
	})
	ui.Form(c, func() {
		ui.Field(c, "File field", func() { ui.TextInput(c, &tab.FileField).Placeholder("file") })
		ui.Field(c, "File", func() {
			ui.Row(c).FillWidth().Gap(8).Children(func() {
				ui.Text(c, shortPath(tab.FilePath)).Grow(1).SingleLine()
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
	tab := a.tab()
	ui.Row(c).FillWidth().Gap(16).Children(func() {
		ui.Checkbox(c, &tab.FollowRedirects, "Follow redirects")
		ui.Checkbox(c, &tab.Stream, "Stream response chunks")
		ui.Checkbox(c, &tab.Download, "Save full response to disk")
	})
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, "Download destination").Bold()
		ui.Text(c, shortPath(tab.DownloadPath)).Grow(1).TextColor(t.TextMuted).SingleLine()
		if ui.Button(c, "").Label("Choose destination").Tooltip("Choose the response download path").Children(func() {
			ui.Icon(c, iconDownload)
			ui.Text(c, "Choose").SingleLine()
		}).Clicked() {
			go a.chooseDownloadPath()
		}
	})
}

func (a *App) responsePanel(c *ui.Context, t *ui.Theme) {
	tab := a.tab()
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, "Response").Bold()
		if tab.Status != "" {
			color := t.TextMuted
			if strings.HasPrefix(tab.Status, "2") {
				color = t.Success
			} else if strings.HasPrefix(tab.Status, "4") || strings.HasPrefix(tab.Status, "5") {
				color = t.Danger
			}
			ui.Text(c, tab.Status).Bold().TextColor(color)
		}
		ui.Text(c, tab.ResponseMeta).TextColor(t.TextMuted)
		ui.Spacer(c)
		if ui.Button(c, "").Label("Copy response").Tooltip("Copy response body").Disabled(tab.Response == nil).Children(func() {
			ui.Icon(c, iconCopy)
			ui.Text(c, "Copy").SingleLine()
		}).Clicked() {
			c.WriteClipboard(tab.ResponseText)
			c.Toast("Response copied")
		}
		if ui.Button(c, "").Label("Save response").Tooltip("Save response body to a file").Disabled(tab.Response == nil || tab.Response.SavedTo != "").Children(func() {
			ui.Icon(c, iconSave)
			ui.Text(c, "Save").SingleLine()
		}).Clicked() {
			go a.saveResponse()
		}
	})
	if tab.Error != "" {
		ui.Text(c, tab.Error).TextColor(t.Danger).Selectable()
	}
	coloredTabs(c, &tab.ResponseTab, []string{"Preview", "Body", "Headers", "Metadata"}, a.TabColor)
	switch tab.ResponseTab {
	case 0:
		a.responsePreviewView(c, t)
	case 1:
		ui.TextArea(c, &tab.ResponseText).ReadOnly(true).Font("monospace").Grow(1).FillHeight().Placeholder("Send a request to see its response")
	case 2:
		ui.TextArea(c, &tab.HeadersText).ReadOnly(true).Font("monospace").Grow(1).FillHeight().Placeholder("Response headers")
	case 3:
		ui.Scroll(c).Grow(1).FillHeight().Children(func() {
			ui.Text(c, a.metadataText()).Font("monospace").Selectable()
		})
	}
}

func (a *App) send() {
	tab := a.tab()
	if tab.Download && strings.TrimSpace(tab.DownloadPath) == "" {
		tab.Error = "Choose a download destination before sending."
		return
	}
	a.parseTimeout()
	req := a.snapshotRequest()
	requestStartedAt := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	tab.cancel = cancel
	tab.Running = true
	tab.Error = ""
	tab.Status = ""
	tab.ResponseMeta = ""
	tab.ResponseText = ""
	tab.HeadersText = ""
	tab.Response = nil
	tab.ResponseImage = nil
	tab.PreviewKind = previewNone
	tab.PreviewMessage = ""
	tab.lastChunk = time.Time{}

	var pendingMu sync.Mutex
	var pending strings.Builder
	onChunk := func(chunk []byte) {
		if !req.Stream {
			return
		}
		pendingMu.Lock()
		pending.Write(chunk)
		flush := time.Since(tab.lastChunk) >= 75*time.Millisecond
		data := ""
		if flush {
			tab.lastChunk = time.Now()
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
			tab.Running = false
			tab.cancel = nil
			if err != nil {
				if errors.Is(err, context.Canceled) {
					tab.Error = "Request canceled"
				} else {
					tab.Error = err.Error()
				}
				a.addHistory(req, "ERROR", requestStartedAt, respondedAt, respondedAt.Sub(requestStartedAt))
				return
			}
			tab.Response = response
			a.setResponsePreview(req.URL, response)
			tab.Status = fmt.Sprintf("%d %s", response.StatusCode, strings.TrimPrefix(response.Status, fmt.Sprintf("%d ", response.StatusCode)))
			tab.ResponseMeta = fmt.Sprintf("%s | %s -> %s | %s", formatBytes(response.BytesRead), formatClock(requestStartedAt), formatClock(respondedAt), formatDuration(respondedAt.Sub(requestStartedAt)))
			if !req.Stream {
				tab.ResponseText = response.Preview
			}
			tab.HeadersText = headersText(response.Headers)
			tab.ResponseTab = 0
			if tab.PreviewKind == previewNone {
				tab.ResponseTab = 1
			}
			a.addHistory(req, tab.Status, requestStartedAt, respondedAt, respondedAt.Sub(requestStartedAt))
		})
	}()
}

func (a *App) chooseUploadFile() {
	tab := a.tab()
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Title: "Choose a file to upload"})
	a.publish(func() {
		if err != nil {
			tab.Error = err.Error()
			return
		}
		if len(paths) > 0 {
			tab.FilePath = paths[0]
			tab.Error = ""
		}
	})
}

// chooseFormFile picks the file sent by the form-data row at index.
func (a *App) chooseFormFile(index int) {
	tab := a.tab()
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Title: "Choose a file for this field"})
	a.publish(func() {
		if err != nil {
			tab.Error = err.Error()
			return
		}
		if len(paths) > 0 && index >= 0 && index < len(tab.FormFields) {
			tab.FormFields[index].Value = paths[0]
			tab.Error = ""
		}
	})
}

func (a *App) chooseDownloadPath() {
	tab := a.tab()
	path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{Title: "Save response"})
	a.publish(func() {
		if err != nil {
			tab.Error = err.Error()
			return
		}
		if path != "" {
			tab.DownloadPath = path
			tab.Error = ""
		}
	})
}

func (a *App) cancelRequest() {
	tab := a.tab()
	if tab.cancel != nil {
		tab.cancel()
	}
}

func (a *App) saveResponse() {
	tab := a.tab()
	response := tab.Response
	if response == nil {
		return
	}
	if response.Truncated {
		a.publish(func() {
			tab.Error = "The response preview is truncated. Enable Download to file for the complete response."
		})
		return
	}
	defaultName := "response.txt"
	if strings.Contains(strings.ToLower(response.ContentType), "json") {
		defaultName = "response.json"
	}
	path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{Title: "Save response", DefaultPath: defaultName})
	if err != nil {
		a.publish(func() { tab.Error = err.Error() })
		return
	}
	if path == "" {
		return
	}
	err = os.WriteFile(path, response.Body, 0o600)
	a.publish(func() {
		if err != nil {
			tab.Error = err.Error()
			return
		}
		tab.Response.SavedTo = path
		tab.ResponseMeta = "Saved to " + shortPath(path)
		tab.Error = ""
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
	tab := a.tab()
	tab.ResponseText += chunk
	if len(tab.ResponseText) > 1024*1024 {
		tab.ResponseText = tab.ResponseText[len(tab.ResponseText)-1024*1024:]
		tab.ResponseText = "... stream preview truncated ...\n" + tab.ResponseText
	}
}

func (a *App) snapshotRequest() httpclient.Request {
	tab := a.tab()
	bodyMode := httpclient.NormalizeBodyMode(httpclient.BodyMode(tab.BodyMode))
	downloadPath := ""
	if tab.Download {
		downloadPath = tab.DownloadPath
	}
	params, headers := withAuth(tab)
	return httpclient.Request{
		Method:          tab.Method,
		URL:             compactURL(tab.URL),
		Params:          params,
		Headers:         headers,
		UserAgent:       tab.UserAgent,
		BodyMode:        bodyMode,
		Body:            tab.Body,
		BodyRawType:     tab.BodyRawType,
		FormFields:      cloneKV(tab.FormFields),
		FileField:       tab.FileField,
		FilePath:        tab.FilePath,
		Timeout:         time.Duration(tab.TimeoutSeconds) * time.Second,
		FollowRedirects: tab.FollowRedirects,
		Stream:          tab.Stream,
		DownloadPath:    downloadPath,
	}
}

func (a *App) loadHistory(item HistoryItem) {
	tab := a.tab()
	reqGroupID := item.GroupID
	req := item.Request
	tab.Method = req.Method
	tab.URL = req.URL
	a.SelectedGroup = "ungrouped"
	for _, group := range a.Groups {
		if group.ID == reqGroupID {
			a.SelectedGroup = group.ID
			break
		}
	}
	tab.Params = cloneKV(req.Params)
	tab.Headers = cloneKV(req.Headers)
	tab.UserAgent = req.UserAgent
	tab.BodyMode = string(httpclient.NormalizeBodyMode(req.BodyMode))
	tab.Body = req.Body
	tab.BodyRawType = req.BodyRawType
	if tab.BodyRawType == "" {
		tab.BodyRawType = "json"
	}
	// The history body is what the user sent, not a sample to swap out.
	tab.rawKindApplied = tab.BodyRawType
	tab.FormFields = cloneKV(req.FormFields)
	tab.FileField = req.FileField
	tab.FilePath = req.FilePath
	tab.TimeoutSeconds = req.Timeout.Seconds()
	tab.TimeoutText = formatTimeout(tab.TimeoutSeconds)
	tab.FollowRedirects = req.FollowRedirects
	tab.Stream = req.Stream
	tab.Download = req.DownloadPath != ""
	tab.DownloadPath = req.DownloadPath
	tab.Error = ""
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
	tab := a.tab()
	if tab.Response == nil {
		return "No response yet."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Status: %s\n", tab.Response.Status)
	fmt.Fprintf(&b, "Content-Type: %s\n", tab.Response.ContentType)
	fmt.Fprintf(&b, "Bytes read: %d\n", tab.Response.BytesRead)
	fmt.Fprintf(&b, "Requested at: %s\n", formatFullTime(tab.Response.StartedAt))
	fmt.Fprintf(&b, "Responded at: %s\n", formatFullTime(tab.Response.CompletedAt))
	fmt.Fprintf(&b, "Duration: %s\n", formatDuration(tab.Response.Duration))
	fmt.Fprintf(&b, "Saved to: %s\n", tab.Response.SavedTo)
	fmt.Fprintf(&b, "Preview truncated: %t\n", tab.Response.Truncated)
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
		// The dropdown beside the mode already names the kind.
		return ""
	case "form":
		return "application/x-www-form-urlencoded"
	case bodyModeFormData, "multipart":
		return "multipart/form-data"
	default:
		return "empty body"
	}
}

// rawBodySamples holds one starter snippet per raw body kind. Switching the
// kind swaps the snippet in, so the editor always shows something valid for
// the Content-Type that is about to be sent.
var rawBodySamples = map[string]string{
	"json":       "{\n  \"message\": \"Hello from NanoAPI\"\n}",
	"xml":        "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<request>\n  <message>Hello from NanoAPI</message>\n</request>",
	"html":       "<!DOCTYPE html>\n<html>\n  <body>\n    <h1>Hello from NanoAPI</h1>\n  </body>\n</html>",
	"javascript": "const message = \"Hello from NanoAPI\";\n\nconsole.log(message);",
	"text":       "Hello from NanoAPI",
}

// rawBodySample is the starter snippet for kind, falling back to JSON.
func rawBodySample(kind string) string {
	if sample, ok := rawBodySamples[strings.ToLower(strings.TrimSpace(kind))]; ok {
		return sample
	}
	return rawBodySamples["json"]
}

// isUntouchedRawBody reports whether body still holds one of the starter
// snippets, or nothing at all. Only such a body may be swapped out.
func isUntouchedRawBody(body string) bool {
	if strings.TrimSpace(body) == "" {
		return true
	}
	for _, sample := range rawBodySamples {
		if body == sample {
			return true
		}
	}
	return false
}

// syncRawBodySample swaps the starter snippet whenever the picked kind changes.
// A body the user has edited is left alone, so switching kinds never throws
// away real work.
func (tab *Tab) syncRawBodySample() {
	kind := strings.ToLower(strings.TrimSpace(tab.BodyRawType))
	if kind == "" {
		kind = "json"
		tab.BodyRawType = kind
	}
	// Only a change of kind swaps the snippet; a stable kind leaves the body
	// as it is, so clearing the editor keeps it cleared.
	if tab.rawKindApplied == kind {
		return
	}
	tab.rawKindApplied = kind
	if isUntouchedRawBody(tab.Body) {
		tab.Body = rawBodySample(kind)
	}
}

// rawBodyLabel names the body field after the kind picked in the dropdown.
func rawBodyLabel(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "json":
		return "JSON body"
	case "xml":
		return "XML body"
	case "html":
		return "HTML body"
	case "javascript":
		return "JavaScript body"
	case "text":
		return "Text body"
	}
	return "Body"
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
