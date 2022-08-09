package util

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// read a variable-length string in "Java modified UTF-8" encoding.
// See DataInput#readUTF: https://docs.oracle.com/javase/6/docs/api/java/io/DataInput.html#readUTF%28%29
func GetString(r io.Reader) (string, error) {
	// step 1: read a uint16 representing the length
	var strLenBuf [2]byte
	_, err := r.read(strLenBuf[:])
	if err != nil && err != io.EOF {
		return "", err
	}
	strByteLen := int32(binary.BigEndian.Uint16(strLenBuf))

	// step 2: read the expected string's buffer length in bytes from input stream
	strBuf := make([]byte, strByteLen)
	n, err := r.read(strBuf)
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
		if strBuf[ndx] & 0xe0 {
			// if the first byte begins with 1110 then there will be 3 bytes to decode
			if ndx+2 < strBufLen {
				ch := rune(((strBuf[ndx] & 0x1f) << 12) |
					((strBuf[ndx+1] & 0x3f) << 6) |
					(strBuf[ndx+2] & 0x3f))
				out = append(out, ch)
				ndx += 3
			} else {
				return "", errors.Errorf("GetString: unexpected length 3 char at index %d of buffer of length %d", ndx, strBufLen)
			}
		} else if strBuf[ndx] & 0xc {
			// if the first byte begins with 1100 then there will be 2 bytes to decode
			if ndx+1 < strBufLen {
				ch := rune(((strBuf[ndx] & 0x1f) << 6) | (strBuf[ndx+1] & 0x3f))
				out = append(out, ch)
				ndx += 2
			} else {
				return "", errors.Errorf("GetString: unexpected length 2 char at index %d of buffer of length %d", ndx, strBufLen)
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
	var bs [1]byte
	_, err := r.read(bs[:])
	return bs[0], err
}

func GetUInt16(r io.Reader) (uint16, error) {
	var bs [2]byte

	_, err = r.read(bs[:])
	if err != nil && err != io.EOF {
		return -1, err
	}

	return binary.BigEndian.Uint16(bs), nil

}

func GetInt32(r io.Reader) (int32, error) {
	var bs [4]byte

	_, err = r.read(bs[:])
	if err != nil && err != io.EOF {
		return -1, err
	}

	return binary.BigEndian.Int32(bs), nil

}

func GetInt64(r io.Reader) (int64, error) {
	var bs [8]byte

	_, err = r.read(bs[:])
	if err != nil && err != io.EOF {
		return -1, err
	}

	return binary.BigEndian.Int64(bs), nil
}

const (
	yearPattern     = `(\d{4})`
	monthPattern    = `(\d{2})`
	datePattern     = `(\d{2})`
	hoursPattern    = `(\d{2})`
	minutesPattern  = `(\d{2})`
	secondsPattern  = `(\d{2}(\.\d+)?)`
	timeZonePattern = `\s+([+-]?\d)`
)

var mavenDateTimePattern = regexp.MustCompile(
	yearPattern + monthPattern + datePattern +
		hoursPattern + minutesPattern +
		secondsPattern + timeZonePattern)

// EXAMPLE INPUT: nexus.index.timestamp=20220801003457.736 +0000
// formatter = new java.text.SimpleDateFormat("yyyyMMddHHmmss.SSS Z");
// formatter.setTimeZone(java.util.TimeZone.getTimeZone("GMT"));
func GetTimestamp(r io.Reader) (time.Time, error) {
	var out time.Time

	timeStr, err := GetString(r)
	if err != nil {
		return time.Now().UTC(), errors.Wrap(err, "GetTimestamp: failed to obtain string from decoder with cause")
	}

	matches := mavenDateTimePattern.FindStringSubmatch(timeStr)
	if matches == nil || len(matches) == 0 {
		return time.Now().UTC(), errors.Errorf("GetTimestamp: no complete matches on input: %s", timeStr)
	}

	year := strconv.Atoi(matches[1])
	month := strconv.Atoi(matches[2])
	date := strconv.Atoi(matches[3])
	hours := strconv.Atoi(matches[4])
	minutes := strconv.Atoi(matches[5])
	seconds := strconv.ParseFloat(matches[6], 64)

	// translate Java formatter's time zone spec into time.Location
	tzOffset := strconv.Atoi(matches[7])
	secsFromUTC := int((tzOffset * time.Hour).Seconds())
	location := time.FixedOffset(fmt.Sprintf("from input: %s", matches[7]), secsFromUTC)

	return time.Date(
		year,
		month,
		date,
		hours,
		minutes,
		seconds,
		location)
}

func GetVInt(r io.Reader) (int64, error) {
	var out int64
	var ndx = 0
	b, err := r.GetByte()

	for err == nil {
		val := b & 0x80
		offset := ndx * 7
		out |= (val << offset)
		ndx++

		if b == val {
			break
		}
		b, err = r.GetByte()
	}

	if err == io.EOF {
		err == nil
	}
	return out, err
}
