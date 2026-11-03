package main

import (
	"encoding/json"
	"testing"

	"nanoapi/locales"
)

// The rule these tests hold: a language change is only ever a change of what
// the app says. What the app stores must not move, because settings.json and
// requests.json on real machines already hold these very strings and the
// readers compare them by literal.

func TestSettingsSnapshotOnlyMovesLanguage(t *testing.T) {
	a := &App{Tabs: []*Tab{newTab()}, ThemeMode: "System", Language: "en", TabBarDisplay: tabBarMultiLine, BackgroundMode: "Solid color"}
	snapshot, err := json.Marshal(a.settingsSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	a.Language = "zh-Hans"
	other, err := json.Marshal(a.settingsSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	// Every stored field but the language itself has to survive untouched, and
	// none of them may turn into a translated label.
	var english, chinese map[string]any
	if err := json.Unmarshal(snapshot, &english); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(other, &chinese); err != nil {
		t.Fatal(err)
	}
	if len(english) != len(chinese) {
		t.Fatalf("the snapshot gained or lost a field: %v vs %v", english, chinese)
	}
	changed := 0
	for key, want := range english {
		got, ok := chinese[key]
		if !ok {
			t.Fatalf("switching language dropped the stored field %q", key)
		}
		if got != want {
			if key != "language" {
				t.Errorf("switching language rewrote the stored field %q: %v, want %v", key, got, want)
			}
			changed++
		}
	}
	if changed != 1 {
		t.Errorf("only the language should differ between snapshots, %d fields did", changed)
	}
	// Spelled out once more, because this is the exact value a reader of
	// settings.json compares against.
	if english["tabBarDisplay"] != tabBarMultiLine {
		t.Errorf(`tabBarDisplay = %v, want the stored literal %q`, english["tabBarDisplay"], tabBarMultiLine)
	}
	if english["backgroundMode"] != "Solid color" || english["themeMode"] != "System" {
		t.Errorf("stored enums moved with the language: %v / %v", english["backgroundMode"], english["themeMode"])
	}
}

func TestUnknownLanguageFallsBackToSystem(t *testing.T) {
	// A settings file from before the feature carries no language at all, and a
	// tag the app has no catalog for must not be handed to MyGo.
	for _, tag := range []string{"", "fr", "zh_CN", "deutsch", "System"} {
		if got := isKnownLanguage(tag); got != (tag == "System") {
			t.Errorf("isKnownLanguage(%q) = %v", tag, got)
		}
	}
	if !isKnownLanguage("en") || !isKnownLanguage("zh-Hans") {
		t.Fatal("the two languages the app ships must be accepted")
	}
}

func TestLanguageNamesStayInTheirOwnLanguage(t *testing.T) {
	// English and 简体中文 are conventionally written in the language itself, so
	// they do not move when the app switches; only "Follow system" is translated.
	if got := languageLabel("en"); got != "English" {
		t.Errorf("languageLabel(en) = %q", got)
	}
	if got := languageLabel("zh-Hans"); got != "简体中文" {
		t.Errorf("languageLabel(zh-Hans) = %q", got)
	}
	if got := languageLabel(languageSystem); got == "" {
		t.Error("the follow-system option must still say something")
	}
}

// TestEnvironmentChosenLanguageSurvivesStartUp keeps the en-XA audit pass usable:
// MYGO_LOCALE picks a language for one run, and start-up must not ask it back.
func TestEnvironmentChosenLanguageSurvivesStartUp(t *testing.T) {
	// No environment override: "System" hands the choice to the operating
	// system, which is SetLocale("").
	t.Setenv(envLocale, "")
	if tag, ok := languageRequest(languageSystem); !ok || tag != "" {
		t.Errorf("follow-system = (%q, %v), want (\"\", true)", tag, ok)
	}
	if tag, ok := languageRequest("zh-Hans"); !ok || tag != "zh-Hans" {
		t.Errorf("explicit tag = (%q, %v), want (zh-Hans, true)", tag, ok)
	}
	// An unknown tag must not be handed to MyGo at all.
	if _, ok := languageRequest("fr"); ok {
		t.Error("a language the app has no catalog for must not be applied")
	}
	// With the environment having chosen this run, "System" leaves it alone
	// rather than clearing it; an explicit pick still wins.
	t.Setenv(envLocale, "en-XA")
	if _, ok := languageRequest(languageSystem); ok {
		t.Error("follow-system overrode the language the environment picked")
	}
	if tag, ok := languageRequest("zh-Hans"); !ok || tag != "zh-Hans" {
		t.Errorf("explicit tag under MYGO_LOCALE = (%q, %v), want (zh-Hans, true)", tag, ok)
	}
}

// TestOptionsKeepTheirStoredValue protects the rule the whole feature turns on:
// a picker may rename what it shows, never what settings.json holds, because the
// readers compare those strings by literal.
func TestOptionsKeepTheirStoredValue(t *testing.T) {
	for _, check := range []struct {
		name    string
		options []labeledOption
		stored  []string
	}{
		{"theme", themeOptions(), []string{"System", "Light", "Dark"}},
		{"tabBarDisplay", tabBarLayoutOptions(), []string{tabBarMultiLine, tabBarSingleLine}},
		{"backgroundMode", backgroundModeOptions(), []string{"Solid color", "Image"}},
		{"language", languageOptions(), []string{languageSystem, "en", "zh-Hans"}},
	} {
		if len(check.options) != len(check.stored) {
			t.Fatalf("%s offers %d choices, want %d", check.name, len(check.options), len(check.stored))
		}
		for i, want := range check.stored {
			if got := check.options[i].value; got != want {
				t.Errorf("%s choice %d stores %q, want the literal %q", check.name, i, got, want)
			}
		}
	}
}

// TestCatalogSaysEveryMessageInBothLanguages keeps the two catalogs level: a key
// translated into English alone would show English to a reader of Chinese.
func TestCatalogSaysEveryMessageInBothLanguages(t *testing.T) {
	// The catalogs are read by tag, whatever language the app happens to show,
	// so this checks the shipped Chinese without needing a window.
	if got := locales.In("en").Send(); got != "Send" {
		t.Errorf("English send = %q, want %q", got, "Send")
	}
	if got := locales.In("zh-Hans").Send(); got != "发送" {
		t.Errorf("Chinese send = %q, want 发送", got)
	}
	// Chinese has no plural distinctions in CLDR, so one branch carries every
	// count while English still splits at one.
	if got := locales.In("zh-Hans").SavedRequestsCount(1); got != "已保存的请求（1）" {
		t.Errorf("Chinese plural for one = %q", got)
	}
	if got := locales.In("en").SavedRequestsCount(1); got != "Saved request (1)" {
		t.Errorf("English plural for one = %q", got)
	}
	if got := locales.In("en").SavedRequestsCount(7); got != "Saved requests (7)" {
		t.Errorf("English plural for many = %q", got)
	}
	// A placeholder is filled in, not printed literally.
	if got := locales.In("zh-Hans").TrayShow("NanoAPI"); got != "显示 NanoAPI" {
		t.Errorf("Chinese tray show = %q", got)
	}
	// requestTabs are labels only; the number, not the label, is what a request
	// stores, so the list may follow the language.
	if len(requestTabs()) != 6 {
		t.Fatalf("requestTabs() = %d labels, want 6", len(requestTabs()))
	}
	for _, label := range append(responseTabs(), requestTabs()...) {
		if label == "" {
			t.Fatal("a tab label came out empty; the catalog is missing a message")
		}
	}
	if locales.SavedRequestsCount(2) == "" || locales.TrayShow("NanoAPI") == "" {
		t.Error("a message with arguments must still be said")
	}
}
