package util

import (
	"encoding/binary"
	"fmt"
	"io"
	//"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// read a variable-length string in "Java modified UTF-8" encoding.
// See DataInput#readUTF: https://docs.oracle.com/javase/6/docs/api/java/io/DataInput.html#readUTF%28%29
func GetString(r io.Reader) (string, error) {
	// step 1: read a uint16 representing the length
	var strLenBuf [2]byte
	_, err := r.Read(strLenBuf[:])
	if err != nil && err != io.EOF {
		return "", err
	}
	strByteLen := int(binary.BigEndian.Uint16(strLenBuf[:]))

	// step 2: read the expected string's buffer length in bytes from input stream
	strBuf := make([]byte, strByteLen)
	n, err := r.Read(strBuf)
	if err != nil && err != io.EOF {
		return "", err
	}
	if n < strByteLen {
		return "", errors.Errorf("GetString: only read %d bytes of expected %d", n, strByteLen)
	}

	// step 3: parse "Java modified UTF-8" encoding from buffer
	ndx := 0
	var out []rune
	for ndx < strByteLen {
		if (strBuf[ndx] & 0xe0) != 0 {
			// if the first byte begins with 1110 then there will be 3 bytes to decode
			if ndx+2 < strByteLen {
				ch := rune(((strBuf[ndx] & 0x1f) << 12) |
					((strBuf[ndx+1] & 0x3f) << 6) |
					(strBuf[ndx+2] & 0x3f))
				out = append(out, ch)
				ndx += 3
			} else {
				return "", errors.Errorf("GetString: unexpected length 3 char at index %d of buffer of length %d", ndx, strByteLen)
			}
		} else if (strBuf[ndx] & 0xc) != 0 {
			// if the first byte begins with 1100 then there will be 2 bytes to decode
			if ndx+1 < strByteLen {
				ch := rune(((strBuf[ndx] & 0x1f) << 6) | (strBuf[ndx+1] & 0x3f))
				out = append(out, ch)
				ndx += 2
			} else {
				return "", errors.Errorf("GetString: unexpected length 2 char at index %d of buffer of length %d", ndx, strByteLen)
			}
		} else { // TODO(eli): check docs, validate this is correct in all cases :)
			// otherwise, just decode the byte we have
			ch := rune(strBuf[ndx])
			out = append(out, ch)
			ndx++
		}
	}

	return string(out), nil
}

func GetByte(r io.Reader) (byte, error) {
	var arr [1]byte
	_, err := r.Read(arr[:])
	return arr[0], err
}

func GetUInt16(r io.Reader) (uint16, error) {
	var arr [2]byte
	_, err := r.Read(arr[:])
	if err != nil && err != io.EOF {
		return 0, err
	}

	return binary.BigEndian.Uint16(arr[:]), nil

}

func GetInt32(r io.Reader) (int32, error) {
	var arr [4]byte
	_, err := r.Read(arr[:])
	if err != nil && err != io.EOF {
		return 0, err
	}

	return int32(binary.BigEndian.Uint32(arr[:])), nil

}

func GetInt64(r io.Reader) (int64, error) {
	var arr [8]byte
	_, err := r.Read(arr[:])
	if err != nil && err != io.EOF {
		return 0, err
	}

	return int64(binary.BigEndian.Uint64(arr[:])), nil
}

const (
	yearPattern     = `(\d{4})`
	monthPattern    = `(\d{2})`
	datePattern     = `(\d{2})`
	hoursPattern    = `(\d{2})`
	minutesPattern  = `(\d{2})`
	secondsPattern  = `(\d{2})`
	millisPattern   = `\.(\d{3})`
	timeZonePattern = `\s+([+-]?\d)`
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
func GetTimestamp(r io.Reader) (time.Time, error) {
	timeStr, err := GetString(r)
	if err != nil {
		return time.Now().UTC(), errors.Wrap(err, "GetTimestamp: failed to obtain string from decoder with cause")
	}

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
	tzOffset, _ := strconv.Atoi(matches[7])
	secsFromUTC := int((time.Duration(tzOffset) * time.Hour).Seconds())
	location := time.FixedZone(fmt.Sprintf("from input: %s", matches[7]), secsFromUTC)

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

func GetVInt(r io.Reader) (int64, error) {
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

	if err == io.EOF {
		err = nil
	}
	return out, err
}

var linesPattern = regexp.MustCompile(`\r\n`)

func GetProperties(r io.Reader) (map[string]string, error) {
	raw, err := GetString(r) //ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "GetProperties: failed to read raw data from input with cause")
	}

	out := map[string]string{}
	for ndx, line := range linesPattern.Split(raw, -1) {
		key, value, found := strings.Cut(line, "=")
		if !found {
			return nil, errors.Errorf("GetProperties: line %d failed to parse into key and value: %s", ndx, line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		out[key] = value
	}

	return out, nil
}
