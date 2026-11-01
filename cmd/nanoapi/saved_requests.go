package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
)

// SavedRequest is a stored request: everything needed to run it again, plus
// the name it was filed under. Open tabs are snapshots of the same shape, so
// quitting and coming back restores what was on screen.
type SavedRequest struct {
	ID string `json:"id"`
	// TabID only appears in the list of open tabs: it is the tab's own id,
	// kept apart from ID, which is the saved request this tab came from.
	TabID     string     `json:"tabId,omitempty"`
	Name      string     `json:"name"`
	Method    string     `json:"method"`
	URL       string     `json:"url"`
	Params    []KeyValue `json:"params,omitempty"`
	Headers   []KeyValue `json:"headers,omitempty"`
	Auth      Auth       `json:"auth,omitempty"`
	UserAgent string     `json:"userAgent,omitempty"`

	BodyMode    string     `json:"bodyMode,omitempty"`
	Body        string     `json:"body,omitempty"`
	BodyRawType string     `json:"bodyRawType,omitempty"`
	FormFields  []KeyValue `json:"formFields,omitempty"`
	FileField   string     `json:"fileField,omitempty"`
	FilePath    string     `json:"filePath,omitempty"`

	TimeoutSeconds  float64 `json:"timeoutSeconds,omitempty"`
	TimeoutText     string  `json:"timeoutText,omitempty"`
	FollowRedirects bool    `json:"followRedirects,omitempty"`
	Stream          bool    `json:"stream,omitempty"`
	Download        bool    `json:"download,omitempty"`
	DownloadPath    string  `json:"downloadPath,omitempty"`

	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// requestsFile is where saved requests and the restored tabs live.
type requestsFile struct {
	Saved  []SavedRequest `json:"saved,omitempty"`
	Open   []SavedRequest `json:"open,omitempty"`
	Active int            `json:"activeTab"`
}

func requestsFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "nanoapi", "requests.json"), nil
}

// savedRequestFromTab copies a tab into a storable request.
func savedRequestFromTab(tab *Tab) SavedRequest {
	return SavedRequest{
		ID:              tab.SavedID,
		Name:            strings.TrimSpace(tab.Name),
		Method:          tab.Method,
		URL:             tab.URL,
		Params:          cloneKV(tab.Params),
		Headers:         cloneKV(tab.Headers),
		Auth:            tab.Auth,
		UserAgent:       tab.UserAgent,
		BodyMode:        tab.BodyMode,
		Body:            tab.Body,
		BodyRawType:     tab.BodyRawType,
		FormFields:      cloneKV(tab.FormFields),
		FileField:       tab.FileField,
		FilePath:        tab.FilePath,
		TimeoutSeconds:  tab.TimeoutSeconds,
		TimeoutText:     tab.TimeoutText,
		FollowRedirects: tab.FollowRedirects,
		Stream:          tab.Stream,
		Download:        tab.Download,
		DownloadPath:    tab.DownloadPath,
		UpdatedAt:       time.Now(),
	}
}

// applySavedRequest loads a stored request into a tab, keeping the tab's own
// identity and response; it is how both "open saved" and "restore tabs" work.
func applySavedRequest(tab *Tab, saved SavedRequest) {
	tab.SavedID = saved.ID
	tab.Name = saved.Name
	tab.Method = orMethod(saved.Method)
	tab.URL = saved.URL
	tab.Params = cloneKV(saved.Params)
	tab.Headers = cloneKV(saved.Headers)
	tab.Auth = saved.Auth
	tab.UserAgent = saved.UserAgent
	tab.BodyMode = saved.BodyMode
	tab.Body = saved.Body
	tab.BodyRawType = saved.BodyRawType
	tab.rawKindApplied = saved.BodyRawType
	tab.FormFields = cloneKV(saved.FormFields)
	tab.FileField = saved.FileField
	tab.FilePath = saved.FilePath
	tab.TimeoutSeconds = saved.TimeoutSeconds
	tab.TimeoutText = saved.TimeoutText
	tab.FollowRedirects = saved.FollowRedirects
	tab.Stream = saved.Stream
	tab.Download = saved.Download
	tab.DownloadPath = saved.DownloadPath
	normalizeTab(tab)
}

