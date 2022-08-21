package utils

import (
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Decode Go UTF-8 string from fixed-length byte buffer in "Java modified UTF-8" encoding.
// See DataInput#readUTF: https://docs.oracle.com/javase/6/docs/api/java/io/DataInput.html#readUTF%28%29
func GetString(strBuf []byte) (string, error) {
	// parse "Java modified UTF-8" encoding from byte buffer of expected length
	var ndx int
	strByteLen := len(strBuf)
	var out []rune
	for ndx < strByteLen {
		if (strBuf[ndx] & 0xe0) == 0xe0 {
			// if the first byte begins with 1110 then there will be 3 bytes to decode
			if ndx+2 < strByteLen && legalTrailingByte(strBuf[ndx+1]) && legalTrailingByte(strBuf[ndx+2]) {
				ch := rune(((strBuf[ndx] & 0x1f) << 12) |
					((strBuf[ndx+1] & 0x3f) << 6) |
					(strBuf[ndx+2] & 0x3f))
				out = append(out, ch)
				ndx += 3
			} else {
				if strBuf[ndx+1] == 0 || strBuf[ndx+2] == 0 {
					return "", errors.Errorf(
						"GetString: unexpected 0 bytes after index %d of buffer of length %d: %v",
						ndx, strByteLen, strBuf[ndx:])
				}
				return "", errors.Errorf(
					"GetString: unexpected length 3 char at index %d of buffer of length %d: %s",
					ndx, strByteLen, string(strBuf[ndx:ndx+3]))
			}
		} else if (strBuf[ndx] & 0xc0) == 0xc0 {
			// if the first byte begins with 1100 then there will be 2 bytes to decode
			if ndx+1 < strByteLen && legalTrailingByte(strBuf[ndx+1]) {
				ch := rune(((strBuf[ndx] & 0x1f) << 6) | (strBuf[ndx+1] & 0x3f))
				out = append(out, ch)
				ndx += 2
			} else {
				if strBuf[ndx+1] == 0 {
					return "", errors.Errorf(
						"GetString: unexpected 0 bytes after index %d of buffer of length %d: %v",
						ndx, strByteLen, strBuf[ndx:])
				}
				return "", errors.Errorf(
					"GetString: unexpected length 2 char at index %d of buffer of length %d: %s",
					ndx, strByteLen, string(strBuf[ndx:ndx+2]))
			}
		} else {
			// if an expected single-byte rune begins with
			// 1111xxxx or 10xxxxxx then it is invalid.
			// if the first byte of a group matches the bit pattern
			// 0xxxxxxx then the group consists of just that byte
			if (strBuf[ndx]&0xf0) == 0xf0 || (strBuf[ndx]&0x80) == 0x80 {
				return "", errors.Errorf(
					"GetString: invalid significant bits on expected char of length 1 at index %d of buffer of length %d: %s",
					ndx, strByteLen, string(strBuf[ndx:ndx+1]))
			}
			// if this is a zero byte, something is wrong
			if strBuf[ndx] == 0 {
				return "", errors.Errorf(
					"GetString: unexpected 0 bytes after index %d of buffer of length %d: %v",
					ndx, strByteLen, strBuf[ndx:])
			}

			// this 1-byte character is well-formed
			ch := rune(strBuf[ndx])
			out = append(out, ch)
			ndx++
		}
	}

	return string(out), nil
}

func legalTrailingByte(b byte) bool {
	return (b & 0x80) == 0x80
}

// GetTimestamp - example input:
// nexus.index.timestamp=20220801003457.736 +0000
//
// Mimics functionality of:
// formatter = new java.text.SimpleDateFormat("yyyyMMddHHmmss.SSS Z");
// formatter.setTimeZone(java.util.TimeZone.getTimeZone("GMT"));
func GetTimestamp(timeStr string) (time.Time, error) {
	matches := mavenDateTimePattern.FindStringSubmatch(timeStr)
	if matches == nil || len(matches) == 0 {
		return time.Now().UTC(), errors.Errorf("GetTimestamp: no complete matches on input: %s", timeStr)
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	date, _ := strconv.Atoi(matches[3])
	hours, _ := strconv.Atoi(matches[4])
	minutes, _ := strconv.Atoi(matches[5])
	seconds, _ := strconv.Atoi(matches[6])

	millis, err := strconv.Atoi(matches[7])
	if err != nil {
		return time.Now().UTC(), errors.Wrapf(err, "GetTimestamp: failed to parse fractional seconds from: %s with cause", timeStr)
	}
	nanos := millis * 1000000

	// translate Java formatter's time zone spec into time.Location
	rawTZ := matches[8]
	isNegOffset := rawTZ[0] == '-'
	tzHours, _ := strconv.Atoi(rawTZ[1:3])
	tzMinutes, _ := strconv.Atoi(rawTZ[3:])

	secsFromUTC := (tzHours * 60 * 60) + (tzMinutes * 60)
	if isNegOffset {
		secsFromUTC = -secsFromUTC
	}
	location := time.FixedZone("TZ offset", secsFromUTC)

	return time.Date(
		year,
		time.Month(month),
		date,
		hours,
		minutes,
		seconds,
		nanos,
		location), nil
}

const (
	yearPattern     = `(\d{4})`
	monthPattern    = `(\d{2})`
	datePattern     = `(\d{2})`
	hoursPattern    = `(\d{2})`
	minutesPattern  = `(\d{2})`
	secondsPattern  = `(\d{2})`
	millisPattern   = `\.(\d{3})`
	timeZonePattern = `\s+([+-]\d{4})`
)

var mavenDateTimePattern = regexp.MustCompile(
	yearPattern +
		monthPattern +
		datePattern +
		hoursPattern +
		minutesPattern +
		secondsPattern +
		millisPattern +
		timeZonePattern)
