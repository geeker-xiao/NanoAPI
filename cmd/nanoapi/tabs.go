package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"nanoapi/internal/httpclient"
)

// Tab is one open request: its draft, its options and the response it got.
// The app keeps a slice of them and shows the active one on the right.
type Tab struct {
	ID string
	// SavedID is the saved request this tab was filed as, empty while it has
	// never been saved. Saving again updates that entry rather than adding one.
	SavedID string
	Name    string
	// Group is the folder this tab's request belongs to once saved; it copies
	// from and back into the saved request so the sidebar groups it correctly.
	Group string

	Method    string
	URL       string
	Params    []KeyValue
	Headers   []KeyValue
	Auth      Auth
	UserAgent string

	RequestTab  int
	ResponseTab int

	BodyMode string
	Body     string
	// BodyRawType is the content kind of a raw body: json, xml, html,
	// javascript or text.
	BodyRawType string
	// rawKindApplied is the kind whose sample is currently in Body, so a
	// change of kind can swap in the matching sample.
	rawKindApplied string
	FormFields     []KeyValue
	FileField      string
	FilePath       string

	TimeoutSeconds  float64
	TimeoutText     string
	FollowRedirects bool
	Stream          bool
	Download        bool
	DownloadPath    string

	Running      bool
	Error        string
	Status       string
	ResponseMeta string
	ResponseText string
	HeadersText  string
	Response     *httpclient.Response
	PreviewKind  string
	// PreviewMessage explains why a preview is missing.
	PreviewMessage string
	ResponseImage  ui.ImageSource

	cancel    context.CancelFunc
	lastChunk time.Time
}

// newTab is a blank request with the defaults every tab starts from.
func newTab() *Tab {
	return &Tab{
		ID:              fmt.Sprintf("tab-%d", time.Now().UnixNano()),
		Method:          "GET",
		Params:          []KeyValue{{Enabled: true}},
		Headers:         []KeyValue{{Enabled: true, Key: "Accept", Value: "application/json", ValueMode: "Fixed", RandomMode: "UUID v4"}},
		Auth:            Auth{Mode: "None"},
		UserAgent:       "NanoAPI/1.0",
		BodyMode:        "none",
		BodyRawType:     "json",
		rawKindApplied:  "json",
		Body:            rawBodySample("json"),
		FormFields:      []KeyValue{{Enabled: true}},
		FileField:       "file",
		TimeoutSeconds:  30,
		TimeoutText:     "30",
		FollowRedirects: true,
	}
}

// Title is the label shown on the tab: the method and a short URL, or
// whatever the request was named.
func (tab *Tab) Title() string {
	if name := strings.TrimSpace(tab.Name); name != "" {
		return name
	}
	url := strings.TrimSpace(tab.URL)
	if url == "" {
		return "Untitled"
	}
	return strings.TrimSpace(tab.Method) + " " + shortURL(url)
}

// ShortTitle is Title without the method, for chips that already show the
// method on their own line or badge.
func (tab *Tab) ShortTitle() string {
	if name := strings.TrimSpace(tab.Name); name != "" {
		return name
	}
	url := strings.TrimSpace(tab.URL)
	if url == "" {
		return "Untitled"
	}
	return shortURL(url)
}

// tabBarModes are the two layouts the request tabs can use, chosen in the
// appearance settings. Multi-line is the default and wraps.
const (
	tabBarMultiLine  = "Multi-line"
	tabBarSingleLine = "Single line"
)

var tabBarModes = []string{tabBarMultiLine, tabBarSingleLine}

// tab is the request being edited. A missing tab is created on demand, so the
// workspace always has something to show.
func (a *App) tab() *Tab {
	if len(a.Tabs) == 0 {
		a.Tabs = append(a.Tabs, newTab())
	}
	if a.ActiveTab < 0 {
		a.ActiveTab = 0
	}
	if a.ActiveTab >= len(a.Tabs) {
		a.ActiveTab = len(a.Tabs) - 1
	}
	return a.Tabs[a.ActiveTab]
}

func (a *App) openTab() {
	a.Tabs = append(a.Tabs, newTab())
	a.ActiveTab = len(a.Tabs) - 1
	a.queueSaveRequests()
}

// closeTab drops the tab at index, cancelling a request still in flight, and
// keeps one tab open so the workspace never empties out.
func (a *App) closeTab(index int) {
	if index < 0 || index >= len(a.Tabs) {
		return
	}
	tab := a.Tabs[index]
	if tab.Running && tab.cancel != nil {
		tab.cancel()
	}
	a.Tabs = append(a.Tabs[:index], a.Tabs[index+1:]...)
	if len(a.Tabs) == 0 {
		a.Tabs = append(a.Tabs, newTab())
	}
	if a.ActiveTab > index {
		a.ActiveTab--
	}
	if a.ActiveTab >= len(a.Tabs) {
		a.ActiveTab = len(a.Tabs) - 1
	}
	if a.ActiveTab < 0 {
		a.ActiveTab = 0
	}
	a.queueSaveRequests()
}

