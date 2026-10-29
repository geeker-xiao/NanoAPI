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

func methodColor(method string) ui.Color {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET":
		return ui.Hex("#16a34a")
	case "POST":
		return ui.Hex("#2563eb")
	case "PUT":
		return ui.Hex("#d97706")
	case "PATCH":
		return ui.Hex("#7c3aed")
	case "DELETE":
		return ui.Hex("#dc2626")
	case "HEAD":
		return ui.Hex("#0f766e")
	case "OPTIONS":
		return ui.Hex("#64748b")
	case "PUSH":
		return ui.Hex("#db2777")
	default:
		return ui.Hex("#475569")
	}
}

func (a *App) methodSelect(c *ui.Context) {
	t := c.Theme()
	color := methodColor(a.Method)
	sel := ui.SelectBase(c, &a.Method)
	sel.Trigger.Width(122).Padding(8, 10).Gap(8).Border(1, t.Border).Radius(t.Radius).Children(func() {
		ui.Box(c).Width(4).Height(18).Background(color).Radius(2)
		ui.Text(c, a.Method).Bold().TextColor(color).SingleLine()
		ui.Spacer(c)
		ui.Icon(c, iconChevronDown).FontSize(14).TextColor(t.TextMuted)
	})
	sel.Popup(func(panel ui.Element) {
		panel.Padding(4).Background(t.Surface).Border(1, t.Border).Radius(t.Radius)
		for _, method := range httpMethods {
			color := methodColor(method)
			item := sel.Item(method).Padding(7, 9).Gap(8).Radius(t.Radius)
			if item.Highlighted() {
				item.Background(t.SurfaceHover)
			}
			item.Children(func() {
				ui.Box(c).Width(4).Height(18).Background(color).Radius(2)
				ui.Text(c, method).Bold().TextColor(color).SingleLine()
			})
		}
	})
}

func (a *App) requestSummary(c *ui.Context, t *ui.Theme) {
	ui.Row(c).Gap(6).Children(func() {
		ui.Text(c, a.Method).Bold().TextColor(methodColor(a.Method)).SingleLine()
		ui.Text(c, shortURL(a.URL)).TextColor(t.TextMuted).SingleLine()
	})
}

func (a *App) timeoutStepper(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Height(32).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.TextInput(c, &a.TimeoutText).Width(72).Height(32).MinHeight(32).MaxHeight(32).TextAlign(ui.Center).OnChange(a.parseTimeout)
		if ui.Button(c, "").Label("Decrease timeout").Tooltip("Decrease timeout").Size(32, 32).MinHeight(32).MaxHeight(32).Padding(0).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			ui.Icon(c, iconMinus).FontSize(14)
		}).Clicked() {
			a.adjustTimeout(-1)
		}
		if ui.Button(c, "").Label("Increase timeout").Tooltip("Increase timeout").Size(32, 32).MinHeight(32).MaxHeight(32).Padding(0).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			ui.Icon(c, iconPlus).FontSize(14)
		}).Clicked() {
			a.adjustTimeout(1)
		}
		ui.Text(c, "seconds").Width(56).AlignSelf(ui.Center).TextColor(t.TextMuted)
	})
}

func (a *App) parseTimeout() {
	value, err := strconv.ParseFloat(strings.TrimSpace(a.TimeoutText), 64)
	if err != nil || value < 1 || value > 600 {
		return
	}
	a.TimeoutSeconds = value
}

func (a *App) adjustTimeout(delta float64) {
	a.TimeoutSeconds += delta
	if a.TimeoutSeconds < 1 {
		a.TimeoutSeconds = 1
	}
	if a.TimeoutSeconds > 600 {
		a.TimeoutSeconds = 600
	}
	a.TimeoutText = formatTimeout(a.TimeoutSeconds)
}

func formatTimeout(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
