package main

import (
	"encoding/base64"
	"strings"

	"github.com/egoist/mygo/ui"
)

// Auth holds how a request identifies itself. It travels with a saved request
// and is kept apart from the header rows, so a credential can be swapped
// without hunting through headers for it.
type Auth struct {
	Mode     string `json:"mode,omitempty"`
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	KeyName  string `json:"keyName,omitempty"`
	KeyValue string `json:"keyValue,omitempty"`
	// KeyIn is where an API key travels: "Header" (the default) or "Query".
	KeyIn string `json:"keyIn,omitempty"`
}

// authModes are the credential types offered, in the order shown.
var authModes = []string{"None", "Bearer token", "Basic auth", "API key"}

// header is the credential as a header row, plus whether this kind carries one
// and is filled in far enough to send.
func (auth Auth) header() (KeyValue, bool) {
	switch auth.Mode {
	case "Bearer token":
		token := strings.TrimSpace(auth.Token)
		if token == "" {
			return KeyValue{}, false
		}
		return KeyValue{Enabled: true, Key: "Authorization", Value: "Bearer " + token}, true
	case "Basic auth":
		if auth.Username == "" && auth.Password == "" {
			return KeyValue{}, false
		}
		encoded := base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password))
		return KeyValue{Enabled: true, Key: "Authorization", Value: "Basic " + encoded}, true
	case "API key":
		if auth.KeyIn == "Query" {
			return KeyValue{}, false
		}
		name := strings.TrimSpace(auth.KeyName)
		if name == "" || strings.TrimSpace(auth.KeyValue) == "" {
			return KeyValue{}, false
		}
		return KeyValue{Enabled: true, Key: name, Value: strings.TrimSpace(auth.KeyValue)}, true
	}
	return KeyValue{}, false
}

// param is the credential as a query row, which only an API key can be.
func (auth Auth) param() (KeyValue, bool) {
	if auth.Mode != "API key" || auth.KeyIn != "Query" {
		return KeyValue{}, false
	}
	name := strings.TrimSpace(auth.KeyName)
	if name == "" || strings.TrimSpace(auth.KeyValue) == "" {
		return KeyValue{}, false
	}
	return KeyValue{Enabled: true, Key: name, Value: strings.TrimSpace(auth.KeyValue)}, true
}

// summary names the credential in one line, shown on the Auth tab.
func (auth Auth) summary() string {
	switch auth.Mode {
	case "Bearer token":
		if strings.TrimSpace(auth.Token) == "" {
			return "No token yet."
		}
		return "Sends Authorization: Bearer <token>."
	case "Basic auth":
		return "Sends Authorization: Basic <base64 of user:password>."
	case "API key":
		return "Sends the key as a header row, or as a query row when Query is picked."
	}
	return "This request sends no credential yet."
}

// authEditor edits the credential of the active request.
func (a *App) authEditor(c *ui.Context, t *ui.Theme) {
	tab := a.tab()
	ui.Form(c, func() {
		ui.Field(c, "Type", func() {
			ui.Select(c, &tab.Auth.Mode, authModes).Width(180)
		})
		switch tab.Auth.Mode {
		case "Bearer token":
			ui.Field(c, "Token", func() {
				ui.TextInput(c, &tab.Auth.Token).Placeholder("eyJhbGciOi...").Grow(1)
			})
		case "Basic auth":
			ui.Field(c, "Username", func() {
				ui.TextInput(c, &tab.Auth.Username).Placeholder("username").Grow(1)
			})
			ui.Field(c, "Password", func() {
				ui.TextInput(c, &tab.Auth.Password).Placeholder("password").Grow(1)
			})
		case "API key":
			ui.Field(c, "Key", func() {
				ui.TextInput(c, &tab.Auth.KeyName).Placeholder("X-API-Key").Grow(1)
			})
			ui.Field(c, "Value", func() {
				ui.TextInput(c, &tab.Auth.KeyValue).Placeholder("key value").Grow(1)
			})
			ui.Field(c, "Add to", func() {
				ui.Select(c, &tab.Auth.KeyIn, []string{"Header", "Query"}).Width(180)
			})
		}
	})
	ui.Text(c, tab.Auth.summary()).FontSize(12).TextColor(t.TextMuted)
}

// withAuth is a tab's parameters and headers with its credential folded in. A
// row typed by hand under the same key wins, so nothing is overwritten.
func withAuth(tab *Tab) ([]KeyValue, []KeyValue) {
	params := cloneKV(tab.Params)
	headers := cloneKV(tab.Headers)
	if row, ok := tab.Auth.param(); ok {
		params = putKV(params, row)
	}
	if row, ok := tab.Auth.header(); ok {
		headers = putKV(headers, row)
	}
	return params, headers
}
func putKV(rows []KeyValue, row KeyValue) []KeyValue {
	for _, existing := range rows {
		if strings.EqualFold(strings.TrimSpace(existing.Key), row.Key) {
			return rows
		}
	}
	return append(rows, row)
}
