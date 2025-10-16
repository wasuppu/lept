package lept

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	EOF = rune(-1)
)

var errExpectValue = errors.New("expect value")
var errInvaildValue = errors.New("invaild value")
var errPluralRoot = errors.New("plural root")
var errOutOfRange = errors.New("number out of range")
var errMissQuotation = errors.New("miss quotation mark")
var errMissComma = errors.New("miss comma")
var errMissSquareBracket = errors.New("miss square bracket")
var errMissCurlyBracket = errors.New("miss curly bracket")
var errMissKey = errors.New("miss object key")
var errMissColon = errors.New("miss colon")

type leptType string

const (
	LEPT_NULL   leptType = "NULL"
	LEPT_FALSE  leptType = "FALSE"
	LEPT_TRUE   leptType = "TRUE"
	LEPT_NUMBER leptType = "NUMBER"
	LEPT_STRING leptType = "STRING"
	LEPT_ARRAY  leptType = "ARRAY"
	LEPT_OBJECT leptType = "OBJECT"
)

type leptMember struct {
	K string
	V *leptValue
}

func (m leptMember) String() string {
	return fmt.Sprintf("%q: %v", m.K, m.V)
}

type leptObject []leptMember

func (obj leptObject) Index(i int) *leptMember {
	if i > len(obj) {
		return nil
	}
	return &obj[i]
}

func (obj leptObject) Get(k string) *leptValue {
	for i := len(obj) - 1; i >= 0; i-- {
		if obj[i].K == k {
			return obj[i].V
		}
	}

	return nil
}

func (obj leptObject) String() string {
	str := "{"
	i := 0
	for ; i < len(obj)-1; i++ {
		str += fmt.Sprint(obj[i]) + ", "
	}
	str += fmt.Sprint(obj[i]) + "}"
	return str
}

type leptArray []*leptValue

func (arr leptArray) Index(i int) *leptValue {
	if i > len(arr) {
		return nil
	}
	return arr[i]
}

func (arr leptArray) String() string {
	str := "["
	i := 0
	for ; i < len(arr)-1; i++ {
		str += fmt.Sprint(arr[i]) + ", "
	}
	str += fmt.Sprint(arr[i]) + "]"
	return str
}

type leptContext struct {
	json  string
	pos   int
	width int
}

func (c *leptContext) parseWhitespace() {
	r := c.next()
	for unicode.IsSpace(r) {
		r = c.next()
	}
	c.backup()
}

func (c *leptContext) next() (r rune) {
	if c.isAtEnd() {
		c.width = 0
		return EOF
	}
	r, c.width = utf8.DecodeRuneInString(c.json[c.pos:])
	c.pos += c.width
	return r
}

func (c *leptContext) backup() {
	c.pos -= c.width
}

func (c *leptContext) peek() rune {
	r := c.next()
	c.backup()
	return r
}

func (c *leptContext) isAtEnd() bool {
	return c.pos >= len(c.json)
}

func newContext(json string) *leptContext {
	return &leptContext{json, 0, 0}
}

type leptValue struct {
	u   any
	typ leptType
}

func (v leptValue) String() string {
	switch v.typ {
	case LEPT_STRING:
		return fmt.Sprintf("%q", v.u)
	case LEPT_OBJECT:
		return fmt.Sprint(v.u)
	default:
		return fmt.Sprint(v.u)
	}
}

func (v *leptValue) parse(json string) error {
	c := newContext(json)
	c.parseWhitespace()
	v.typ = LEPT_NULL
	err := v.parseValue(c)
	if err == nil {
		c.parseWhitespace()
		if !c.isAtEnd() {
			err = errPluralRoot
		}
	}
	return err
}

func (v *leptValue) parseValue(c *leptContext) error {
	if !c.isAtEnd() {
		switch c.peek() {
		case 'n':
			return v.parseLiteral(c, "null", LEPT_NULL)
		case 't':
			return v.parseLiteral(c, "true", LEPT_TRUE)
		case 'f':
			return v.parseLiteral(c, "false", LEPT_FALSE)
		case '"':
			return v.parseString(c)
		case '[':
			return v.parseArray(c)
		case '{':
			return v.parseObject(c)
		default:
			if unicode.IsDigit(c.peek()) || c.peek() == '-' {
				return v.parseNumber(c)
			} else {
				return fmt.Errorf("unexpected character %q", c.peek())
			}
		}
	} else {
		return errExpectValue
	}
}

func (v *leptValue) parseObject(c *leptContext) error {
	ms := leptObject{}
	c.next()
	c.parseWhitespace()
	if c.peek() == '}' {
		c.next()
		v.typ = LEPT_OBJECT
		v.u = ms
		return nil
	}

	for {
		m := leptMember{}

		if !c.isAtEnd() {
			if c.peek() != '"' {
				return errMissKey
			}
			s, err := v.parseStringRaw(c)
			if err != nil {
				return err
			}
			m.K = s

			c.parseWhitespace()
			if c.peek() != ':' {
				return errMissColon
			}
			c.next()

			c.parseWhitespace()
			e := &leptValue{}
			err = e.parseValue(c)
			if err != nil {
				return err
			}
			m.V = e
			ms = append(ms, m)
			c.parseWhitespace()
			if c.peek() == ',' {
				c.next()
				c.parseWhitespace()
			} else if c.peek() == '}' {
				c.next()
				v.typ = LEPT_OBJECT
				v.u = ms
				return nil
			} else {
				return errMissComma
			}
		} else {
			return errMissCurlyBracket
		}
	}
}

