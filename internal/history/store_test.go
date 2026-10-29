package history

import (
	"strings"
	"testing"
	"time"

	"nanoapi/internal/httpclient"
)

func TestExportImportRoundTrip(t *testing.T) {
	bundle := Bundle{
		Groups: []Group{{ID: "g1", Name: "Production", Description: "Live APIs"}},
		Records: []Record{{ID: "r1", GroupID: "g1", Label: "GET example", Status: "200 OK", RequestedAt: time.UnixMilli(1700000000123), RespondedAt: time.UnixMilli(1700000000456), DurationMillis: 333, Request: httpclient.Request{
			Method: "GET", URL: "https://example.com", UserAgent: "tester/1.0",
		}}},
	}
	data, err := Export(bundle)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Import(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Records) != 1 || got.Records[0].Request.UserAgent != "tester/1.0" || got.Records[0].DurationMillis != 333 {
		t.Fatalf("records = %#v", got.Records)
	}
	if len(got.Groups) != 1 || got.Groups[0].Name != "Production" {
		t.Fatalf("groups = %#v", got.Groups)
	}
}

func TestImportRejectsUnknownVersion(t *testing.T) {
	_, err := Import([]byte(`{"version":2,"records":[]}`))
	if err == nil || !strings.Contains(err.Error(), "unsupported version") {
		t.Fatalf("err = %v", err)
	}
}

func TestImportRejectsUnknownGroup(t *testing.T) {
	_, err := Import([]byte(`{"version":1,"records":[{"request":{"method":"GET","url":"https://example.com"},"groupId":"missing"}]}`))
	if err == nil || !strings.Contains(err.Error(), "unknown group") {
		t.Fatalf("err = %v", err)
	}
}

func TestMergeDeduplicatesRecordsAndGroups(t *testing.T) {
	group := Group{ID: "g1", Name: "API"}
	record := Record{GroupID: "g1", Request: httpclient.Request{Method: "GET", URL: "https://example.com"}}
	got := Merge(Bundle{Groups: []Group{group}, Records: []Record{record}}, Bundle{Groups: []Group{group}, Records: []Record{record}})
	if len(got.Records) != 1 || len(got.Groups) != 1 {
		t.Fatalf("merge = %#v", got)
	}
}
