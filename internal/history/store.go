package history

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"nanoapi/internal/httpclient"
)

const (
	Version     = 1
	MaxRecords  = 1000
	MaxGroups   = 200
	MaxFileSize = 10 << 20
)

var (
	ErrInvalidFile = errors.New("history file is invalid")
	ErrTooLarge    = errors.New("history file is too large")
)

type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Record struct {
	ID             string             `json:"id,omitempty"`
	GroupID        string             `json:"groupId,omitempty"`
	Label          string             `json:"label"`
	Request        httpclient.Request `json:"request"`
	Status         string             `json:"status,omitempty"`
	CreatedAt      time.Time          `json:"createdAt,omitempty"`
	RequestedAt    time.Time          `json:"requestedAt,omitempty"`
	RespondedAt    time.Time          `json:"respondedAt,omitempty"`
	DurationMillis int64              `json:"durationMillis,omitempty"`
}

type Bundle struct {
	Groups  []Group  `json:"groups,omitempty"`
	Records []Record `json:"records"`
}

type File struct {
	Version    int       `json:"version"`
	ExportedAt time.Time `json:"exportedAt"`
	Bundle
}

func Export(bundle Bundle) ([]byte, error) {
	if len(bundle.Records) > MaxRecords {
		bundle.Records = bundle.Records[:MaxRecords]
	}
	if len(bundle.Groups) > MaxGroups {
		bundle.Groups = bundle.Groups[:MaxGroups]
	}
	return json.MarshalIndent(File{Version: Version, ExportedAt: time.Now().UTC(), Bundle: bundle}, "", "  ")
}

func Import(data []byte) (Bundle, error) {
	if len(data) > MaxFileSize {
		return Bundle{}, ErrTooLarge
	}
	var file File
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return Bundle{}, fmt.Errorf("%w: %v", ErrInvalidFile, err)
	}
	if file.Version != Version {
		return Bundle{}, fmt.Errorf("%w: unsupported version %d", ErrInvalidFile, file.Version)
	}
	if len(file.Records) > MaxRecords || len(file.Groups) > MaxGroups {
		return Bundle{}, fmt.Errorf("%w: too many records or groups", ErrInvalidFile)
	}
	groupIDs := make(map[string]bool, len(file.Groups))
	for _, group := range file.Groups {
		if group.ID == "" || group.Name == "" || groupIDs[group.ID] {
			return Bundle{}, fmt.Errorf("%w: invalid or duplicate group", ErrInvalidFile)
		}
		groupIDs[group.ID] = true
	}
	for _, record := range file.Records {
		if record.Request.Method == "" || record.Request.URL == "" {
			return Bundle{}, fmt.Errorf("%w: request is missing method or URL", ErrInvalidFile)
		}
		if record.GroupID != "" && !groupIDs[record.GroupID] {
			return Bundle{}, fmt.Errorf("%w: request references an unknown group", ErrInvalidFile)
		}
	}
	return file.Bundle, nil
}

func Merge(existing, imported Bundle) Bundle {
	result := Bundle{Groups: mergeGroups(existing.Groups, imported.Groups), Records: mergeRecords(existing.Records, imported.Records)}
	return result
}

func mergeGroups(existing, imported []Group) []Group {
	byID := make(map[string]Group, len(existing)+len(imported))
	order := make([]string, 0, len(existing)+len(imported))
	for _, groups := range [][]Group{existing, imported} {
		for _, group := range groups {
			if group.ID == "" {
				continue
			}
			if _, ok := byID[group.ID]; !ok {
				order = append(order, group.ID)
			}
			byID[group.ID] = group
		}
	}
	result := make([]Group, 0, len(order))
	for _, id := range order {
		result = append(result, byID[id])
		if len(result) >= MaxGroups {
			break
		}
	}
	return result
}

func mergeRecords(existing, imported []Record) []Record {
	seen := make(map[string]bool, len(existing)+len(imported))
	result := make([]Record, 0, len(existing)+len(imported))
	for _, records := range [][]Record{imported, existing} {
		for _, record := range records {
			key := recordKey(record)
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, record)
			if len(result) >= MaxRecords {
				return result
			}
		}
	}
	return result
}

func recordKey(record Record) string {
	request := record.Request
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", record.GroupID, request.Method, request.URL, request.BodyMode, request.Body)
}
