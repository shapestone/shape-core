package hostiletest_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/shapestone/shape-core/pkg/data"
	"github.com/shapestone/shape-core/pkg/hostiletest"
)

// This file holds two test-only JSON implementations for the corpus:
//
//   - parseJSON, a small strict iterative parser built on data.Guard. It has
//     no recursion, so no input can overflow the Go stack. It shows that every
//     expectation in the corpus is reachable with pkg/data, and exercises the
//     harness end to end.
//   - stdlibJSON, the reference for the differential comparison, built on
//     encoding/json. The standard library is allowed here because this is
//     test code only.

type frame struct {
	isMap bool
	list  data.List
	m     *data.Map
	key   string
}

type jsonParser struct {
	in      []byte
	pos     int
	g       *data.Guard
	rejectD bool // reject duplicate keys
	stack   []frame
}

func parseJSON(in []byte, req hostiletest.Request) (data.Value, error) {
	v, _, err := parseJSONGuard(in, req)
	return v, err
}

// parseJSONGuard is parseJSON that also returns the guard, so a test can read
// its memory estimate.
func parseJSONGuard(in []byte, req hostiletest.Request) (data.Value, *data.Guard, error) {
	g := data.NewGuard(req.Limits)
	if err := g.Check(); err != nil {
		return data.Value{}, g, err
	}
	if err := g.Bytes(len(in)); err != nil {
		return data.Value{}, g, err
	}
	jp := &jsonParser{in: in, g: g, rejectD: req.Policy.RejectDuplicateKeys}
	v, err := jp.run()
	return v, g, err
}

func (p *jsonParser) fail(msg string) error {
	return &data.SyntaxError{Format: "json", Offset: p.pos, Msg: msg}
}

func (p *jsonParser) ws() {
	for p.pos < len(p.in) {
		switch p.in[p.pos] {
		case ' ', '\t', '\r', '\n':
			p.pos++
		default:
			return
		}
	}
}

func (p *jsonParser) peek() (byte, bool) {
	if p.pos < len(p.in) {
		return p.in[p.pos], true
	}
	return 0, false
}

func (p *jsonParser) run() (data.Value, error) {
	for {
		// Expect a value.
		p.ws()
		c, ok := p.peek()
		if !ok {
			return data.Value{}, p.fail("unexpected end of input")
		}
		var v data.Value
		switch c {
		case '[', '{':
			if err := p.g.Enter(); err != nil {
				return data.Value{}, err
			}
			if err := p.g.Node(); err != nil {
				return data.Value{}, err
			}
			p.pos++
			p.ws()
			if c == '[' {
				p.stack = append(p.stack, frame{})
				if d, _ := p.peek(); d == ']' {
					p.pos++
					v = p.pop()
					break
				}
				continue
			}
			p.stack = append(p.stack, frame{isMap: true, m: data.NewMap(0)})
			if d, _ := p.peek(); d == '}' {
				p.pos++
				v = p.pop()
				break
			}
			if err := p.readKey(); err != nil {
				return data.Value{}, err
			}
			continue
		default:
			var err error
			if v, err = p.scalar(c); err != nil {
				return data.Value{}, err
			}
		}

		// A value is complete: attach it, closing containers as they end.
		for {
			if len(p.stack) == 0 {
				p.ws()
				if p.pos != len(p.in) {
					return data.Value{}, p.fail("trailing data after value")
				}
				return v, nil
			}
			top := &p.stack[len(p.stack)-1]
			if top.isMap {
				if err := p.addEntry(top, v); err != nil {
					return data.Value{}, err
				}
			} else {
				top.list = append(top.list, v)
			}
			p.ws()
			d, ok := p.peek()
			closer := byte(']')
			if top.isMap {
				closer = '}'
			}
			switch {
			case ok && d == ',':
				p.pos++
				if top.isMap {
					if err := p.readKey(); err != nil {
						return data.Value{}, err
					}
				}
			case ok && d == closer:
				p.pos++
				v = p.pop()
				continue
			default:
				return data.Value{}, p.fail("expected ',' or closing bracket")
			}
			break
		}
	}
}

// pop closes the top container and returns its value.
func (p *jsonParser) pop() data.Value {
	top := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]
	p.g.Leave()
	if top.isMap {
		return data.MapValue(top.m)
	}
	return data.ListValue(top.list)
}

func (p *jsonParser) addEntry(f *frame, v data.Value) error {
	if p.rejectD && f.m.Has(f.key) {
		return p.fail("duplicate key " + clipKey(f.key))
	}
	f.m.Append(f.key, v)
	return p.g.MapSize(f.m.Len())
}

func clipKey(k string) string {
	if len(k) > 32 {
		return k[:32] + "..."
	}
	return k
}

// readKey reads `"key" :` and stores the key in the top frame.
func (p *jsonParser) readKey() error {
	p.ws()
	if c, ok := p.peek(); !ok || c != '"' {
		return p.fail("expected a string key")
	}
	k, err := p.str()
	if err != nil {
		return err
	}
	if err := p.g.Key(len(k)); err != nil {
		return err
	}
	p.ws()
	if c, ok := p.peek(); !ok || c != ':' {
		return p.fail("expected ':'")
	}
	p.pos++
	p.stack[len(p.stack)-1].key = k
	return nil
}

