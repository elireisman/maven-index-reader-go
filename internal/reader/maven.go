package reader

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

type Maven interface {
	GetString() (string, error)

	GetBool() (bool, error)

	GetByte() (byte, error)

	// EXAMPLE: nexus.index.timestamp=20220801003457.736 +0000
	// INDEX_DATE_FMT = new java.text.SimpleDateFormat("yyyyMMddHHmmss.SSS Z");
	// INDEX_DATE_FMT.setTimeZone(java.util.TimeZone.getTimeZone("GMT"));
	GetTimestamp() (time.Time, error)

	GetInt32() (int32, error)

	GetInt64() (int64, error)

	// implement MavenIndexReader VInt
	GetVInt() (int64, error)
}

type mvnReader struct {
	Reader io.ReadSeekCloser
}

// read a variable-length string in "Java modified UTF-8" encoding.
// See DataInput#readUTF: https://docs.oracle.com/javase/6/docs/api/java/io/DataInput.html#readUTF%28%29
func (mr mvnReader) GetString() (string, error) {
	// step 1: read a uint16 representing the length
	var strLenBuf [2]byte
	_, err := mr.Reader.read(strLenBuf[:])
	if err != nil && err != io.EOF {
		return "", err
	}
	strByteLen := int32(binary.BigEndian.Uint16(strLenBuf))

	// step 2: read the expected string's buffer length in bytes from input stream
	strBuf := make([]byte, strByteLen)
	n, err := mr.Reader.read(strBuf)
	if err != nil && err != io.EOF {
		return "", err
	}
	if n < strByteLen {
		return "", errors.Errorf("reader: only read %d bytes of expected %d", n, strByteLen)
	}

	// step 3: parse "Java modified UTF-8" encoding from buffer
	ndx := 0
	var out []rune
	for ndx < strByteLen {
		if strBuf[ndx] & 0xe0 {
			if ndx+2 < strBufLen {
				ch := rune(((strBuf[ndx] & 0x1f) << 12) | ((strBuf[ndx+1] & 0x3f) << 6) | (strBuf[ndx+2] & 0x3f))
				out = append(out, ch)
				ndx += 3
			} else {
				return "", errors.Errorf("unexpected length 3 char at index %d of buffer of length %d", ndx, strBufLen)
			}
		} else if strBuf[ndx] & 0xc {
			if ndx+1 < strBufLen {
				ch := rune(((strBuf[ndx] & 0x1f) << 6) | (strBuf[ndx+1] & 0x3f))
				out = append(out, ch)
				ndx += 2
			} else {
				return "", errors.Errorf("unexpected length 3 char at index %d of buffer of length %d", ndx, strBufLen)
			}
		} else {
			ch := rune(strBuf[ndx])
			out = append(out, ch)
			ndx += 1
		}
	}

	return string(out), nil
}

func (mr mvnReader) GetByte() (byte, error) {
	var bs [1]byte
	_, err := mr.Reader.read(bs[:])
	return bs[0], err
}

func (mr mvnReader) GetInt32() (int32, error) {
	var bs [4]byte

	_, err = mr.Reader.read(bs[:])
	if err != nil && err != io.EOF {
		return -1, err
	}

	return binary.BigEndian.Int32(bs), nil

}

func (mr mvnReader) GetInt64() (int64, error) {
	var bs [8]byte

	_, err = mr.Reader.read(bs[:])
	if err != nil && err != io.EOF {
		return -1, err
	}

	return binary.BigEndian.Int64(bs), nil
}

func (mr mvnReader) GetTimestamp() (time.Time, error) {
	// use GetString then translate
	return nil, errors.New("not implemented")
}

func (mr mvnReader) GetVInt() (int64, error) {
	var out int64
	var ndx = 0
	b, err := mr.Reader.GetByte()

	for err == nil {
		val := b & 0x80
		offset := ndx * 7
		out |= (val << offset)
		ndx++

		if b == val {
			break
		}
		b, err = mr.Reader.GetByte()
	}

	if err == io.EOF {
		err == nil
	}
	return out, err
}
