package main

import (
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"nanoapi/locales"
)

// languageSystem is the stored value meaning "ask the operating system", which
// is what SetLocale("") asks of MyGo.
const languageSystem = "System"

// envLocale is the variable that picks a language for one run and keeps nothing
// (`mygo dev -lang` sets it). It is how untranslated text gets hunted down with
// the en-XA pseudo locale, so the app must not step on it at start-up.
const envLocale = "MYGO_LOCALE"

// The Language field offers the tags as they are stored in settings.json, under
// the name each language writes for itself; only "Follow system" follows the
// app's language. See labeledOption in request_ui.go.
func languageOptions() []labeledOption {
	return []labeledOption{
		{languageSystem, locales.FollowSystem()},
		{"en", "English"},
		{"zh-Hans", "简体中文"},
	}
}

// isKnownLanguage reports whether tag is one the app offers, so a settings file
// from elsewhere cannot leave the app showing a language it has no catalog for.
func isKnownLanguage(tag string) bool {
	for _, option := range languageOptions() {
		if option.value == tag {
			return true
		}
	}
	return false
}

func languageLabel(tag string) string {
	return optionLabel(languageOptions(), tag)
}

// languageRequest works out what to hand SetLocale, without touching MyGo: ok is
// false when the language must be left alone. "System" means an empty tag, which
// tells MyGo to follow the reader's own languages again — but when the run was
// given a language through the environment, asking it back would clear that and
// put the system language in its place, which is exactly what the en-XA audit
// pass needs to survive.
func languageRequest(language string) (tag string, ok bool) {
	if !isKnownLanguage(language) {
		return "", false
	}
	if language == languageSystem {
		if os.Getenv(envLocale) != "" {
			return "", false
		}
		return "", true
	}
	return language, true
}

// applyLanguage makes the app show the saved language. It is kept apart from the
// picker so the start-up path and a later change run through the same code.
func (a *App) applyLanguage() {
	a.languageApplied = a.Language
	tag, ok := languageRequest(a.Language)
	if !ok {
		return
	}
	mygo.App.SetLocale(tag)
}

// languageSelect picks the language. The tag is stored and the name shown, so
// switching languages never rewrites the setting; applying it waits for the next
// tick because SetLocale redraws the windows, and redrawing from the middle of a
// redraw would re-enter the view that is still laying this field out.
func (a *App) languageSelect(c *ui.Context) {
	t := c.Theme()
	optionSelect(c, &a.Language, &a.languageApplied, t.Rem(12), languageOptions(), func() {
		a.queueSaveSettings()
		a.publish(func() { a.applyLanguage() })
	})
}

// installTrayMenu says the tray's two items in the language the app shows. MyGo
// relabels its own menus when the language changes, but this menu belongs to the
// app, so OnLocaleChanged has to rebuild it.
func (a *App) installTrayMenu(win *mygo.Window) {
	if a.tray == nil {
		return
	}
	app := locales.AppName()
	a.tray.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		{Label: locales.TrayShow(app), Click: func(_ *mygo.MenuItem, _ *mygo.Window) {
			win.Show()
			win.Focus()
		}},
		mygo.Separator(),
		{Label: locales.TrayQuit(app), Click: func(_ *mygo.MenuItem, _ *mygo.Window) {
			a.quitApp()
		}},
	}))
}
