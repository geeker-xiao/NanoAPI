package httpclient

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func TestResolveValueFixed(t *testing.T) {
	value := ResolveValue(KeyValue{ValueMode: ValueModeFixed, Value: "fixed-value"}, time.Now())
	if value != "fixed-value" {
		t.Fatalf("value = %q", value)
	}
}

func TestGenerateTimestampValues(t *testing.T) {
	now := time.UnixMilli(1700000000123)
	if got := generateValue(RandomTimestampMS, "", now, zeroReader{}); got != "1700000000123" {
		t.Fatalf("timestamp ms = %q", got)
	}
	if got := generateValue(RandomTimestampSec, "", now, zeroReader{}); got != "1700000000" {
		t.Fatalf("timestamp sec = %q", got)
	}
}

func TestGenerateUUIDAndULID(t *testing.T) {
	now := time.UnixMilli(1700000000123)
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[47][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	uuid := generateValue(RandomUUIDv4, "", now, zeroReader{})
	if !uuidPattern.MatchString(uuid) {
		t.Fatalf("uuid v4 = %q", uuid)
	}
	uuid = generateValue(RandomUUIDv7, "", now, zeroReader{})
	if !uuidPattern.MatchString(uuid) {
		t.Fatalf("uuid v7 = %q", uuid)
	}
	ulid := generateValue(RandomULID, "", now, zeroReader{})
	if len(ulid) != 26 || strings.Trim(ulid, "0123456789ABCDEFGHJKMNPQRSTVWXYZ") != "" {
		t.Fatalf("ulid = %q", ulid)
	}
}

func TestGenerateTemplateConcatenation(t *testing.T) {
	now := time.UnixMilli(1700000000123)
	got := generateValue(RandomTemplate, "req-{{timestamp}}-{{random:4}}-{{date}}", now, zeroReader{})
	want := "req-1700000000123-AAAA-2023-11-14"
	if got != want {
		t.Fatalf("template = %q, want %q", got, want)
	}
}

func TestRandomStringAndTimePlusRandom(t *testing.T) {
	now := time.UnixMilli(1700000000123)
	if got := generateValue(RandomString, "12", now, zeroReader{}); len(got) != 12 {
		t.Fatalf("random string = %q", got)
	}
	if got := generateValue(RandomTimeAndString, "trace", now, zeroReader{}); got != "trace-1700000000123-AAAAAA" {
		t.Fatalf("time + random = %q", got)
	}
}
