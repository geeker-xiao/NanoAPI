package main

import (
	"math"

	"github.com/egoist/mygo/ui"

	"nanoapi/locales"
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

// labeledOption is one choice whose stored value is not the name shown for it.
// The value goes into settings.json and is compared by literal, so it cannot be
// translated; the name must be.
type labeledOption struct {
	value string
	label string
}

// The three pickers below say their choices in the language the app shows. Each
// is a function, for the same reason requestTabs is: a package-level value would
// freeze the labels at whatever language the process happened to start in, and
// they would not move when the reader switches.
func themeOptions() []labeledOption {
	return []labeledOption{
		{"System", locales.FollowSystem()},
		{"Light", locales.ThemeLight()},
		{"Dark", locales.ThemeDark()},
	}
}

func tabBarLayoutOptions() []labeledOption {
	return []labeledOption{
		{tabBarMultiLine, locales.TabBarMultiLineLabel()},
		{tabBarSingleLine, locales.TabBarSingleLineLabel()},
	}
}

func backgroundModeOptions() []labeledOption {
	return []labeledOption{
		{"Solid color", locales.BackgroundSolid()},
		{"Image", locales.BackgroundImageOption()},
	}
}

// optionLabel finds the name a value is shown under, or the value itself when
// nothing claims it.
func optionLabel(options []labeledOption, value string) string {
	for _, option := range options {
		if option.value == value {
			return option.label
		}
	}
	return value
}

// optionSelect draws a picker that stores value but shows a label. ui.Select
// cannot do this: it writes back the very string it displays, which would put a
// translated name into settings.json. SelectBase has no OnChange of its own, so
// the pick is noticed on the frame after it by comparing against applied, and
// onChange runs there.
func optionSelect(c *ui.Context, value *string, applied *string, width float32, options []labeledOption, onChange func()) {
	t := c.Theme()
	sel := ui.SelectBase(c, value)
	sel.Trigger.Width(width).Padding(t.Space(1.5), t.Space(2)).Gap(t.Space(2)).AlignItems(ui.Center).Radius(t.Radius)
	if sel.Trigger.Hovered() {
		sel.Trigger.Background(t.SurfaceHover)
	}
	sel.Trigger.Children(func() {
		ui.Text(c, optionLabel(options, *value)).SingleLine()
		ui.Spacer(c)
		ui.Icon(c, iconChevronDown).FontSize(t.FontSize).TextColor(t.TextMuted)
	})
	sel.Popup(func(panel ui.Element) {
		panel.Padding(t.Space(1)).Background(t.Surface).Border(1, t.Border).Radius(t.Radius)
		for _, option := range options {
			item := sel.Item(option.value).Padding(t.Space(1.75), t.Space(2.25)).Gap(t.Space(2)).Radius(t.Radius)
			if item.Highlighted() {
				item.Background(t.SurfaceHover)
			}
			item.Children(func() {
				ui.Text(c, option.label).SingleLine()
			})
		}
	})
	if *value != *applied {
		*applied = *value
		onChange()
	}
}

func (a *App) appearanceEditor(c *ui.Context, t *ui.Theme) {
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, locales.Appearance()).Bold()
		ui.Text(c, locales.AppearanceNote()).TextColor(t.TextMuted)
	})

	ui.Form(c, func() {
		ui.Field(c, locales.Theme(), func() {
			optionSelect(c, &a.ThemeMode, &a.themeApplied, t.Rem(12.9), themeOptions(), a.queueSaveSettings)
		})
		ui.Field(c, locales.Language(), func() {
			a.languageSelect(c)
		})
		ui.Field(c, locales.Font(), func() {
			ui.Select(c, &a.FontFamily, a.fontOptions()).Width(240).OnChange(a.queueSaveSettings)
		})
		ui.Field(c, locales.FontSize(), func() {
			ui.NumberInput(c, &a.FontSize, 8, 32, 1).Width(104).OnChange(a.queueSaveSettings)
		})
		ui.Field(c, locales.TabColor(), func() {
			ui.Row(c).Height(28).AlignItems(ui.Center).Children(func() {
				compactColorWell(c, &a.TabColor, locales.TabColor()).OnChange(a.queueSaveSettings)
			})
		})
		ui.Field(c, locales.RequestTabsField(), func() {
			optionSelect(c, &a.TabBarDisplay, &a.tabBarDisplayApplied, t.Rem(12.9), tabBarLayoutOptions(), a.queueSaveSettings)
		}).Description(locales.RequestTabsFieldHint())
		ui.Field(c, locales.BackgroundMode(), func() {
			optionSelect(c, &a.BackgroundMode, &a.backgroundModeApplied, t.Rem(12.9), backgroundModeOptions(), a.queueSaveSettings)
		})
		ui.Field(c, locales.BackgroundColor(), func() {
			ui.Row(c).Height(28).AlignItems(ui.Center).Children(func() {
				compactColorWell(c, &a.BackgroundColor, locales.BackgroundColor()).OnChange(a.queueSaveSettings)
			})
		})
		ui.Field(c, locales.BackgroundImageField(), func() {
			ui.Row(c).FillWidth().Gap(8).Children(func() {
				if a.backgroundBitmap != nil {
					ui.Image(c, a.backgroundBitmap).Size(72, 40).Fit(ui.Cover).Radius(6)
				} else {
					ui.Box(c).Size(72, 40).Background(t.Surface).Border(1, t.Border).Radius(6)
				}
				ui.Text(c, shortPath(a.BackgroundPath)).Grow(1).SingleLine()
				if ui.Button(c, "").Label(locales.ChooseImage()).Tooltip(locales.ChooseImageTooltip()).Children(func() {
					ui.Icon(c, iconUpload)
					ui.Text(c, locales.Choose()).SingleLine()
				}).Clicked() {
					go a.chooseBackgroundImage()
				}
				if ui.Button(c, "").Label(locales.ClearBackground()).Tooltip(locales.ClearBackgroundTooltip()).Disabled(a.BackgroundPath == "").Children(func() {
					ui.Icon(c, iconCancel)
					ui.Text(c, locales.Clear()).SingleLine()
				}).Clicked() {
					a.clearBackgroundImage()
				}
			})
		}).Description(locales.BackgroundImageHint())
		ui.Field(c, locales.CloseWindowAction(), func() {
			ui.Checkbox(c, &a.CloseToTray, locales.CloseToTrayLabel()).OnChange(a.queueSaveSettings)
		}).Description(locales.CloseToTrayHint())
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
