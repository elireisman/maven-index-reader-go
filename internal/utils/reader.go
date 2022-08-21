package utils

import (
	"encoding/binary"
	"io"
	"time"

	"github.com/pkg/errors"
)

// read a variable-length string in "Java modified UTF-8" encoding
func ReadString(r io.Reader) (string, error) {
	size, err := ReadUint16(r)
	if err != nil {
		return "", errors.Wrap(err, "ReadString: failed to read expected string length uint16 with cause")
	}

	return readUTF8String(r, int(size))
}

// read a variable-length string in "Java modified UTF-8" encoding.
func ReadLargeString(r io.Reader) (string, error) {
	size, err := ReadInt32(r)
	if err != nil {
		return "", errors.Wrap(err, "ReadStringLong: failed to read expected string length int32 with cause")
	}

	return readUTF8String(r, int(size))
}

// read a variable-length string in "Java modified UTF-8" encoding
func readUTF8String(r io.Reader, strByteLen int) (string, error) {
	// read the expected string's buffer length in bytes from input stream
	var bytesRead int
	var err error
	var strBuf = make([]byte, strByteLen)

	for bytesRead < strByteLen && err == nil {
		var n int
		n, err = r.Read(strBuf[bytesRead:])
		bytesRead += n
		if err != nil && errors.Cause(err) != io.EOF {
			return "", errors.Wrapf(err,
				"readUTF8String: failed to read expected buffer of size %d (got %d) with cause",
				strByteLen, bytesRead)
		}
	}

	// parse the buffer into std UTF-8. if no parse error,
	// conserve possible reader io.EOF for caller
	s, sErr := GetString(strBuf)
	if sErr == nil {
		sErr = err
	}
	return s, sErr
}

// ReadByte -
func ReadByte(r io.Reader) (byte, error) {
	var arr [1]byte
	_, err := r.Read(arr[:])
	return arr[0], err
}

// ReadUint16 -
func ReadUint16(r io.Reader) (uint16, error) {
	var bytesRead int
	var err error
	var arr [2]byte

	for bytesRead < 2 && err == nil {
		var n int
		n, err = r.Read(arr[bytesRead:])
		bytesRead += n
		if err != nil && errors.Cause(err) != io.EOF {
			return 0, err
		}
	}

	if bytesRead != 2 && errors.Cause(err) != io.EOF {
		return 0, errors.Errorf("ReadUint16: expected to read 2 bytes, got: %d", bytesRead)
	}

	// if no parse error, conserve possible reader io.EOF for caller
	return binary.BigEndian.Uint16(arr[:]), err

}

// ReadInt32 -
func ReadInt32(r io.Reader) (int32, error) {
	var bytesRead int
	var err error
	var arr [4]byte

	for bytesRead < 4 && err == nil {
		var n int
		n, err = r.Read(arr[bytesRead:])
		bytesRead += n
		if err != nil && errors.Cause(err) != io.EOF {
			return 0, err
		}
	}

	if bytesRead != 4 && errors.Cause(err) != io.EOF {
		return 0, errors.Errorf("ReadInt32: expected to read 4 bytes, got: %d", bytesRead)
	}

	// if no parse error, conserve possible reader io.EOF for caller
	return int32(binary.BigEndian.Uint32(arr[:])), err

}

// ReadInt64 -
func ReadInt64(r io.Reader) (int64, error) {
	var bytesRead int
	var err error
	var arr [8]byte

	for bytesRead < 8 && err == nil {
		var n int
		n, err = r.Read(arr[bytesRead:])
		bytesRead += n
		if err != nil && errors.Cause(err) != io.EOF {
			return 0, err
		}
	}

	if bytesRead != 8 && errors.Cause(err) != io.EOF {
		return 0, errors.Errorf("ReadInt64: expected to read 8 bytes, got: %d", bytesRead)
	}

	// if no parse error, conserve possible reader io.EOF for caller
	return int64(binary.BigEndian.Uint64(arr[:])), err
}

// ReadTimestamp -
func ReadTimestamp(r io.Reader) (time.Time, error) {
	timeStr, err := ReadString(r)
	if err != nil && errors.Cause(err) != io.EOF {
		return time.Now().UTC(), errors.Wrap(err, "ReadTimestamp: failed to obtain string from decoder with cause")
	}

	// if no parse error, conserve possible reader io.EOF for caller
	t, tErr := GetTimestamp(timeStr)
	if tErr == nil {
		tErr = err
	}
	return t, tErr
}

// ReadVInt -
func ReadVInt(r io.Reader) (int64, error) {
	var out int64
	var ndx = 0
	var buf [1]byte
	slc := buf[:]
	_, err := r.Read(slc)
	b := buf[0]

	for err == nil {
		val := b & 0x7F
		offset := ndx * 7
		out |= (int64(val) << offset)
		ndx++

		if b == val {
			break
		}

		_, err = r.Read(slc)
		b = buf[0]
	}

	// if no parse error, conserve possible reader io.EOF for caller
	return out, err
}