func (v *leptValue) parseArray(c *leptContext) error {
	arr := leptArray{}
	c.next()
	c.parseWhitespace()
	if c.peek() == ']' {
		c.next()
		v.typ = LEPT_ARRAY
		v.u = arr
		return nil
	}
	for {
		if !c.isAtEnd() {
			e := &leptValue{}
			err := e.parseValue(c)
			if err != nil {
				return err
			}
			arr = append(arr, e)
			c.parseWhitespace()
			if c.peek() == ',' {
				c.next()
				c.parseWhitespace()
			} else if c.peek() == ']' {
				c.next()
				v.typ = LEPT_ARRAY
				v.u = arr
				return nil
			} else {
				return errMissComma
			}
		} else {
			return errMissSquareBracket
		}
	}
}

func (v *leptValue) parseStringRaw(c *leptContext) (string, error) {
	c.next()
	start := c.pos
	for c.peek() != '"' {
		if !c.isAtEnd() {
			c.next()
		} else {
			return "", errMissQuotation
		}
	}
	s := c.json[start:c.pos]
	c.next()

	return s, nil
}

func (v *leptValue) parseString(c *leptContext) error {
	s, err := v.parseStringRaw(c)
	if err != nil {
		return err
	}
	v.u = s
	v.typ = LEPT_STRING
	return nil
}

func (v *leptValue) parseNumber(c *leptContext) error {
	start := c.pos
	if c.peek() == '-' {
		c.next()
	}

	if c.peek() == '0' {
		c.next()
	} else {
		if c.peek() == '0' {
			return errInvaildValue
		} else {
			c.next()
		}
		for unicode.IsDigit(c.peek()) {
			c.next()
		}
	}

	if c.peek() == '.' {
		c.next()
		if !unicode.IsDigit(c.peek()) {
			return errInvaildValue
		}
		for unicode.IsDigit(c.peek()) {
			c.next()
		}
	}

	if c.peek() == 'e' || c.peek() == 'E' {
		c.next()
		if c.peek() == '+' || c.peek() == '-' {
			c.next()
		}
		if !unicode.IsDigit(c.peek()) {
			return errInvaildValue
		}

		for unicode.IsDigit(c.peek()) {
			c.next()
		}
	}

	n, err := strconv.ParseFloat(c.json[start:c.pos], 64)
	if err != nil {
		return errOutOfRange
	}

	if !c.isAtEnd() {
		if !unicode.IsSpace(c.peek()) && c.peek() != ',' && c.peek() != ']' {
			return errInvaildValue
		}
	}

	v.u = n
	v.typ = LEPT_NUMBER
	return nil
}

func (v *leptValue) parseLiteral(c *leptContext, litetal string, typ leptType) error {
	if !strings.HasPrefix(c.json[c.pos:], litetal) {
		return errInvaildValue
	}

	c.pos += len(litetal)

	switch typ {
	case LEPT_TRUE, LEPT_FALSE:
		b, _ := strconv.ParseBool(litetal)
		v.u = b
	case LEPT_NULL:
		v.u = "null"
	}

	v.typ = typ

	return nil
}

func (v *leptValue) Get(k string) *leptValue {
	if v.typ == LEPT_OBJECT {
		ms, ok := v.u.(leptObject)
		if !ok {
			return nil
		}
		return ms.Get(k)
	}

	return nil
}

func (v *leptValue) Seek(ks ...string) *leptValue {
	if v.typ == LEPT_OBJECT {
		c := v
		for _, k := range ks {
			r := c.Get(k)
			if r != nil {
				c = r
			} else {
				return r
			}
		}
		return c
	}

	return nil
}

func (v *leptValue) Append(e ...*leptValue) *leptValue {
	if v.typ == LEPT_ARRAY {
		v.u = append(v.u.(leptArray), e...)
		return v
	}

	return nil
}

func (v *leptValue) Set(key string, val *leptValue) *leptValue {
	if v.typ == LEPT_OBJECT {
		m := leptMember{key, val}
		v.u = append(v.u.(leptObject), m)
		return v
	}

	return nil
}

func (v *leptValue) BOOL() bool {
	if v.typ == LEPT_TRUE || v.typ == LEPT_FALSE {
		return v.u.(bool)
	} else {
		return false
	}
}

func (v *leptValue) NULL() string {
	if v.typ == LEPT_NULL {
		return "null"
	} else {
		return ""
	}
}

func (v *leptValue) STRING() string {
	if v.typ == LEPT_STRING {
		return v.u.(string)
	} else {
		return ""
	}
}

func (v *leptValue) NUMBER() float64 {
	if v.typ == LEPT_NUMBER {
		return v.u.(float64)
	} else {
		return 0
	}
}

func (v *leptValue) ARRAY() leptArray {
	if v.typ == LEPT_ARRAY {
		return v.u.(leptArray)
	} else {
		return leptArray{}
	}
}

func (v *leptValue) OBJECT() leptObject {
	if v.typ == LEPT_OBJECT {
		return v.u.(leptObject)
	} else {
		return leptObject{}
	}
}

func NewBool(b bool) *leptValue {
	if b {
		return &leptValue{b, LEPT_TRUE}
	} else {
		return &leptValue{b, LEPT_FALSE}
	}
}

func NewString(s string) *leptValue {
	return &leptValue{s, LEPT_STRING}
}

func NewNumber(n float64) *leptValue {
	return &leptValue{n, LEPT_NUMBER}
}

func NewArray() *leptValue {
	return &leptValue{leptArray{}, LEPT_ARRAY}
}

func NewObject() *leptValue {
	return &leptValue{leptObject{}, LEPT_OBJECT}
}

func NewNull() *leptValue {
	return &leptValue{"null", LEPT_NULL}
}

func Parse(data string) (*leptValue, error) {
	v := &leptValue{}
	return v, v.parse(data)
}
