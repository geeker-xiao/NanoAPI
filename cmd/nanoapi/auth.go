package main

import (
	"encoding/base64"
	"strings"

	"github.com/egoist/mygo/ui"

	"nanoapi/locales"
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

// authModes are the credential types offered, in the order shown. They are
// stored in requests.json and compared by literal below, so they are also what
// the picker shows; authModeOptions is the one place that may rename them.
var authModes = []string{"None", "Bearer token", "Basic auth", "API key"}

// authModeOptions says the same four types in the language the app shows. Only
// the label translates: "Bearer" and "Basic" are HTTP authentication scheme
// names and "API key" is the term of art, so the stored values stay English.
func authModeOptions() []labeledOption {
	return []labeledOption{
		{"None", locales.AuthModeNone()},
		{"Bearer token", locales.AuthModeBearer()},
		{"Basic auth", locales.AuthModeBasic()},
		{"API key", locales.AuthModeApiKey()},
	}
}

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
			return locales.AuthNoToken()
		}
		return locales.AuthBearerSends()
	case "Basic auth":
		return locales.AuthBasicSends()
	case "API key":
		return locales.AuthApiKeySends()
	}
	return locales.AuthNoCredential()
}

// authEditor edits the credential of the active request.
func (a *App) authEditor(c *ui.Context, t *ui.Theme) {
	tab := a.tab()
	ui.Form(c, func() {
		ui.Field(c, locales.AuthType(), func() {
			// The stored mode stays the English literal; only the name follows
			// the language, so this picker cannot be a plain ui.Select.
			optionSelect(c, &tab.Auth.Mode, &tab.authApplied, t.Rem(12.9), authModeOptions(), nil)
		})
		switch tab.Auth.Mode {
		case "Bearer token":
			// "Token" is the credential's name in the protocol, not prose.
			ui.Field(c, "Token", func() {
				ui.TextInput(c, &tab.Auth.Token).Placeholder("eyJhbGciOi...").Grow(1)
			})
		case "Basic auth":
			ui.Field(c, locales.AuthUsername(), func() {
				ui.TextInput(c, &tab.Auth.Username).Placeholder("username").Grow(1)
			})
			ui.Field(c, locales.AuthPassword(), func() {
				ui.TextInput(c, &tab.Auth.Password).Placeholder("password").Grow(1)
			})
		case "API key":
			ui.Field(c, locales.AuthKeyName(), func() {
				ui.TextInput(c, &tab.Auth.KeyName).Placeholder("X-API-Key").Grow(1)
			})
			ui.Field(c, locales.AuthKeyValue(), func() {
				ui.TextInput(c, &tab.Auth.KeyValue).Placeholder("key value").Grow(1)
			})
			// Header and Query are the two stored destinations and the names
			// they are called by on the wire, so they stay as they are.
			ui.Field(c, locales.AuthAddTo(), func() {
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
