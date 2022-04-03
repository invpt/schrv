package server

import (
	"errors"
	"fmt"
	"io"
)

type charStream struct {
	reader  io.Reader
	hasPeek bool
	peek    byte
}

func newCharStream(rd io.Reader) charStream {
	return charStream{reader: rd}
}

func (chars *charStream) Next() (byte, error) {
	if chars.hasPeek {
		chars.hasPeek = false
		return chars.peek, nil
	}

	result := make([]byte, 1)
	n, err := chars.reader.Read(result)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, errors.New("Unexpected end of stream")
	}

	return result[0], nil
}

func (chars *charStream) Peek() (byte, error) {
	if !chars.hasPeek {
		peek, err := chars.Next()
		if err != nil {
			return 0, err
		}

		chars.peek = peek
		chars.hasPeek = true
	}

	return chars.peek, nil
}

func (chars *charStream) Expect(expected string) error {
	for i := 0; i < len(expected); i++ {
		found, err := chars.Next()
		if err != nil {
			return err
		}

		if found != expected[i] {
			return fmt.Errorf("Unexpected character %q; expected %q", found, expected[i])
		}
	}

	return nil
}

func (chars *charStream) Rest() (string, error) {
	const ChunkSize = 1024

	pos := 0
	buffer := make([]byte, ChunkSize)
	if chars.hasPeek {
		buffer[0] = chars.peek
		pos += 1
	}

	for {
		n, err := chars.reader.Read(buffer[pos:])

		pos += n

		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		if pos == len(buffer) {
			newBuffer := make([]byte, len(buffer)+ChunkSize)
			buffer = newBuffer
		}
	}

	return string(buffer[:pos]), nil
}

func (chars *charStream) Read(amount uint) (string, error) {
	buf := make([]byte, amount)
	pos := 0
	if chars.hasPeek && amount != 0 {
		next, err := chars.Next()
		if err != nil {
			return "", err
		}

		buf[0] = next
		pos += 1
	}

	for {
		n, err := chars.reader.Read(buf[pos:])

		pos += n

		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
	}

	return string(buf), nil
}
