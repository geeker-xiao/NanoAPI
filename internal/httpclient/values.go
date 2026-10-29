package httpclient

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	ValueModeFixed  = "Fixed"
	ValueModeRandom = "Random"

	RandomUUIDv4        = "UUID v4"
	RandomUUIDv7        = "UUID v7"
	RandomULID          = "ULID"
	RandomString        = "Random string"
	RandomInteger       = "Random integer"
	RandomTimestampMS   = "Timestamp milliseconds"
	RandomTimestampSec  = "Timestamp seconds"
	RandomISOTime       = "ISO 8601"
	RandomTimeAndString = "Time + random"
	RandomTemplate      = "Template / concatenation"
)

var templateToken = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)(?::([0-9]+))?\s*\}\}`)

// ResolveValue returns the fixed value, or generates a random value from the
// item's algorithm and pattern.
func ResolveValue(item KeyValue, now time.Time) string {
	if !strings.EqualFold(strings.TrimSpace(item.ValueMode), ValueModeRandom) {
		return item.Value
	}
	mode := strings.TrimSpace(item.RandomMode)
	if mode == "" {
		mode = RandomUUIDv4
	}
	return generateValue(mode, item.Pattern, now, rand.Reader)
}

func generateValue(mode, pattern string, now time.Time, random io.Reader) string {
	switch mode {
	case RandomUUIDv4, "UUID", "uuid":
		return uuidV4(random)
	case RandomUUIDv7:
		return uuidV7(now, random)
	case RandomULID:
		return ulid(now, random)
	case RandomString:
		length := positiveInt(pattern, 8, 128)
		return randomString(random, length)
	case RandomInteger:
		return randomInteger(random, now)
	case RandomTimestampMS, "Timestamp", "timestamp":
		return strconv.FormatInt(now.UnixMilli(), 10)
	case RandomTimestampSec, "Unix timestamp":
		return strconv.FormatInt(now.Unix(), 10)
	case RandomISOTime, "ISO timestamp":
		return now.UTC().Format(time.RFC3339Nano)
	case RandomTimeAndString:
		value := strconv.FormatInt(now.UnixMilli(), 10) + "-" + randomString(random, 6)
		if strings.TrimSpace(pattern) != "" {
			value = strings.TrimSpace(pattern) + "-" + value
		}
		return value
	case RandomTemplate, "Template":
		return expandTemplate(pattern, now, random)
	default:
		return expandTemplate(pattern, now, random)
	}
}

func expandTemplate(pattern string, now time.Time, random io.Reader) string {
	if strings.TrimSpace(pattern) == "" {
		return uuidV4(random)
	}
	return templateToken.ReplaceAllStringFunc(pattern, func(token string) string {
		match := templateToken.FindStringSubmatch(token)
		if len(match) < 2 {
			return token
		}
		name := strings.ToLower(match[1])
		length := 8
		if len(match) > 2 && match[2] != "" {
			length = positiveInt(match[2], 1, 128)
		}
		switch name {
		case "uuid", "uuidv4":
			return uuidV4(random)
		case "uuidv7":
			return uuidV7(now, random)
		case "ulid":
			return ulid(now, random)
		case "random", "random_string", "randomstring":
			return randomString(random, length)
		case "random_int", "randomint":
			return randomInteger(random, now)
		case "timestamp", "timestamp_ms", "unix_ms":
			return strconv.FormatInt(now.UnixMilli(), 10)
		case "timestamp_seconds", "timestamp_s", "unix", "unix_seconds":
			return strconv.FormatInt(now.Unix(), 10)
		case "nano", "timestamp_nano":
			return strconv.FormatInt(now.UnixNano(), 10)
		case "iso8601", "iso", "time":
			return now.UTC().Format(time.RFC3339Nano)
		case "date":
			return now.UTC().Format("2006-01-02")
		default:
			return token
		}
	})
}

func uuidV4(random io.Reader) string {
	value := randomBytes(random, 16)
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return formatUUID(value)
}

func uuidV7(now time.Time, random io.Reader) string {
	value := randomBytes(random, 16)
	millis := uint64(now.UnixMilli())
	value[0] = byte(millis >> 40)
	value[1] = byte(millis >> 32)
	value[2] = byte(millis >> 24)
	value[3] = byte(millis >> 16)
	value[4] = byte(millis >> 8)
	value[5] = byte(millis)
	value[6] = (value[6] & 0x0f) | 0x70
	value[8] = (value[8] & 0x3f) | 0x80
	return formatUUID(value)
}

func formatUUID(value []byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16])
}

func ulid(now time.Time, random io.Reader) string {
	value := randomBytes(random, 16)
	millis := uint64(now.UnixMilli())
	value[0] = byte(millis >> 40)
	value[1] = byte(millis >> 32)
	value[2] = byte(millis >> 24)
	value[3] = byte(millis >> 16)
	value[4] = byte(millis >> 8)
	value[5] = byte(millis)

	const alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	number := new(big.Int).SetBytes(value)
	result := make([]byte, 26)
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = alphabet[number.Int64()&31]
		number.Rsh(number, 5)
	}
	return string(result)
}

const randomAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func randomString(random io.Reader, length int) string {
	bytes := randomBytes(random, length)
	result := make([]byte, length)
	for i := range result {
		result[i] = randomAlphabet[int(bytes[i])%len(randomAlphabet)]
	}
	return string(result)
}

func randomInteger(random io.Reader, now time.Time) string {
	number, err := rand.Int(random, big.NewInt(1_000_000_000))
	if err != nil {
		return strconv.FormatInt(now.UnixNano()%1_000_000_000, 10)
	}
	return number.String()
}

func randomBytes(random io.Reader, length int) []byte {
	value := make([]byte, length)
	if _, err := io.ReadFull(random, value); err != nil {
		binary.BigEndian.PutUint64(value, uint64(time.Now().UnixNano()))
	}
	return value
}

func positiveInt(value string, fallback, max int) int {
	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || number < 1 {
		return fallback
	}
	if number > max {
		return max
	}
	return number
}
