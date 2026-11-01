package main

import (
	"strings"
	"testing"
)

func appWithTabs(count int) *App {
	app := &App{TabBarDisplay: tabBarMultiLine, SavedSectionOpen: true}
	for i := 0; i < count; i++ {
		app.Tabs = append(app.Tabs, newTab())
	}
	app.ActiveTab = 0
	return app
}

func TestSaveRequestDraftStoresEverythingUnderTheName(t *testing.T) {
	app := appWithTabs(1)
	tab := aNanoTab(app)
	tab.Method = "POST"
	tab.URL = "https://api.example.com/v1/things"
	tab.Params = []KeyValue{{Enabled: true, Key: "page", Value: "3"}}
	tab.Headers = []KeyValue{{Enabled: true, Key: "Accept", Value: "application/json"}}
	tab.Auth = Auth{Mode: "Bearer token", Token: "tok"}
	tab.UserAgent = "NanoAPI/1.0"

	app.openSaveDialog()
	if aNanoSaveDraft(app) != suggestRequestName(tab) {
		t.Fatal("the dialog should offer a name built from the request")
	}
	app.SaveDraftName = "Create thing"
	app.saveRequestDraft()

	if len(app.SavedRequests) != 1 {
		t.Fatalf("expected one saved request, got %d", len(app.SavedRequests))
	}
	saved := app.SavedRequests[0]
	if saved.Name != "Create thing" {
		t.Fatalf("unexpected name %q", saved.Name)
	}
	if saved.URL != tab.URL || saved.Method != "POST" {
		t.Fatalf("request identity not stored: %+v", saved)
	}
	if len(saved.Params) != 1 || len(saved.Headers) != 1 {
		t.Fatalf("params or headers not stored: %+v", saved)
	}
	if saved.Auth.Token != "tok" {
		t.Fatalf("credential not stored: %+v", saved.Auth)
	}
	if strings.TrimSpace(tab.Name) != "Create thing" {
		t.Fatalf("the tab should carry the name too, got %q", tab.Name)
	}
	if tab.SavedID == "" {
		t.Fatal("the tab should be linked to the saved entry")
	}
}

func TestSaveRequestDraftNeedsAName(t *testing.T) {
	app := appWithTabs(1)
	app.openSaveDialog()
	app.SaveDraftName = "   "
	app.saveRequestDraft()
	if app.SaveDraftError == "" {
		t.Fatal("an empty name should be refused")
	}
	if len(app.SavedRequests) != 0 {
		t.Fatal("nothing should be saved without a name")
	}
}

func TestSaveRequestDraftRefusesADuplicateName(t *testing.T) {
	app := appWithTabs(2)
	aNanoTab(app).URL = "https://one.example.com"
	app.openSaveDialog()
	app.SaveDraftName = "First"
	app.saveRequestDraft()

	app.ActiveTab = 1
	app.openSaveDialog()
	app.SaveDraftName = "first"
	app.saveRequestDraft()
	if app.SaveDraftError == "" {
		t.Fatal("a duplicate name should be refused")
	}
	if len(app.SavedRequests) != 1 {
		t.Fatalf("expected one saved request, got %d", len(app.SavedRequests))
	}
}

func TestSavingAgainUpdatesTheSameEntry(t *testing.T) {
	app := appWithTabs(1)
	tab := aNanoTab(app)
	app.openSaveDialog()
	app.SaveDraftName = "One"
	app.saveRequestDraft()
	id := tab.SavedID

	tab.URL = "https://api.example.com/changed"
	app.openSaveDialog()
	app.SaveDraftName = "Two"
	app.saveRequestDraft()

	if len(app.SavedRequests) != 1 {
		t.Fatalf("expected the entry to be updated, got %d", len(app.SavedRequests))
	}
	if app.SavedRequests[0].ID != id || app.SavedRequests[0].Name != "Two" {
		t.Fatalf("unexpected entry: %+v", app.SavedRequests[0])
	}
	if app.SavedRequests[0].URL != "https://api.example.com/changed" {
		t.Fatal("the stored request should follow the tab")
	}
}

