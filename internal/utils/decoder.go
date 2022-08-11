package utils

import (
	"encoding/binary"
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// read a variable-length string in "Java modified UTF-8" encoding
func ReadString(r io.Reader) (string, error) {
	// step 1: read a uint16 representing the length
	strByteLen, err := ReadUint16(r)
	if err != nil {
		return "", errors.Wrap(err, "ReadString: failed to read expected string length uint16 with cause")
	}

	// step 2: read the expected string's buffer length in bytes from input stream
	strBuf := make([]byte, int(strByteLen))
	n, err := r.Read(strBuf)
	if err != nil && err != io.EOF {
		return "", errors.Wrapf(err, "ReadString: failed to read expected buffer of size %d (got %d) with cause", strByteLen, n)
	}
	if n < int(strByteLen) {
		return "", errors.Errorf("ReadString: only read %d bytes of expected %d", n, strByteLen)
	}

	// if no parse error, conserve possible reader io.EOF for caller
	s, sErr := GetString(strBuf)
	if sErr == nil {
		sErr = err
	}
	return s, sErr
}

func ReadByte(r io.Reader) (byte, error) {
	var arr [1]byte
	_, err := r.Read(arr[:])
	return arr[0], err
}

func ReadUint16(r io.Reader) (uint16, error) {
	var arr [2]byte
	n, err := r.Read(arr[:])
	if err != nil && err != io.EOF {
		return 0, err
	}
	if n != 2 {
		return 0, errors.Errorf("GetUint16: expected to read 2 bytes, got: %d", n)
	}

	// if no parse error, conserve possible reader io.EOF for caller
	return binary.BigEndian.Uint16(arr[:]), err

}

func ReadInt32(r io.Reader) (int32, error) {
	var arr [4]byte
	n, err := r.Read(arr[:])
	if err != nil && err != io.EOF {
		return 0, err
	}
	if n != 4 {
		return 0, errors.Errorf("GetUint16: expected to read 4 bytes, got: %d", n)
	}

	// if no parse error, conserve possible reader io.EOF for caller
	return int32(binary.BigEndian.Uint32(arr[:])), err

}

func ReadInt64(r io.Reader) (int64, error) {
	var arr [8]byte
	n, err := r.Read(arr[:])
	if err != nil && err != io.EOF {
		return 0, err
	}
	if n != 8 {
		return 0, errors.Errorf("GetUint16: expected to read 8 bytes, got: %d", n)
	}

	// if no parse error, conserve possible reader io.EOF for caller
	return int64(binary.BigEndian.Uint64(arr[:])), err
}

func ReadTimestamp(r io.Reader) (time.Time, error) {
	timeStr, err := ReadString(r)
	if err != nil && err != io.EOF {
		return time.Now().UTC(), errors.Wrap(err, "ReadTimestamp: failed to obtain string from decoder with cause")
	}

	// if no parse error, conserve possible reader io.EOF for caller
	t, tErr := GetTimestamp(timeStr)
	if tErr == nil {
		tErr = err
	}
	return t, tErr
}

func ReadVInt(r io.Reader) (int64, error) {
	var out int64
	var ndx = 0
	var buf [1]byte
	slc := buf[:]
	_, err := r.Read(slc)
	b := buf[0]

	for err == nil {
		val := b & 0x80
		offset := ndx * 7
		out |= (int64(val) << offset)
		ndx++

		if b == val {
			break
		}

		_, err = r.Read(slc)
		b = buf[0]
	}

	// a well-formed "out" can be returned with an io.EOF
	return out, err
}

// Decode Go UTF-8 string from fixed-length byte buffer in "Java modified UTF-8" encoding.
// See DataInput#readUTF: https://docs.oracle.com/javase/6/docs/api/java/io/DataInput.html#readUTF%28%29
func GetString(strBuf []byte) (string, error) {
	// parse "Java modified UTF-8" encoding from byte buffer of expected length
	ndx := 0
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
				return "", errors.Errorf("GetString: unexpected length 3 char at index %d of buffer of length %d", ndx, strByteLen)
			}
		} else if (strBuf[ndx] & 0xc0) == 0xc0 {
			// if the first byte begins with 1100 then there will be 2 bytes to decode
			if ndx+1 < strByteLen && legalTrailingByte(strBuf[ndx+1]) {
				ch := rune(((strBuf[ndx] & 0x1f) << 6) | (strBuf[ndx+1] & 0x3f))
				out = append(out, ch)
				ndx += 2
			} else {
				return "", errors.Errorf("GetString: unexpected length 2 char at index %d of buffer of length %d", ndx, strByteLen)
			}
		} else {
			// if an expected single-byte rune begins with
			// 1111xxxx or 10xxxxxx then it is invalid
			if (strBuf[ndx]&0xf0) == 0xf0 || (strBuf[ndx]&0x80) == 0x80 {
				return "", errors.Errorf("GetString: unexpected length 1 char at index %d of buffer of length %d", ndx, strByteLen)
			}

			// if the first byte does not begin with any of these
			// bit prefix sequences, it can be decoded alone
			ch := rune(strBuf[ndx])
			out = append(out, ch)
			ndx++
		}
	}

	return string(out), nil
}

func legalTrailingByte(b byte) bool {
	return (b & 0x80) != 0x80
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

// EXAMPLE INPUT: nexus.index.timestamp=20220801003457.736 +0000
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
