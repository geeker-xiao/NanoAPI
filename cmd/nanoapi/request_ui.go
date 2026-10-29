package main

import (
	"github.com/egoist/mygo/ui"
)

func (a *App) headersEditor(c *ui.Context, t *ui.Theme) {
	ui.Field(c, "User-Agent", func() {
		ui.TextInput(c, &a.UserAgent).Placeholder("NanoAPI/1.0").Grow(1).OnChange(a.queueSaveSettings)
	}).Description("Sent as the request User-Agent unless a custom User-Agent header overrides it.")
	a.headerKeyValueEditor(c, t)
	ui.Text(c, "Header rows can carry Authorization, Content-Type, cookies, tracing IDs, or any custom header.").FontSize(12).TextColor(t.TextMuted)
}

func coloredTabs(c *ui.Context, selected *int, labels []string, color ui.Color) {
	t := c.Theme()
	tabs := ui.TabsBase(c, selected, len(labels))
	tabs.List.FillWidth().Gap(2).Children(func() {
		for i, label := range labels {
			tab := tabs.Tab(i).Padding(8, 12).Radius(t.Radius)
			if i == *selected {
				tab.Background(t.Surface)
			}
			tab.Children(func() {
				ui.Text(c, label).TextColor(color)
			})
		}
	})
}

func (a *App) appearanceEditor(c *ui.Context, t *ui.Theme) {
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, "Appearance").Bold()
		ui.Text(c, "Changes apply immediately and are kept on this device.").TextColor(t.TextMuted)
	})

	ui.Form(c, func() {
		ui.Field(c, "Theme", func() {
			ui.Select(c, &a.ThemeMode, []string{"System", "Light", "Dark"}).Width(180).OnChange(a.queueSaveSettings)
		})
		ui.Field(c, "Font", func() {
			ui.Select(c, &a.FontFamily, a.fontOptions()).Width(240).OnChange(a.queueSaveSettings)
		})
		ui.Field(c, "Font size", func() {
			ui.NumberInput(c, &a.FontSize, 8, 32, 1).Width(104).OnChange(a.queueSaveSettings)
		})
		ui.Field(c, "Tab color", func() {
			ui.Row(c).Height(28).AlignItems(ui.Center).Children(func() {
				compactColorWell(c, &a.TabColor, "Tab color").OnChange(a.queueSaveSettings)
			})
		})
		ui.Field(c, "Background mode", func() {
			ui.Select(c, &a.BackgroundMode, []string{"Solid color", "Image"}).Width(180).Height(28).OnChange(a.queueSaveSettings)
		})
		ui.Field(c, "Background color", func() {
			ui.Row(c).Height(28).AlignItems(ui.Center).Children(func() {
				compactColorWell(c, &a.BackgroundColor, "Background color").OnChange(a.queueSaveSettings)
			})
		})
		ui.Field(c, "Background image", func() {
			ui.Row(c).FillWidth().Gap(8).Children(func() {
				if a.backgroundBitmap != nil {
					ui.Image(c, a.backgroundBitmap).Size(72, 40).Fit(ui.Cover).Radius(6)
				} else {
					ui.Box(c).Size(72, 40).Background(t.Surface).Border(1, t.Border).Radius(6)
				}
				ui.Text(c, shortPath(a.BackgroundPath)).Grow(1).SingleLine()
				if ui.Button(c, "").Label("Choose image").Tooltip("Choose a PNG, JPEG, GIF, WebP, or BMP background").Children(func() {
					ui.Icon(c, iconUpload)
					ui.Text(c, "Choose").SingleLine()
				}).Clicked() {
					go a.chooseBackgroundImage()
				}
				if ui.Button(c, "").Label("Clear background image").Tooltip("Remove the background image").Disabled(a.BackgroundPath == "").Children(func() {
					ui.Icon(c, iconCancel)
					ui.Text(c, "Clear").SingleLine()
				}).Clicked() {
					a.clearBackgroundImage()
				}
			})
		}).Description("Image is cropped to cover the window; solid color remains available as a fallback.")
	})
}

func compactColorWell(c *ui.Context, value *ui.Color, label string) ui.Element {
	return ui.ColorWell(c, value).Label(label).Size(46, 30).MinWidth(46).MaxWidth(46).Height(30).Shrink(0).AlignSelf(ui.Center)
}

func (a *App) applyTheme(c *ui.Context) {
	var theme ui.Theme
	switch a.ThemeMode {
	case "Light":
		theme = *ui.LightTheme()
	case "Dark":
		theme = *ui.DarkTheme()
	default:
		theme = *c.Theme()
	}
	if a.FontFamily != "" && a.FontFamily != "System" {
		theme.Font = a.FontFamily
	}
	if a.FontSize >= 8 {
		theme.FontSize = float32(a.FontSize)
	}
	if a.BackgroundMode == "Solid color" {
		theme.Background = a.BackgroundColor
	}
	c.SetTheme(&theme)
}

func (a *App) fontOptions() []string {
	options := []string{"System"}
	for _, family := range ui.FontFamilies() {
		if family.Name != "" {
			options = append(options, family.Name)
		}
	}
	return options
}