// normalizeTab fills in the pieces a stored request may be missing, so a tab
// loaded from disk behaves like a freshly opened one.
func normalizeTab(tab *Tab) {
	if tab.Method == "" {
		tab.Method = "GET"
	}
	if tab.BodyMode == "" {
		tab.BodyMode = "none"
	}
	if tab.UserAgent == "" {
		tab.UserAgent = "NanoAPI/1.0"
	}
	if tab.BodyRawType == "" {
		tab.BodyRawType = "json"
	}
	if tab.rawKindApplied == "" {
		tab.rawKindApplied = tab.BodyRawType
	}
	if tab.FileField == "" {
		tab.FileField = "file"
	}
	if tab.Auth.Mode == "" {
		tab.Auth.Mode = "None"
	}
	if tab.Auth.KeyIn == "" {
		tab.Auth.KeyIn = "Header"
	}
	if tab.TimeoutSeconds == 0 {
		tab.TimeoutSeconds = 30
	}
	if tab.TimeoutText == "" {
		tab.TimeoutText = "30"
	}
	if len(tab.Params) == 0 {
		tab.Params = []KeyValue{{Enabled: true}}
	}
	if len(tab.FormFields) == 0 {
		tab.FormFields = []KeyValue{{Enabled: true}}
	}
}

func orMethod(method string) string {
	if strings.TrimSpace(method) == "" {
		return "GET"
	}
	return method
}

// newTabFromSaved opens a stored request as a fresh tab.
func newTabFromSaved(saved SavedRequest) *Tab {
	tab := newTab()
	applySavedRequest(tab, saved)
	return tab
}

func tabSnapshot(tab *Tab) SavedRequest {
	snapshot := savedRequestFromTab(tab)
	// ID stays the saved request behind the tab, so reopening keeps the link
	// instead of mistaking the tab itself for a saved entry.
	snapshot.TabID = tab.ID
	return snapshot
}

