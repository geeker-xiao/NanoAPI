package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

type groupOption struct {
	id   string
	name string
}

func (a *App) groupSelect(c *ui.Context) {
	t := c.Theme()
	sel := ui.SelectBase(c, &a.SelectedGroup)
	sel.Trigger.FillWidth().Padding(7, 9).Gap(8).Border(1, t.Border).Radius(t.Radius).Children(func() {
		ui.Text(c, a.selectedGroupLabel()).Bold().SingleLine()
		ui.Spacer(c)
		ui.Icon(c, iconChevronDown).FontSize(14).TextColor(t.TextMuted)
	})
	sel.Popup(func(panel ui.Element) {
		panel.Padding(4).Background(t.Surface).Border(1, t.Border).Radius(t.Radius)
		for _, option := range a.groupOptions() {
			item := sel.Item(option.id).Padding(7, 9).Radius(t.Radius)
			if item.Highlighted() {
				item.Background(t.SurfaceHover)
			}
			item.Children(func() {
				ui.Text(c, option.name).SingleLine()
			})
		}
	})
}

func (a *App) groupOptions() []groupOption {
	options := []groupOption{{id: "all", name: "All groups"}, {id: "ungrouped", name: "Ungrouped"}}
	for _, group := range a.Groups {
		options = append(options, groupOption{id: group.ID, name: group.Name})
	}
	return options
}

func (a *App) selectedGroupLabel() string {
	for _, option := range a.groupOptions() {
		if option.id == a.SelectedGroup {
			return option.name
		}
	}
	return "All groups"
}

func (a *App) selectedGroup() (RequestGroup, bool) {
	for _, group := range a.Groups {
		if group.ID == a.SelectedGroup {
			return group, true
		}
	}
	return RequestGroup{}, false
}

func (a *App) groupTools(c *ui.Context, t *ui.Theme) {
	ui.Row(c).FillWidth().Gap(6).Children(func() {
		if ui.Button(c, "").Label("New request group").Tooltip("Create a request group").Children(func() {
			ui.Icon(c, iconPlus)
			ui.Text(c, "New").SingleLine()
		}).Clicked() {
			a.openGroupDialog("")
		}
		_, selected := a.selectedGroup()
		if ui.Button(c, "Edit").Label("Edit request group").Tooltip("Edit the selected group").Disabled(!selected).Clicked() {
			a.openGroupDialog(a.SelectedGroup)
		}
		if ui.Button(c, "").Label("Delete request group").Tooltip("Delete the selected group").Disabled(!selected).Children(func() {
			ui.Icon(c, iconTrash)
		}).Clicked() {
			a.GroupDeleteID = a.SelectedGroup
			a.GroupDeleteOpen = true
		}
	})
	if group, ok := a.selectedGroup(); ok && strings.TrimSpace(group.Description) != "" {
		ui.Text(c, group.Description).FontSize(11).TextColor(t.TextMuted)
	}
}

func (a *App) groupDialogs(c *ui.Context, t *ui.Theme) {
	ui.Modal(c, &a.GroupDialogOpen, func() {
		title := "New request group"
		if a.GroupDraftID != "" {
			title = "Edit request group"
		}
		ui.Column(c).Width(440).Padding(20).Gap(12).Background(t.Surface).Border(1, t.Border).Radius(12).Children(func() {
			ui.Text(c, title).FontSize(20).Bold()
			ui.Field(c, "Name", func() {
				ui.TextInput(c, &a.GroupDraftName).Placeholder("Group name").AutoFocus()
			})
			ui.Field(c, "Description", func() {
				ui.TextArea(c, &a.GroupDraftDescription).Placeholder("What belongs in this group?").Lines(2, 4)
			})
			if a.GroupDraftError != "" {
				ui.Text(c, a.GroupDraftError).TextColor(t.Danger)
			}
			ui.Row(c).FillWidth().Gap(8).Justify(ui.End).Children(func() {
				if ui.Button(c, "Cancel").Clicked() {
					a.GroupDialogOpen = false
				}
				if ui.PrimaryButton(c, "Save").Clicked() {
					a.saveGroupDraft()
				}
			})
		})
	})

	ui.Modal(c, &a.GroupDeleteOpen, func() {
		ui.Column(c).Width(400).Padding(20).Gap(12).Background(t.Surface).Border(1, t.Border).Radius(12).Children(func() {
			ui.Text(c, "Delete request group?").FontSize(20).Bold()
			ui.Text(c, "The group information will be removed. Requests in this group will move to Ungrouped.").TextColor(t.TextMuted)
			ui.Row(c).FillWidth().Gap(8).Justify(ui.End).Children(func() {
				if ui.Button(c, "Cancel").Clicked() {
					a.GroupDeleteOpen = false
					a.GroupDeleteID = ""
				}
				if ui.Button(c, "Delete").TextColor(t.Danger).Clicked() {
					a.deleteSelectedGroup()
				}
			})
		})
	})
}