// duplicateTab copies the request at index into a fresh tab. The copy keeps the
// same method, URL, parameters and options but starts unsaved, so duplicating a
// saved request does not clobber the original entry.
func (a *App) duplicateTab(index int) {
	if index < 0 || index >= len(a.Tabs) {
		return
	}
	src := a.Tabs[index]
	copy := newTab()
	copy.Method = src.Method
	copy.URL = src.URL
	copy.Params = cloneKV(src.Params)
	copy.Headers = cloneKV(src.Headers)
	copy.Auth = src.Auth
	copy.UserAgent = src.UserAgent
	copy.BodyMode = src.BodyMode
	copy.Body = src.Body
	copy.BodyRawType = src.BodyRawType
	copy.rawKindApplied = src.rawKindApplied
	copy.FormFields = cloneKV(src.FormFields)
	copy.FileField = src.FileField
	copy.FilePath = src.FilePath
	copy.TimeoutSeconds = src.TimeoutSeconds
	copy.TimeoutText = src.TimeoutText
	copy.FollowRedirects = src.FollowRedirects
	copy.Stream = src.Stream
	copy.Download = src.Download
	copy.DownloadPath = src.DownloadPath
	copy.Group = src.Group
	copy.Name = src.Name
	a.Tabs = append(a.Tabs, copy)
	a.ActiveTab = len(a.Tabs) - 1
	a.queueSaveRequests()
}

// tabBar lists the open requests on the right pane, with a button to add one.
// Two layouts are available and picked in the appearance settings: multi-line
// wraps the tabs onto as many rows as they need and shows the method above the
// name; single line keeps one compact row that scrolls sideways, with a
// narrow method, the close button inside each tab, and a bare "+" to add one.
func (a *App) tabBar(c *ui.Context, t *ui.Theme) {
	single := a.TabBarDisplay == tabBarSingleLine
	selectIndex := -1
	closeIndex := -1
	duplicateIndex := -1
	resendIndex := -1
	newTab := false

	// Both layouts lay their children out left to right; they differ in what
	// happens once the row is full.
	build := func(children func()) {
		if single {
			ui.ScrollHorizontal(c).FillWidth().Children(func() {
				ui.Row(c).Gap(6).AlignItems(ui.Center).NoWrap().Children(children)
			})
			return
		}
		ui.Row(c).FillWidth().Gap(6).AlignItems(ui.Center).Wrap().Children(children)
	}

	build(func() {
		for i, tab := range a.Tabs {
			active := i == a.ActiveTab
			chip := ui.Row(c.Key(tab.ID)).Gap(4).Padding(3).Radius(t.Radius).AlignItems(ui.Center).Border(1, t.Border)
			if active {
				chip.Background(t.Surface)
			}
			// Right-clicking a tab opens a menu to copy, re-send or close it.
			chip.ContextMenu(func(m *ui.Menu) {
				if m.Item("复制").Chosen() {
					duplicateIndex = i
				}
				if m.Item("重新发起请求").Disabled(tab.Running).Chosen() {
					resendIndex = i
				}
				m.Separator()
				if m.Item("关闭").Chosen() {
					closeIndex = i
				}
			})
			chip.Children(func() {
				a.tabButton(c, tab, i, single, &selectIndex)
				// Single line has no room beside the chip, so the close
				// button moves inside it, next to the label.
				if single {
					a.tabCloseButton(c, t, i, &closeIndex)
				}
			})
			// Multi-line keeps the close button standing next to its chip.
			if !single {
				a.tabCloseButton(c, t, i, &closeIndex)
			}
		}
		newButton := ui.Button(c, "").Label("New request").Tooltip("Open a new request tab").Padding(6, 8).Radius(t.Radius)
		newButton.Children(func() {
			ui.Icon(c, iconPlus).FontSize(14).TextColor(t.TextMuted)
			if !single {
				ui.Text(c, "New").SingleLine()
			}
		})
		if newButton.Clicked() {
			newTab = true
		}
	})

	if selectIndex >= 0 {
		a.ActiveTab = selectIndex
	}
	if duplicateIndex >= 0 {
		a.duplicateTab(duplicateIndex)
	}
	if resendIndex >= 0 {
		// Make the tab active first so its response lands on the tab the
		// user clicked, then send it with the shared request path.
		a.ActiveTab = resendIndex
		a.send()
	}
	if closeIndex >= 0 {
		a.closeTab(closeIndex)
	}
	if newTab {
		a.openTab()
	}
}

// tabButton is the part of a chip that selects its request. Single line lays
// it out sideways with a narrow method, multi-line stacks the two.
func (a *App) tabButton(c *ui.Context, tab *Tab, index int, single bool, selectIndex *int) {
	t := c.Theme()
	button := ui.Button(c, "").Label(tab.Title()).Tooltip(tab.Title()).Padding(6, 8).Radius(t.Radius)
	button.Children(func() {
		if single {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Width(3).Height(t.Rem(0.9)).Background(methodColor(t, tab.Method)).Radius(2)
				ui.Text(c, tab.Method).Bold().FontSize(t.FontSize * 0.72).TextColor(methodColor(t, tab.Method)).NoWrap()
				ui.Text(c, tab.ShortTitle()).NoWrap().MaxWidth(t.Rem(9))
			})
			return
		}
		ui.Text(c, tab.Method).Bold().TextColor(methodColor(t, tab.Method)).SingleLine()
		ui.Text(c, tab.ShortTitle()).MaxWidth(t.Rem(11)).FontSize(11).TextColor(t.TextMuted).SingleLine()
	})
	if button.Clicked() {
		*selectIndex = index
	}
}

// tabCloseButton drops a request from the tab bar.
func (a *App) tabCloseButton(c *ui.Context, t *ui.Theme, index int, closeIndex *int) {
	button := ui.Button(c, "").Label("Close tab").Tooltip("Close this request").Padding(4, 4).Radius(t.Radius)
	button.Children(func() {
		ui.Icon(c, iconCancel).FontSize(12).TextColor(t.TextMuted)
	})
	if button.Clicked() {
		*closeIndex = index
	}
}