func TestOpenSavedRequestReusesTheTabShowingIt(t *testing.T) {
	app := appWithTabs(1)
	app.openSaveDialog()
	app.SaveDraftName = "Kept"
	app.saveRequestDraft()
	id := app.SavedRequests[0].ID

	app.ActiveTab = 0
	app.openSavedRequest(id)
	if len(app.Tabs) != 1 || app.ActiveTab != 0 {
		t.Fatalf("expected the open tab to be reused, got %d tabs", len(app.Tabs))
	}
	if aNanoTab(app).SavedID != id {
		t.Fatal("the reused tab lost its link")
	}
}

func TestOpenSavedRequestOpensANewTabWhenClosed(t *testing.T) {
	app := appWithTabs(1)
	app.openSaveDialog()
	app.SaveDraftName = "Stored"
	app.saveRequestDraft()
	saved := app.SavedRequests[0]
	savedURL := saved.URL

	before := len(app.Tabs)
	app.closeTab(0)
	// Closing the last tab leaves a blank one behind, so the saved request
	// has nowhere to be restored to and must open its own tab.
	app.openSavedRequest(saved.ID)

	if got := len(app.Tabs); got != before+1 {
		t.Fatalf("expected a tab to open, went from %d to %d", before, got)
	}
	if app.ActiveTab != len(app.Tabs)-1 {
		t.Fatal("the new tab should become active")
	}
	if aNanoTab(app).Name != "Stored" {
		t.Fatalf("the restored tab lost its name, got %q", aNanoTab(app).Name)
	}
	if aNanoTab(app).URL != savedURL {
		t.Fatalf("the restored tab lost its URL, got %q", aNanoTab(app).URL)
	}
}

func TestDeleteSavedRequestLeavesOpenTabsAlone(t *testing.T) {
	app := appWithTabs(1)
	app.openSaveDialog()
	app.SaveDraftName = "Gone"
	app.saveRequestDraft()
	id := app.SavedRequests[0].ID

	app.deleteSavedRequest(id)
	if len(app.SavedRequests) != 0 {
		t.Fatalf("expected the entry to go, got %d", len(app.SavedRequests))
	}
	if len(app.Tabs) != 1 {
		t.Fatal("closing nothing: the tab should stay open")
	}
	if aNanoTab(app).SavedID != "" {
		t.Fatal("the tab should be free to be filed again")
	}
}

func TestSuggestRequestNameFallsBackToTheRequest(t *testing.T) {
	tab := newTab()
	tab.URL = "https://api.example.com/v1/items?x=1"
	if got := suggestRequestName(tab); !strings.EqualFold(got, "GET "+shortURL(tab.URL)) {
		t.Fatalf("unexpected suggestion %q", got)
	}
	tab.Name = "Named"
	if got := suggestRequestName(tab); got != "Named" {
		t.Fatalf("the given name should win, got %q", got)
	}
	empty := newTab()
	empty.URL = ""
	if got := suggestRequestName(empty); !strings.Contains(got, "Untitled") {
		t.Fatalf("unexpected suggestion %q", got)
	}
}

func TestTabTitles(t *testing.T) {
	tab := newTab()
	tab.URL = "https://api.example.com/v1/x"
	if tab.Title() != "GET "+shortURL(tab.URL) {
		t.Fatalf("unexpected title %q", tab.Title())
	}
	if tab.ShortTitle() != shortURL(tab.URL) {
		t.Fatalf("unexpected short title %q", tab.ShortTitle())
	}
	tab.Name = "Named"
	if tab.Title() != "Named" || tab.ShortTitle() != "Named" {
		t.Fatal("a name should replace the generated titles")
	}
}

func TestTabBarModesOfferBothLayouts(t *testing.T) {
	if len(tabBarModes) != 2 || tabBarModes[0] != tabBarMultiLine || tabBarModes[1] != tabBarSingleLine {
		t.Fatalf("unexpected modes %v", tabBarModes)
	}
}

// aNanoTab and aNanoSaveDraft read the private state the way the app does,
// keeping the tests free of the app's plumbing.
func aNanoTab(app *App) *Tab { return app.tab() }

func aNanoSaveDraft(app *App) string { return app.SaveDraftName }