func (p *jsonParser) scalar(c byte) (data.Value, error) {
	switch {
	case c == '"':
		s, err := p.str()
		if err != nil {
			return data.Value{}, err
		}
		if err := p.g.Node(); err != nil {
			return data.Value{}, err
		}
		if err := p.g.String(len(s)); err != nil {
			return data.Value{}, err
		}
		return data.StringValue(s), nil
	case c == '-' || (c >= '0' && c <= '9'):
		start := p.pos
		for p.pos < len(p.in) && isNumberByte(p.in[p.pos]) {
			p.pos++
		}
		text := string(p.in[start:p.pos])
		v, err := data.NumberValue(text)
		if err != nil {
			p.pos = start
			return data.Value{}, p.fail("invalid number")
		}
		if err := p.g.Node(); err != nil {
			return data.Value{}, err
		}
		if err := p.g.Number(len(text)); err != nil {
			return data.Value{}, err
		}
		return v, nil
	}
	for lit, v := range map[string]data.Value{
		"true": data.BoolValue(true), "false": data.BoolValue(false), "null": data.NullValue(),
	} {
		if bytes.HasPrefix(p.in[p.pos:], []byte(lit)) {
			p.pos += len(lit)
			return v, p.g.Node()
		}
	}
	return data.Value{}, p.fail("unexpected character")
}

func isNumberByte(c byte) bool {
	return c >= '0' && c <= '9' || c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E'
}

// str reads a string starting at the opening quote.
func (p *jsonParser) str() (string, error) {
	p.pos++ // opening quote
	start := p.pos
	for p.pos < len(p.in) {
		switch c := p.in[p.pos]; {
		case c == '"':
			s := p.in[start:p.pos]
			p.pos++
			return string(s), p.g.CheckStringLen(len(s))
		case c == '\\':
			return p.escaped(start)
		case c < 0x20:
			return "", p.fail("control character in string")
		default:
			p.pos++
		}
	}
	return "", p.fail("unterminated string")
}

// escaped finishes a string that contains escapes; start is where the string
// body began.
func (p *jsonParser) escaped(start int) (string, error) {
	buf := append([]byte(nil), p.in[start:p.pos]...)
	for p.pos < len(p.in) {
		c := p.in[p.pos]
		switch {
		case c == '"':
			p.pos++
			return string(buf), p.g.CheckStringLen(len(buf))
		case c < 0x20:
			return "", p.fail("control character in string")
		case c != '\\':
			buf = append(buf, c)
			p.pos++
			continue
		}
		p.pos++
		if p.pos >= len(p.in) {
			break
		}
		e := p.in[p.pos]
		p.pos++
		switch e {
		case '"', '\\', '/':
			buf = append(buf, e)
		case 'b':
			buf = append(buf, '\b')
		case 'f':
			buf = append(buf, '\f')
		case 'n':
			buf = append(buf, '\n')
		case 'r':
			buf = append(buf, '\r')
		case 't':
			buf = append(buf, '\t')
		case 'u':
			r, ok := p.hex4()
			if !ok {
				return "", p.fail("invalid unicode escape")
			}
			if utf16.IsSurrogate(r) {
				save := p.pos
				if bytes.HasPrefix(p.in[p.pos:], []byte(`\u`)) {
					p.pos += 2
					if r2, ok := p.hex4(); ok {
						if d := utf16.DecodeRune(r, r2); d != utf8.RuneError {
							buf = utf8.AppendRune(buf, d)
							continue
						}
					}
				}
				p.pos = save
				r = utf8.RuneError
			}
			buf = utf8.AppendRune(buf, r)
		default:
			return "", p.fail("invalid escape")
		}
		if err := p.g.CheckStringLen(len(buf)); err != nil {
			return "", err
		}
	}
	return "", p.fail("unterminated string")
}

func (p *jsonParser) hex4() (rune, bool) {
	if p.pos+4 > len(p.in) {
		return 0, false
	}
	var r rune
	for _, c := range p.in[p.pos : p.pos+4] {
		switch {
		case c >= '0' && c <= '9':
			r = r<<4 | rune(c-'0')
		case c >= 'a' && c <= 'f':
			r = r<<4 | rune(c-'a'+10)
		case c >= 'A' && c <= 'F':
			r = r<<4 | rune(c-'A'+10)
		default:
			return 0, false
		}
	}
	p.pos += 4
	return r, true
}

// stdlibJSON is the differential reference: json.Valid decides the verdict
// and the token stream gives the value with key order and number text kept.
func stdlibJSON(in []byte) (data.Value, error) {
	if !json.Valid(in) {
		return data.Value{}, errors.New("encoding/json: invalid JSON")
	}
	dec := json.NewDecoder(bytes.NewReader(in))
	dec.UseNumber()
	return refValue(dec)
}

func refValue(dec *json.Decoder) (data.Value, error) {
	tok, err := dec.Token()
	if err != nil {
		return data.Value{}, err
	}
	switch t := tok.(type) {
	case nil:
		return data.NullValue(), nil
	case bool:
		return data.BoolValue(t), nil
	case string:
		return data.StringValue(t), nil
	case json.Number:
		return data.NumberValue(string(t))
	case json.Delim:
		if t == '[' {
			var l data.List
			for dec.More() {
				v, err := refValue(dec)
				if err != nil {
					return data.Value{}, err
				}
				l = append(l, v)
			}
			_, err := dec.Token() // ]
			return data.ListValue(l), err
		}
		m := data.NewMap(0)
		for dec.More() {
			k, err := dec.Token()
			if err != nil {
				return data.Value{}, err
			}
			v, err := refValue(dec)
			if err != nil {
				return data.Value{}, err
			}
			m.Append(k.(string), v)
		}
		_, err := dec.Token() // }
		return data.MapValue(m), err
	}
	return data.Value{}, io.ErrUnexpectedEOF
}
