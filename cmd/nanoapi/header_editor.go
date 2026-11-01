package main

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"nanoapi/internal/httpclient"
)

var (
	headerValueModes  = []string{httpclient.ValueModeFixed, httpclient.ValueModeRandom}
	headerRandomModes = []string{
		httpclient.RandomUUIDv4,
		httpclient.RandomUUIDv7,
		httpclient.RandomULID,
		httpclient.RandomString,
		httpclient.RandomInteger,
		httpclient.RandomTimestampMS,
		httpclient.RandomTimestampSec,
		httpclient.RandomISOTime,
		httpclient.RandomTimeAndString,
		httpclient.RandomTemplate,
	}
)

func (a *App) headerKeyValueEditor(c *ui.Context, t *ui.Theme) {
	tab := a.tab()
	ui.Row(c).FillWidth().Children(func() {
		ui.Text(c, "Custom headers").Bold()
		ui.Spacer(c)
		if ui.Button(c, "").Label("Add header").Tooltip("Add a key/value header").Children(func() {
			ui.Icon(c, iconPlus)
			ui.Text(c, "Add").SingleLine()
		}).Clicked() {
			tab.Headers = append(tab.Headers, KeyValue{
				Enabled:    true,
				ValueMode:  httpclient.ValueModeFixed,
				RandomMode: httpclient.RandomUUIDv4,
			})
		}
	})

	deleteIndex := -1
	ui.Column(c).FillWidth().Gap(6).Children(func() {
		for i := range tab.Headers {
			item := &tab.Headers[i]
			normalizeHeader(item)
			ui.Row(c.Key(i)).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Checkbox(c, &item.Enabled, "").Label("Enable header").Tooltip("Enable this header")
				ui.TextInput(c, &item.Key).Placeholder("header-name").Width(170)
				ui.Select(c, &item.ValueMode, headerValueModes).Width(92)
				if item.ValueMode == httpclient.ValueModeRandom {
					ui.Select(c, &item.RandomMode, headerRandomModes).Width(178)
					ui.TextInput(c, &item.Pattern).Placeholder(randomPlaceholder(item.RandomMode)).Grow(1)
				} else {
					ui.TextInput(c, &item.Value).Placeholder("value").Grow(1)
				}
				if ui.Button(c, "").Label("Delete header").Tooltip("Delete this header").Size(28, 28).Children(func() {
					ui.Icon(c, iconTrash).FontSize(14)
				}).Clicked() {
					deleteIndex = i
				}
			})
		}
	})
	if deleteIndex >= 0 {
		tab.Headers = append(tab.Headers[:deleteIndex], tab.Headers[deleteIndex+1:]...)
	}
	ui.Text(c, "Template values support {{uuid}}, {{uuidv7}}, {{ulid}}, {{timestamp}}, {{random:12}}, {{iso8601}} and literal text.").
		FontSize(11).TextColor(t.TextMuted)
}

func normalizeHeader(item *KeyValue) {
	if !strings.EqualFold(strings.TrimSpace(item.ValueMode), httpclient.ValueModeRandom) {
		item.ValueMode = httpclient.ValueModeFixed
		return
	}
	item.ValueMode = httpclient.ValueModeRandom
	if strings.TrimSpace(item.RandomMode) == "" {
		item.RandomMode = httpclient.RandomUUIDv4
	}
}

func randomPlaceholder(mode string) string {
	switch mode {
	case httpclient.RandomString:
		return "length, default 8"
	case httpclient.RandomTemplate:
		return "prefix-{{timestamp}}-{{random:12}}"
	case httpclient.RandomTimeAndString:
		return "optional prefix"
	case httpclient.RandomTimestampMS, httpclient.RandomTimestampSec, httpclient.RandomISOTime:
		return "generated automatically"
	default:
		return "generated automatically"
	}
}
