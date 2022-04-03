package server

import (
	"errors"
	"fmt"
)

func parseHttpMessage(chars *charStream) (HttpRequest, error) {
	requestLine, err := parseRequestLine(chars)
	if err != nil {
		return HttpRequest{}, err
	}

	headers := make(map[string][]string, 0)
	for {
		peek, err := chars.Peek()
		if err != nil {
			return HttpRequest{}, err
		}

		if peek == '\r' {
			break
		}

		headerField, err := parseHeaderField(chars)
		if err != nil {
			return HttpRequest{}, err
		}

		headers[headerField.Name] = append(headers[headerField.Name], headerField.Value)

		if err := chars.Expect("\r\n"); err != nil {
			return HttpRequest{}, err
		}
	}

	if err := chars.Expect("\r\n"); err != nil {
		return HttpRequest{}, err
	}

	return HttpRequest{RequestLine: requestLine, Headers: headers}, nil
}

func parseHeaderField(chars *charStream) (HeaderField, error) {
	name, err := parseFieldName(chars)
	if err != nil {
		return HeaderField{}, err
	}

	if err := chars.Expect(":"); err != nil {
		return HeaderField{}, err
	}

	skipOptionalWhitespace(chars)

	value, err := parseFieldValue(chars)
	if err != nil {
		return HeaderField{}, err
	}

	skipOptionalWhitespace(chars)

	return HeaderField{Name: name, Value: value}, nil
}

func parseFieldValue(chars *charStream) (string, error) {
	result := ""

	whitespace := -1
	for {
		char, err := chars.Peek()
		if err != nil {
			return "", err
		}

		if char >= 0x21 && char <= 0x7E || char >= 0x80 && char <= 0xFF {
			whitespace = -1
		} else if char == ' ' || char == '\t' {
			whitespace = len(result)
		} else if whitespace != -1 {
			return result[:whitespace], nil
		} else {
			return result, nil
		}

		_, err = chars.Next()
		if err != nil {
			return "", err
		}

		result += string(char)
	}
}

func parseFieldName(chars *charStream) (string, error) {
	return parseToken(chars)
}

