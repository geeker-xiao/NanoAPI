package main

import (
	"math"

	"github.com/egoist/mygo/ui"
)

func (a *App) headersEditor(c *ui.Context, t *ui.Theme) {
	tab := a.tab()
	ui.Field(c, "User-Agent", func() {
		ui.TextInput(c, &tab.UserAgent).Placeholder("NanoAPI/1.0").Grow(1).OnChange(a.queueSaveSettings)
	}).Description("Sent as the request User-Agent unless a custom User-Agent header overrides it.")
	a.headerKeyValueEditor(c, t)
	ui.Text(c, "Header rows can carry Authorization, Content-Type, cookies, tracing IDs, or any custom header.").FontSize(12).TextColor(t.TextMuted)
}

// Light palette, taken from the reference design: a light grey page, white
// surfaces, a near-black primary and a green accent.
var (
	styleBackground     = ui.Hex("#f5f5f5")
	styleSurface        = ui.Hex("#ffffff")
	styleSurfaceHover   = ui.Hex("#f0f0f0")
	styleSurfacePressed = ui.Hex("#e8e8e8")
	styleBorder         = ui.Hex("#e5e5e5")
	styleText           = ui.Hex("#111111")
	styleTextMuted      = ui.Hex("#8c8c8c")
	styleAccent         = ui.Hex("#111111")
	styleAccentText     = ui.Hex("#ffffff")
	styleSuccess        = ui.Hex("#22c55e")
)

// Dark counterpart of the palette: the page takes the near-black of the
// reference design, cards lift off it, and the primary button becomes light ink
// with dark text, the way the reference design's button sits on its dark card.
// The muted grey and the green accent are shared, so both themes read alike.
var (
	styleDarkBackground     = ui.Hex("#111111")
	styleDarkSurface        = ui.Hex("#1c1c1c")
	styleDarkSurfaceHover   = ui.Hex("#242424")
	styleDarkSurfacePressed = ui.Hex("#2c2c2c")
	styleDarkBorder         = ui.Hex("#2e2e2e")
	styleDarkText           = ui.Hex("#f5f5f5")
	styleDarkAccent         = ui.Hex("#f5f5f5")
	styleDarkAccentText     = ui.Hex("#111111")
)

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
				if i == *selected {
					ui.Text(c, label).Bold().TextColor(readableTextColor(t, color))
				} else {
					ui.Text(c, label).TextColor(t.TextMuted)
				}
			})
		}
	})
}

// luminance ranks a color from black (0) to white (1), enough to tell whether
// two colors read against each other.
func luminance(color ui.Color) float64 {
	return (0.2126*float64(color.R) + 0.7152*float64(color.G) + 0.0722*float64(color.B)) / 255
}

// readableTextColor keeps a picked color when it stands out from the page
// background, and falls back to body text when it does not: the near-black tab
// color of the palette is invisible once the theme turns dark.
func readableTextColor(t *ui.Theme, color ui.Color) ui.Color {
	if math.Abs(luminance(color)-luminance(t.Background)) >= 0.35 {
		return color
	}
	return t.Text
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
		ui.Field(c, tabBarFieldLabel, func() {
			ui.Select(c, &a.TabBarDisplay, tabBarModes).Width(180).OnChange(a.queueSaveSettings)
		}).Description(tabBarFieldDescription)
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

// tabBarFieldLabel and tabBarFieldDescription explain the two ways the request
// tabs can be laid out, offered in Appearance.
const (
	tabBarFieldLabel       = "Request tabs"
	tabBarFieldDescription = "Multi-line wraps the tabs onto as many rows as they need, showing method above name. Single line keeps one compact row that scrolls sideways, with the close button inside each tab."
)

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
	if theme.Dark {
		theme = applyDarkPalette(theme)
	} else {
		theme = applyLightPalette(theme)
	}
	if a.BackgroundMode == "Solid color" {
		// A background still at the other theme's default follows this theme
		// instead; a color the user actually picked is kept as it is.
		switch {
		case theme.Dark && a.BackgroundColor == styleBackground:
			theme.Background = styleDarkBackground
		case !theme.Dark && a.BackgroundColor == styleDarkBackground:
			theme.Background = styleBackground
		default:
			theme.Background = a.BackgroundColor
		}
	}
	c.SetTheme(&theme)
}

// applyLightPalette restyles a light theme with the reference palette: a light
// grey page behind white surfaces, a near-black primary and green success, and
// rounder corners. Paired with applyDarkPalette.
func applyLightPalette(theme ui.Theme) ui.Theme {
	theme.Background = styleBackground
	theme.Surface = styleSurface
	theme.SurfaceHover = styleSurfaceHover
	theme.SurfacePressed = styleSurfacePressed
	theme.Border = styleBorder
	theme.Text = styleText
	theme.TextMuted = styleTextMuted
	theme.Accent = styleAccent
	theme.AccentHover = ui.Hex("#000000")
	theme.AccentPressed = ui.Hex("#000000")
	theme.AccentText = styleAccentText
	theme.Success = styleSuccess
	theme.Selection = ui.RGBA(17, 17, 17, 0.15)
	theme.Focus = ui.RGBA(17, 17, 17, 0.45)
	theme.Scrollbar = ui.RGBA(0, 0, 0, 0.22)
	theme.Radius = 10
	return theme
}

// applyDarkPalette is the dark half of applyLightPalette: same identity, the
// page and the ink exchanging places while the green accent is kept.
func applyDarkPalette(theme ui.Theme) ui.Theme {
	theme.Background = styleDarkBackground
	theme.Surface = styleDarkSurface
	theme.SurfaceHover = styleDarkSurfaceHover
	theme.SurfacePressed = styleDarkSurfacePressed
	theme.Border = styleDarkBorder
	theme.Text = styleDarkText
	theme.TextMuted = styleTextMuted
	theme.Accent = styleDarkAccent
	theme.AccentHover = ui.Hex("#ffffff")
	theme.AccentPressed = ui.Hex("#e0e0e0")
	theme.AccentText = styleDarkAccentText
	theme.Success = styleSuccess
	theme.Danger = ui.Hex("#ef4444")
	theme.Warning = ui.Hex("#f59e0b")
	theme.Selection = ui.RGBA(255, 255, 255, 0.16)
	theme.Focus = ui.RGBA(255, 255, 255, 0.5)
	theme.Scrollbar = ui.RGBA(255, 255, 255, 0.28)
	theme.Radius = 10
	return theme
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
