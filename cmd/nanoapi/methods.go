package main

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

var (
	httpMethods     = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "PUSH"}
	iconChevronDown = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m6 9 6 6 6-6"/></svg>`))
	iconMinus       = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M5 12h14"/></svg>`))
)

// Method colors come in pairs: the deeper tones read on the light page and the
// brighter ones on the dark page, so a method keeps its hue in either theme.
var (
	methodLight = map[string]ui.Color{
		"GET":     ui.Hex("#16a34a"),
		"POST":    ui.Hex("#2563eb"),
		"PUT":     ui.Hex("#d97706"),
		"PATCH":   ui.Hex("#7c3aed"),
		"DELETE":  ui.Hex("#dc2626"),
		"HEAD":    ui.Hex("#0f766e"),
		"OPTIONS": ui.Hex("#64748b"),
		"PUSH":    ui.Hex("#db2777"),
	}
	methodDark = map[string]ui.Color{
		"GET":     ui.Hex("#4ade80"),
		"POST":    ui.Hex("#60a5fa"),
		"PUT":     ui.Hex("#fbbf24"),
		"PATCH":   ui.Hex("#a78bfa"),
		"DELETE":  ui.Hex("#f87171"),
		"HEAD":    ui.Hex("#2dd4bf"),
		"OPTIONS": ui.Hex("#94a3b8"),
		"PUSH":    ui.Hex("#f472b6"),
	}
)

func methodColor(t *ui.Theme, method string) ui.Color {
	name := strings.ToUpper(strings.TrimSpace(method))
	if t != nil && t.Dark {
		if color, ok := methodDark[name]; ok {
			return color
		}
		return ui.Hex("#cbd5e1")
	}
	if color, ok := methodLight[name]; ok {
		return color
	}
	return ui.Hex("#475569")
}

func (a *App) methodSelect(c *ui.Context) {
	tab := a.tab()
	t := c.Theme()
	color := methodColor(t, tab.Method)
	sel := ui.SelectBase(c, &tab.Method)
	// Width and inner sizes follow the font, so the longest method still fits
	// and the trigger keeps pace with the URL field beside it.
	// No border of its own: the trigger is one half of the box that also holds
	// the URL, and only lights up when hovered.
	sel.Trigger.Width(t.Rem(8.7)).Padding(t.Space(1.5), t.Space(2)).Gap(t.Space(2)).AlignItems(ui.Center).Radius(t.Radius)
	if sel.Trigger.Hovered() {
		sel.Trigger.Background(t.SurfaceHover)
	}
	sel.Trigger.Children(func() {
		ui.Box(c).Width(t.Space(1)).Height(t.Rem(1.3)).Background(color).Radius(2)
		ui.Text(c, tab.Method).Bold().TextColor(color).SingleLine()
		ui.Spacer(c)
		ui.Icon(c, iconChevronDown).FontSize(t.FontSize).TextColor(t.TextMuted)
	})
	sel.Popup(func(panel ui.Element) {
		panel.Padding(t.Space(1)).Background(t.Surface).Border(1, t.Border).Radius(t.Radius)
		for _, method := range httpMethods {
			color := methodColor(t, method)
			item := sel.Item(method).Padding(t.Space(1.75), t.Space(2.25)).Gap(t.Space(2)).Radius(t.Radius)
			if item.Highlighted() {
				item.Background(t.SurfaceHover)
			}
			item.Children(func() {
				ui.Box(c).Width(t.Space(1)).Height(t.Rem(1.3)).Background(color).Radius(2)
				ui.Text(c, method).Bold().TextColor(color).SingleLine()
			})
		}
	})
}

func (a *App) timeoutStepper(c *ui.Context) {
	tab := a.tab()
	t := c.Theme()
	// Every size follows the font: at 14 they land on the 32 and 72 the fixed
	// layout used, and grow together from there.
	size := t.Rem(2.3)
	ui.Row(c).Gap(t.Space(1.5)).AlignItems(ui.Center).Children(func() {
		ui.TextInput(c, &tab.TimeoutText).Width(t.Rem(5.2)).Height(size).MinHeight(size).MaxHeight(size).TextAlign(ui.Center).OnChange(a.parseTimeout)
		if ui.Button(c, "").Label("Decrease timeout").Tooltip("Decrease timeout").Size(size, size).MinHeight(size).MaxHeight(size).Padding(0).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			ui.Icon(c, iconMinus).FontSize(t.FontSize)
		}).Clicked() {
			a.adjustTimeout(-1)
		}
		if ui.Button(c, "").Label("Increase timeout").Tooltip("Increase timeout").Size(size, size).MinHeight(size).MaxHeight(size).Padding(0).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			ui.Icon(c, iconPlus).FontSize(t.FontSize)
		}).Clicked() {
			a.adjustTimeout(1)
		}
		// No fixed width: "seconds" takes the room its font needs, so it stops
		// wrapping and pushes what follows further right as the font grows.
		ui.Text(c, "seconds").AlignSelf(ui.Center).TextColor(t.TextMuted).SingleLine()
	})
}

func (a *App) parseTimeout() {
	tab := a.tab()
	value, err := strconv.ParseFloat(strings.TrimSpace(tab.TimeoutText), 64)
	if err != nil || value < 1 || value > 600 {
		return
	}
	tab.TimeoutSeconds = value
}

func (a *App) adjustTimeout(delta float64) {
	tab := a.tab()
	tab.TimeoutSeconds += delta
	if tab.TimeoutSeconds < 1 {
		tab.TimeoutSeconds = 1
	}
	if tab.TimeoutSeconds > 600 {
		tab.TimeoutSeconds = 600
	}
	tab.TimeoutText = formatTimeout(tab.TimeoutSeconds)
}

func formatTimeout(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