func skipOptionalWhitespace(chars *charStream) error {
	for {
		char, err := chars.Peek()
		if err != nil {
			return err
		}

		if char == ' ' || char == '\t' {
			_, err = chars.Next()
			if err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func parseRequestLine(chars *charStream) (RequestLine, error) {
	method, err := parseMethod(chars)
	if err != nil {
		return RequestLine{}, err
	}

	if err := chars.Expect(" "); err != nil {
		return RequestLine{}, err
	}

	target, err := parseRequestTarget(chars)
	if err != nil {
		return RequestLine{}, err
	}

	if err := chars.Expect(" "); err != nil {
		return RequestLine{}, err
	}

	version, err := parseHttpVersion(chars)

	if err := chars.Expect("\r\n"); err != nil {
		return RequestLine{}, err
	}

	return RequestLine{Method: method, Target: target, Version: version}, nil
}

func parseRequestTarget(chars *charStream) ([]string, error) {
	peek, err := chars.Peek()
	if err != nil {
		return nil, err
	}

	if peek == '/' {
		// origin-form
		return parseOriginForm(chars)
	} else if peek >= 'a' && peek <= 'z' || peek >= 'A' && peek <= 'Z' {
		// absolute-form / authority-form
		panic("Absolute and authority form are unsupported.")
	} else if peek == '*' {
		// asterisk-form
		panic("Asterisk form is unsupported.")
	} else {
		return nil, fmt.Errorf("Unexpected character %q; expected a request target", peek)
	}
}

func parseOriginForm(chars *charStream) ([]string, error) {
	// We assume the request target is in origin form

	segments := make([]string, 0, 1)
	for {
		peek, err := chars.Peek()
		if err != nil {
			return nil, err
		}

		if peek != '/' {
			break
		}

		segment, err := parseSegment(chars)
		if err != nil {
			return nil, err
		}

		segments = append(segments, segment)
	}

	return segments, nil
}

func parseSegment(chars *charStream) (string, error) {
	if err := chars.Expect("/"); err != nil {
		return "", err
	}

	segment := ""
	for {
		char, err := chars.Peek()
		if err != nil {
			return "", err
		}

		switch char {
		case '-':
		case '.':
		case '_':
		case '~':
		case '!':
		case '$':
		case '&':
		case '\'':
		case '(':
		case ')':
		case '*':
		case '+':
		case ',':
		case ';':
		case '=':
		case ':':
		case '@':
		case '%':
			upper, err := parseHexDigit(chars)
			if err != nil {
				return "", err
			}

			lower, err := parseHexDigit(chars)
			if err != nil {
				return "", err
			}

			segment += string(upper<<4 | lower)
			continue
		default:
			if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') {
				return segment, nil
			}
		}

		_, err = chars.Next()
		if err != nil {
			return "", err
		}

		segment += string(char)
	}
}

func parseHexDigit(chars *charStream) (uint8, error) {
	char, err := chars.Next()
	if err != nil {
		return 0, err
	}

	if char == 0 {
		return 0, errors.New("Unexpected EOF, expected digit")
	} else if char >= '0' && char <= '9' {
		return char - '0', nil
	} else if char >= 'a' && char <= 'f' {
		return char - 'a' + 10, nil
	} else if char >= 'A' && char <= 'F' {
		return char - 'A' + 10, nil
	} else {
		return 0, fmt.Errorf("Invalid digit character: %q", char)
	}
}

func parseMethod(chars *charStream) (string, error) {
	return parseToken(chars)
}

func parseToken(chars *charStream) (string, error) {
	result := ""

	for {
		char, err := chars.Peek()
		if err != nil {
			return "", err
		}

		if char == '!' ||
			char == '#' ||
			char == '$' ||
			char == '%' ||
			char == '&' ||
			char == '\'' ||
			char == '*' ||
			char == '+' ||
			char == '-' ||
			char == ',' ||
			char == '^' ||
			char == '_' ||
			char == '`' ||
			char == '|' ||
			char == '~' ||
			char >= 'a' && char <= 'z' ||
			char >= 'A' && char <= 'Z' ||
			char >= '0' && char <= '9' {
			_, err = chars.Next()
			if err != nil {
				return "", err
			}

			result += string(char)
		} else {
			return result, nil
		}
	}
}

func parseStatusLine(chars *charStream) (uint8, uint16, error) {
	versionNumber, err := parseHttpVersion(chars)
	if err != nil {
		return 0, 0, err
	}

	if err := chars.Expect(" "); err != nil {
		return 0, 0, err
	}

	statusCode, err := parseStatusCode(chars)
	if err != nil {
		return 0, 0, err
	}

	if err := chars.Expect(" "); err != nil {
		return 0, 0, err
	}

	// skip over reason-phrase and break on CRLF
	for {
		char, err := chars.Next()
		if err != nil {
			return 0, 0, err
		}

		for char == '\r' {
			char, err = chars.Next()
			if err != nil {
				return 0, 0, err
			}

			if char == '\n' || char != '\r' {
				return versionNumber, statusCode, err
			}
		}
	}
}

func parseStatusCode(chars *charStream) (uint16, error) {
	top, err := parseDigit(chars)
	if err != nil {
		return 0, err
	}

	mid, err := parseDigit(chars)
	if err != nil {
		return 0, err
	}

	bot, err := parseDigit(chars)
	if err != nil {
		return 0, err
	}

	return 100*uint16(top) + 10*uint16(mid) + uint16(bot), nil
}

// parseVersion parses the HTTP-version grammar rule, returning the major/minor
// version with the major version as the top 4 bits, and the minor as the bottom 4.
func parseHttpVersion(chars *charStream) (uint8, error) {
	if err := chars.Expect("HTTP/"); err != nil {
		return 0, err
	}

	major, err := parseDigit(chars)
	if err != nil {
		return 0, err
	}

	if err := chars.Expect("."); err != nil {
		return 0, err
	}

	minor, err := parseDigit(chars)
	if err != nil {
		return 0, err
	}

	return major<<4 | minor, nil
}

// parseDigit parses a single decimal digit, returning the parsed value.
func parseDigit(chars *charStream) (uint8, error) {
	char, err := chars.Next()
	if err != nil {
		return 0, err
	}

	if char == 0 {
		return 0, errors.New("Unexpected EOF, expected digit")
	} else if char < '0' || char > '9' {
		return 0, fmt.Errorf("Invalid digit character: %q", char)
	}

	return char - '0', nil
}