// loadRequests reads saved requests and the tabs left open last time.
func (a *App) loadRequests() {
	path, err := requestsFilePath()
	if err != nil {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var stored requestsFile
	if err := json.Unmarshal(data, &stored); err != nil {
		return
	}
	a.SavedRequests = stored.Saved
	sortSavedRequests(a.SavedRequests)
	tabs := make([]*Tab, 0, len(stored.Open))
	for _, saved := range stored.Open {
		tab := newTabFromSaved(saved)
		if saved.TabID != "" {
			tab.ID = saved.TabID
		}
		tabs = append(tabs, tab)
	}
	if len(tabs) == 0 {
		return
	}
	a.Tabs = tabs
	if stored.Active >= 0 && stored.Active < len(a.Tabs) {
		a.ActiveTab = stored.Active
		return
	}
	a.ActiveTab = 0
}

// writeRequests stores the saved requests and the open tabs, so both survive
// quitting. It runs on the UI event loop, which is where the tabs are edited.
func (a *App) writeRequests() error {
	path, err := requestsFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	open := make([]SavedRequest, 0, len(a.Tabs))
	for _, tab := range a.Tabs {
		open = append(open, tabSnapshot(tab))
	}
	data, err := json.MarshalIndent(requestsFile{Saved: a.SavedRequests, Open: open, Active: a.ActiveTab}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// queueSaveRequests writes in the background, taking the snapshot first so the
// tabs are read only on the UI event loop.
func (a *App) queueSaveRequests() {
	saved := a.SavedRequests
	open := make([]SavedRequest, 0, len(a.Tabs))
	for _, tab := range a.Tabs {
		open = append(open, tabSnapshot(tab))
	}
	active := a.ActiveTab
	go func() {
		if err := writeRequestsFile(requestsFile{Saved: saved, Open: open, Active: active}); err != nil {
			a.publish(func() { a.Error = "Could not save requests: " + err.Error() })
		}
	}()
}

// writeMu keeps the autosave and a manual Save from writing the file at once.
var writeMu sync.Mutex

func writeRequestsFile(file requestsFile) error {
	path, err := requestsFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	writeMu.Lock()
	defer writeMu.Unlock()
	return os.WriteFile(path, data, 0o600)
}

// savedSection lists the requests kept on this device, each ready to reopen.
func (a *App) savedSection(c *ui.Context, t *ui.Theme) {
	openID := ""
	deleteID := ""
	ui.Collapsible(c, fmt.Sprintf("Saved requests (%d)", len(a.SavedRequests)), &a.SavedSectionOpen, func() {
		if len(a.SavedRequests) == 0 {
			ui.Text(c, "Save a request to keep it here.").FontSize(12).TextColor(t.TextMuted)
			return
		}
		ui.Scroll(c).FillWidth().MaxHeight(t.Rem(11)).Children(func() {
			ui.Column(c).FillWidth().Gap(4).Children(func() {
				for _, saved := range a.SavedRequests {
					ui.Row(c.Key(saved.ID)).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
						item := ui.Button(c, "").Label(saved.Name).Tooltip(saved.Method+" "+shortURL(saved.URL)).Grow(1).Padding(6, 8).Radius(t.Radius).Children(func() {
							ui.Row(c).FillWidth().Gap(6).Children(func() {
								ui.Text(c, saved.Method).Bold().TextColor(methodColor(t, saved.Method)).SingleLine()
								ui.Text(c, saved.Name).Grow(1).SingleLine()
							})
							ui.Text(c, shortURL(saved.URL)).FontSize(11).TextColor(t.TextMuted).SingleLine()
						})
						if item.Clicked() {
							openID = saved.ID
						}
						if ui.Button(c, "").Label("Delete saved request").Tooltip("Delete "+saved.Name).Padding(4, 3).Radius(t.Radius).Children(func() {
							ui.Icon(c, iconTrash).FontSize(12).TextColor(t.TextMuted)
						}).Clicked() {
							deleteID = saved.ID
						}
					})
				}
			})
		})
	})
	if openID != "" {
		a.openSavedRequest(openID)
	}
	if deleteID != "" {
		a.SavedDeleteID = deleteID
		a.SavedDeleteOpen = true
	}
}

// suggestRequestName names the request after what it does, which is what the
// save dialog offers until a name is typed.
func suggestRequestName(tab *Tab) string {
	if name := strings.TrimSpace(tab.Name); name != "" {
		return name
	}
	host := shortURL(tab.URL)
	if host == "" {
		return "Untitled request"
	}
	return tab.Method + " " + host
}

// openSaveDialog prepares the dialog that files the active request under a
// name. A request saved before keeps its entry, so saving again updates it.
func (a *App) openSaveDialog() {
	tab := a.tab()
	a.SaveDraftName = suggestRequestName(tab)
	a.SaveDraftError = ""
	a.SaveDialogOpen = true
}

// saveRequestDraft stores the active request under the drafted name.
func (a *App) saveRequestDraft() {
	name := strings.TrimSpace(a.SaveDraftName)
	if name == "" {
		a.SaveDraftError = "Give the request a name."
		return
	}
	tab := a.tab()
	for _, saved := range a.SavedRequests {
		if saved.ID != tab.SavedID && strings.EqualFold(saved.Name, name) {
			a.SaveDraftError = "Another saved request already uses this name."
			return
		}
	}
	if tab.SavedID == "" {
		tab.SavedID = fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	saved := savedRequestFromTab(tab)
	saved.ID = tab.SavedID
	saved.Name = name
	replaced := false
	for i := range a.SavedRequests {
		if a.SavedRequests[i].ID == saved.ID {
			a.SavedRequests[i] = saved
			replaced = true
			break
		}
	}
	if !replaced {
		a.SavedRequests = append(a.SavedRequests, saved)
	}
	tab.Name = name
	sortSavedRequests(a.SavedRequests)
	a.SaveDraftError = ""
	a.SaveDialogOpen = false
	a.queueSaveRequests()
}

// openSavedRequest brings a saved request back: it reuses the tab already
// showing it, or opens it in a new one.
func (a *App) openSavedRequest(id string) {
	for i, tab := range a.Tabs {
		if tab.SavedID == id {
			a.ActiveTab = i
			return
		}
	}
	saved, ok := a.savedRequestByID(id)
	if !ok {
		return
	}
	a.Tabs = append(a.Tabs, newTabFromSaved(saved))
	a.ActiveTab = len(a.Tabs) - 1
	a.queueSaveRequests()
}

func (a *App) savedRequestByID(id string) (SavedRequest, bool) {
	for _, saved := range a.SavedRequests {
		if saved.ID == id {
			return saved, true
		}
	}
	return SavedRequest{}, false
}

// deleteSavedRequest drops a saved request but leaves tabs showing it alone.
func (a *App) deleteSavedRequest(id string) {
	kept := a.SavedRequests[:0]
	for _, saved := range a.SavedRequests {
		if saved.ID != id {
			kept = append(kept, saved)
		}
	}
	a.SavedRequests = kept
	for _, tab := range a.Tabs {
		if tab.SavedID == id {
			tab.SavedID = ""
		}
	}
	a.queueSaveRequests()
}

// sortSavedRequests keeps the list alphabetical, by name and without case.
func sortSavedRequests(items []SavedRequest) {
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
}

// startRequestAutosave keeps the open tabs on disk while the app runs. The
// work is posted to the UI event loop, because that is where tabs are edited.
func startRequestAutosave(update func(func()), save func()) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			update(save)
		}
	}()
}
