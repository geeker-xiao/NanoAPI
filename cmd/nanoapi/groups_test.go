package main

import (
	"testing"

	"nanoapi/internal/httpclient"
)

func TestFilteredHistorySearchesRequestAndGroup(t *testing.T) {
	app := App{
		Groups: []RequestGroup{{ID: "g1", Name: "Payments", Description: "Billing APIs"}},
		History: []HistoryItem{{
			ID: "r1", GroupID: "g1", Label: "POST charge",
			Request: httpclient.Request{Method: "POST", URL: "https://api.example.test/charge"},
		}},
		SelectedGroup: "g1",
		HistorySearch: "billing",
	}
	if got := app.filteredHistory(); len(got) != 1 {
		t.Fatalf("filtered history = %#v", got)
	}
	app.HistorySearch = "missing"
	if got := app.filteredHistory(); len(got) != 0 {
		t.Fatalf("filtered history = %#v", got)
	}
}

func TestGroupCreateEditAndDeleteMovesRequests(t *testing.T) {
	app := App{SelectedGroup: "all"}
	app.GroupDraftName = "Production"
	app.GroupDraftDescription = "Live APIs"
	app.saveGroupDraft()
	if len(app.Groups) != 1 || app.SelectedGroup != app.Groups[0].ID {
		t.Fatalf("groups = %#v", app.Groups)
	}
	groupID := app.Groups[0].ID
	app.History = []HistoryItem{{ID: "r1", GroupID: groupID}}

	app.GroupDraftName = "Production APIs"
	app.saveGroupDraft()
	if app.Groups[0].Name != "Production APIs" {
		t.Fatalf("group = %#v", app.Groups[0])
	}

	app.GroupDeleteID = groupID
	app.deleteSelectedGroup()
	if len(app.Groups) != 0 || app.History[0].GroupID != "" || app.SelectedGroup != "all" {
		t.Fatalf("after delete: groups=%#v history=%#v selected=%q", app.Groups, app.History, app.SelectedGroup)
	}
}
