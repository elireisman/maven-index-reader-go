package utils

import (
	"bytes"
	"io"
	"testing"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestReadString(t *testing.T) {
	code, _ := utf8.DecodeRune([]byte{0xc3, 0xbd})
	payload := "Hello " + string(code) + " World"
	content := []byte(payload)
	size := uint16(len(content))
	sizeBytes := []byte{
		byte((size >> 8)),
		byte(size & 0xff),
	}
	buffer := bytes.NewBuffer(append(sizeBytes, content...))

	got, err := ReadString(buffer)
	require.NoError(t, err, payload)
	require.Equal(t, payload, got)
}

func TestReadLargeString(t *testing.T) {
	code, _ := utf8.DecodeRune([]byte{0xc3, 0xbd})
	payload := "Hello " + string(code) + " World"
	content := []byte(payload)
	size := int32(len(content))
	sizeBytes := []byte{
		byte((size & 0xff) >> 24),
		byte((size & 0xff) >> 16),
		byte((size & 0xff) >> 8),
		byte(size & 0xff),
	}
	buffer := bytes.NewBuffer(append(sizeBytes, content...))

	got, err := ReadLargeString(buffer)
	require.NoError(t, err, payload)
	require.Equal(t, payload, got)
}

func TestReadUint16(t *testing.T) {
	buffer := bytes.NewBuffer(
		[]byte{
			0b11011011,
			0b00100100,
		})
	got, err := ReadUint16(buffer)
	require.NoError(t, err)
	require.Equal(t, uint16(56100), got)

	_, err = ReadUint16(buffer)
	require.Error(t, err)
	require.Equal(t, io.EOF, errors.Cause(err))

	bigBuffer := bytes.NewBuffer(
		[]byte{
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
		})
	got, err = ReadUint16(bigBuffer)
	require.NoError(t, err)
	require.Equal(t, uint16(56100), got)
	got, err = ReadUint16(bigBuffer)
	require.NoError(t, err)
	require.Equal(t, uint16(56100), got)
}

func TestReadInt32(t *testing.T) {
	buffer := bytes.NewBuffer(
		[]byte{
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
		})
	got, err := ReadInt32(buffer)
	require.NoError(t, err)
	require.Equal(t, int32(-618341596), got)

	_, err = ReadInt32(buffer)
	require.Error(t, err)
	require.Equal(t, io.EOF, errors.Cause(err))

	bigBuffer := bytes.NewBuffer(
		[]byte{
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
		})
	got, err = ReadInt32(bigBuffer)
	require.NoError(t, err)
	require.Equal(t, int32(-618341596), got)
	got, err = ReadInt32(bigBuffer)
	require.NoError(t, err)
	require.Equal(t, int32(-618341596), got)
}

func TestReadInt64(t *testing.T) {
	buffer := bytes.NewBuffer(
		[]byte{
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
		})
	got, err := ReadInt64(buffer)
	require.NoError(t, err)
	require.Equal(t, int64(-2655756928899818716), got)

	_, err = ReadInt64(buffer)
	require.Error(t, err)
	require.Equal(t, io.EOF, errors.Cause(err))

	bigBuffer := bytes.NewBuffer(
		[]byte{
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
			0b11011011,
			0b00100100,
		})
	got, err = ReadInt64(bigBuffer)
	require.NoError(t, err)
	require.Equal(t, int64(-2655756928899818716), got)
	got, err = ReadInt64(bigBuffer)
	require.NoError(t, err)
	require.Equal(t, int64(-2655756928899818716), got)
}

func TestReadVInt(t *testing.T) {
	got, err := ReadVInt(bytes.NewBuffer(
		[]byte{
			0b00001000,
		}))
	require.NoError(t, err)
	require.Equal(t, int64(0b00000000000000000000000000001000), got)

	buffer := bytes.NewBuffer(
		[]byte{
			0b10001000,
			0b10001000,
			0b00001000,
		})
	got, err = ReadVInt(buffer)
	require.NoError(t, err)
	require.Equal(t, int64(0b00000000000000100000010000001000), got)

	_, err = ReadVInt(buffer)
	require.Error(t, err)
	require.Equal(t, io.EOF, errors.Cause(err))
}
