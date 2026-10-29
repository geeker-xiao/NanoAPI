package main

import (
	"os"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"nanoapi/internal/history"
)

func (a *App) historyTools(c *ui.Context, t *ui.Theme) {
	ui.Row(c).FillWidth().Gap(6).Children(func() {
		if ui.Button(c, "").Label("Export request history").Tooltip("Export request groups and records as JSON").Children(func() {
			ui.Icon(c, iconDownload)
			ui.Text(c, "Export").SingleLine()
		}).Clicked() {
			bundle := toHistoryBundle(a.Groups, a.History)
			go a.exportHistory(bundle)
		}
		if ui.Button(c, "").Label("Import request history").Tooltip("Import request groups and records from JSON").Children(func() {
			ui.Icon(c, iconUpload)
			ui.Text(c, "Import").SingleLine()
		}).Clicked() {
			existing := toHistoryBundle(a.Groups, a.History)
			go a.importHistory(existing)
		}
	})
}

func (a *App) exportHistory(bundle history.Bundle) {
	path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
		Title:       "Export request history",
		DefaultPath: "nanoapi-history.json",
		Filters:     []mygo.FileFilter{{Name: "JSON", Extensions: []string{"json"}}},
	})
	if err != nil {
		a.publish(func() { a.Error = err.Error() })
		return
	}
	if path == "" {
		return
	}
	data, err := history.Export(bundle)
	if err == nil {
		err = os.WriteFile(path, data, 0o600)
	}
	a.publish(func() {
		if err != nil {
			a.Error = err.Error()
			return
		}
		a.Error = ""
	})
}

func (a *App) importHistory(existing history.Bundle) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Title:   "Import request history",
		Filters: []mygo.FileFilter{{Name: "JSON", Extensions: []string{"json"}}},
	})
	if err != nil {
		a.publish(func() { a.Error = err.Error() })
		return
	}
	if len(paths) == 0 {
		return
	}
	data, err := os.ReadFile(paths[0])
	if err == nil {
		var imported history.Bundle
		imported, err = history.Import(data)
		if err == nil {
			merged := history.Merge(existing, imported)
			groups, items := fromHistoryBundle(merged)
			a.publish(func() {
				a.Groups = groups
				a.History = items
				a.ensureSelectedGroup()
				a.Error = ""
			})
			return
		}
	}
	a.publish(func() { a.Error = err.Error() })
}

func toHistoryBundle(groups []RequestGroup, items []HistoryItem) history.Bundle {
	bundle := history.Bundle{
		Groups:  make([]history.Group, len(groups)),
		Records: make([]history.Record, len(items)),
	}
	for i, group := range groups {
		bundle.Groups[i] = history.Group{ID: group.ID, Name: group.Name, Description: group.Description}
	}
	for i, item := range items {
		bundle.Records[i] = history.Record{
			ID:             item.ID,
			GroupID:        item.GroupID,
			Label:          item.Label,
			Request:        item.Request,
			Status:         item.Status,
			CreatedAt:      item.CreatedAt,
			RequestedAt:    item.RequestedAt,
			RespondedAt:    item.RespondedAt,
			DurationMillis: item.Duration.Milliseconds(),
		}
	}
	return bundle
}

func fromHistoryBundle(bundle history.Bundle) ([]RequestGroup, []HistoryItem) {
	groups := make([]RequestGroup, len(bundle.Groups))
	items := make([]HistoryItem, len(bundle.Records))
	for i, group := range bundle.Groups {
		groups[i] = RequestGroup{ID: group.ID, Name: group.Name, Description: group.Description}
	}
	for i, record := range bundle.Records {
		items[i] = HistoryItem{
			ID:          record.ID,
			GroupID:     record.GroupID,
			Label:       record.Label,
			Request:     record.Request,
			Status:      record.Status,
			CreatedAt:   record.CreatedAt,
			RequestedAt: record.RequestedAt,
			RespondedAt: record.RespondedAt,
			Duration:    time.Duration(record.DurationMillis) * time.Millisecond,
		}
	}
	return groups, items
}

func (a *App) ensureSelectedGroup() {
	if a.SelectedGroup == "all" || a.SelectedGroup == "ungrouped" {
		return
	}
	if _, ok := a.groupByID(a.SelectedGroup); !ok {
		a.SelectedGroup = "all"
	}
}
