package main

import (
	"encoding/json"
	"fmt"
	"github.com/egoist/mygo"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/ui"
)

type settingsData struct {
	UserAgent       string  `json:"userAgent"`
	ThemeMode       string  `json:"themeMode"`
	FontFamily      string  `json:"fontFamily"`
	FontSize        float64 `json:"fontSize"`
	BackgroundMode  string  `json:"backgroundMode"`
	BackgroundPath  string  `json:"backgroundPath"`
	BackgroundColor string  `json:"backgroundColor"`
	TabColor        string  `json:"tabColor"`
	TabBarDisplay   string  `json:"tabBarDisplay"`
	CloseToTray     bool    `json:"closeToTray"`
}

func (a *App) settingsSnapshot() settingsData {
	return settingsData{
		UserAgent:       a.tab().UserAgent,
		ThemeMode:       a.ThemeMode,
		FontFamily:      a.FontFamily,
		FontSize:        a.FontSize,
		BackgroundMode:  a.BackgroundMode,
		BackgroundPath:  a.BackgroundPath,
		BackgroundColor: hexColor(a.BackgroundColor),
		TabColor:        hexColor(a.TabColor),
		TabBarDisplay:   a.TabBarDisplay,
		CloseToTray:     a.CloseToTray,
	}
}

func (a *App) loadSettings() {
	path, err := settingsFile()
	if err != nil {
		return
	}
	data, err := os.ReadFile(path)
	legacy := false
	if err != nil {
		legacyPath := filepath.Join(filepath.Dir(filepath.Dir(path)), "mygo-api-studio", "settings.json")
		data, err = os.ReadFile(legacyPath)
		legacy = err == nil
	}
	if err != nil {
		return
	}
	var saved settingsData
	if err := json.Unmarshal(data, &saved); err != nil {
		return
	}
	if saved.UserAgent == "MyGo API Studio/1.0" {
		saved.UserAgent = "NanoAPI/1.0"
	}
	if saved.UserAgent != "" {
		a.tab().UserAgent = saved.UserAgent
	}
	if saved.ThemeMode == "System" || saved.ThemeMode == "Light" || saved.ThemeMode == "Dark" {
		a.ThemeMode = saved.ThemeMode
	}
	if saved.FontFamily != "" {
		a.FontFamily = saved.FontFamily
	}
	if saved.FontSize >= 8 && saved.FontSize <= 32 {
		a.FontSize = saved.FontSize
	}
	if saved.BackgroundMode == "Solid color" || saved.BackgroundMode == "Image" {
		a.BackgroundMode = saved.BackgroundMode
	}
	if saved.BackgroundPath != "" {
		a.BackgroundPath = saved.BackgroundPath
	}
	if strings.HasPrefix(saved.TabColor, "#") {
		a.TabColor = ui.Hex(saved.TabColor)
	}
	switch saved.TabBarDisplay {
	case tabBarMultiLine, tabBarSingleLine:
		a.TabBarDisplay = saved.TabBarDisplay
	}
	a.CloseToTray = saved.CloseToTray
	if strings.HasPrefix(saved.BackgroundColor, "#") {
		a.BackgroundColor = ui.Hex(saved.BackgroundColor)
	}
	if legacy {
		_ = writeSettings(a.settingsSnapshot())
	}
}

func (a *App) queueSaveSettings() {
	saved := a.settingsSnapshot()
	go func() {
		if err := writeSettings(saved); err != nil {
			a.publish(func() { a.Error = "Could not save appearance settings: " + err.Error() })
		}
	}()
}

func writeSettings(saved settingsData) error {
	path, err := settingsFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func settingsFile() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "nanoapi", "settings.json"), nil
}

func (a *App) loadBackgroundImage() {
	if a.BackgroundMode != "Image" || a.BackgroundPath == "" {
		return
	}
	data, err := os.ReadFile(a.BackgroundPath)
	if err != nil {
		return
	}
	bitmap, err := ui.DecodeBitmap(data)
	if err == nil {
		a.backgroundBitmap = bitmap
	}
}

func (a *App) chooseBackgroundImage() {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Title: "Choose a background image", Filters: []mygo.FileFilter{{Name: "Images", Extensions: []string{"png", "jpg", "jpeg", "gif", "webp", "bmp"}}}})
	if err != nil {
		a.publish(func() { a.Error = err.Error() })
		return
	}
	if len(paths) == 0 {
		return
	}
	path := paths[0]
	data, err := os.ReadFile(path)
	if err != nil {
		a.publish(func() { a.Error = err.Error() })
		return
	}
	a.publish(func() {
		bitmap, err := ui.DecodeBitmap(data)
		if err != nil {
			a.Error = err.Error()
			return
		}
		a.backgroundBitmap = bitmap
		a.BackgroundPath = path
		a.BackgroundMode = "Image"
		a.Error = ""
		a.queueSaveSettings()
	})
}

func (a *App) clearBackgroundImage() {
	a.backgroundBitmap = nil
	a.BackgroundPath = ""
	a.BackgroundMode = "Solid color"
	a.Error = ""
	a.queueSaveSettings()
}

func hexColor(color ui.Color) string {
	return fmt.Sprintf("#%02x%02x%02x", color.R, color.G, color.B)
}