func (a *App) openGroupDialog(groupID string) {
	a.GroupDraftID = groupID
	a.GroupDraftError = ""
	if groupID == "" {
		a.GroupDraftName = ""
		a.GroupDraftDescription = ""
	} else if group, ok := a.groupByID(groupID); ok {
		a.GroupDraftName = group.Name
		a.GroupDraftDescription = group.Description
	}
	a.GroupDialogOpen = true
}

func (a *App) saveGroupDraft() {
	name := strings.TrimSpace(a.GroupDraftName)
	if name == "" {
		a.GroupDraftError = "Group name is required."
		return
	}
	for _, group := range a.Groups {
		if group.ID != a.GroupDraftID && strings.EqualFold(group.Name, name) {
			a.GroupDraftError = "Another group already uses this name."
			return
		}
	}
	if a.GroupDraftID == "" {
		group := RequestGroup{
			ID:          fmt.Sprintf("group-%d", time.Now().UnixNano()),
			Name:        name,
			Description: strings.TrimSpace(a.GroupDraftDescription),
		}
		a.Groups = append(a.Groups, group)
		a.GroupDraftID = group.ID
		a.SelectedGroup = group.ID
	} else {
		for i := range a.Groups {
			if a.Groups[i].ID == a.GroupDraftID {
				a.Groups[i].Name = name
				a.Groups[i].Description = strings.TrimSpace(a.GroupDraftDescription)
				break
			}
		}
	}
	a.GroupDialogOpen = false
	a.GroupDraftError = ""
}

func (a *App) deleteSelectedGroup() {
	groupID := a.GroupDeleteID
	groups := a.Groups[:0]
	for _, group := range a.Groups {
		if group.ID != groupID {
			groups = append(groups, group)
		}
	}
	a.Groups = groups
	for i := range a.History {
		if a.History[i].GroupID == groupID {
			a.History[i].GroupID = ""
		}
	}
	if a.SelectedGroup == groupID {
		a.SelectedGroup = "all"
	}
	a.GroupDeleteOpen = false
	a.GroupDeleteID = ""
}

func (a *App) groupByID(id string) (RequestGroup, bool) {
	for _, group := range a.Groups {
		if group.ID == id {
			return group, true
		}
	}
	return RequestGroup{}, false
}

func (a *App) filteredHistory() []HistoryItem {
	query := strings.ToLower(strings.TrimSpace(a.HistorySearch))
	result := make([]HistoryItem, 0, len(a.History))
	for _, item := range a.History {
		if a.SelectedGroup == "ungrouped" && item.GroupID != "" {
			continue
		}
		if a.SelectedGroup != "all" && a.SelectedGroup != "ungrouped" && item.GroupID != a.SelectedGroup {
			continue
		}
		if query != "" && !historyMatches(item, query, a.Groups) {
			continue
		}
		result = append(result, item)
	}
	return result
}

func historyMatches(item HistoryItem, query string, groups []RequestGroup) bool {
	var builder strings.Builder
	builder.WriteString(item.Label)
	builder.WriteByte(' ')
	builder.WriteString(item.Request.Method)
	builder.WriteByte(' ')
	builder.WriteString(item.Request.URL)
	builder.WriteByte(' ')
	builder.WriteString(item.Status)
	builder.WriteByte(' ')
	builder.WriteString(item.Request.Body)
	for _, header := range item.Request.Headers {
		builder.WriteByte(' ')
		builder.WriteString(header.Key)
		builder.WriteByte(' ')
		builder.WriteString(header.Value)
	}
	for _, group := range groups {
		if group.ID == item.GroupID {
			builder.WriteByte(' ')
			builder.WriteString(group.Name)
			builder.WriteByte(' ')
			builder.WriteString(group.Description)
		}
	}
	return strings.Contains(strings.ToLower(builder.String()), query)
}
