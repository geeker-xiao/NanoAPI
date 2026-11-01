package main

import "testing"

func TestTabOpensAndCloses(t *testing.T) {
	a := &App{}
	first := a.tab()
	if first == nil {
		t.Fatal("tab() must hand out a request even when none is open")
	}
	if len(a.Tabs) != 1 || a.ActiveTab != 0 {
		t.Fatalf("tabs = %d, active = %d, want 1 and 0", len(a.Tabs), a.ActiveTab)
	}

	a.openTab()
	if len(a.Tabs) != 2 || a.ActiveTab != 1 {
		t.Fatalf("after openTab tabs = %d, active = %d, want 2 and 1", len(a.Tabs), a.ActiveTab)
	}
	if a.tab() == first {
		t.Fatal("the new tab should become the active one")
	}

	// Closing a tab before the active one keeps the same request active.
	a.Tabs[1].URL = "https://example.com/second"
	a.closeTab(0)
	if len(a.Tabs) != 1 || a.ActiveTab != 0 {
		t.Fatalf("after closeTab tabs = %d, active = %d, want 1 and 0", len(a.Tabs), a.ActiveTab)
	}
	if a.tab().URL != "https://example.com/second" {
		t.Fatalf("active tab = %q, want the second request", a.tab().URL)
	}

	// Closing the last tab leaves a fresh request rather than an empty pane.
	a.closeTab(0)
	if len(a.Tabs) != 1 || a.ActiveTab != 0 {
		t.Fatalf("after closing the last tab tabs = %d, active = %d, want 1 and 0", len(a.Tabs), a.ActiveTab)
	}
}

func TestTabRequestsAreIndependent(t *testing.T) {
	a := &App{}
	first := a.tab()
	first.URL = "https://example.com/first"
	first.Method = "GET"

	a.openTab()
	second := a.tab()
	if second.URL == first.URL {
		t.Fatalf("a new tab should start blank, got %q", second.URL)
	}
	second.URL = "https://example.com/second"
	second.Method = "POST"

	if first.URL != "https://example.com/first" || first.Method != "GET" {
		t.Fatalf("editing a tab changed another: %s %s", first.Method, first.URL)
	}
	a.ActiveTab = 0
	if a.tab().URL != "https://example.com/first" {
		t.Fatalf("switching back gave %q", a.tab().URL)
	}
}

func TestTabTitle(t *testing.T) {
	cases := []struct {
		tab  Tab
		want string
	}{
		{Tab{Method: "POST", URL: "https://example.com/orders"}, "POST example.com/orders"},
		{Tab{Method: "GET", URL: ""}, "Untitled"},
		{Tab{Method: "GET", URL: "https://example.com/orders", Name: "Create order"}, "Create order"},
	}
	for _, c := range cases {
		if got := c.tab.Title(); got != c.want {
			t.Errorf("Title() = %q, want %q", got, c.want)
		}
	}
}
