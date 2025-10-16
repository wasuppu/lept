package lept

import (
	"errors"
	"fmt"
	"reflect"
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

var errMismatchType = errors.New("mismatch type")
var errUnsupportedType = func(v any) error { return errorf("unsupported bencode type %T", v) }

func errorf(msg string, args ...any) error {
	return fmt.Errorf(msg, args...)
}

type Type string

const (
	TypeNull   Type = "NULL"
	TypeFalse  Type = "FALSE"
	TypeTrue   Type = "TRUE"
	TypeNumber Type = "NUMBER"
	TypeString Type = "STRING"
	TypeArray  Type = "ARRAY"
	TypeObject Type = "OBJECT"
)

type Member struct {
	K string
	V *Value
}

func (m Member) String() string {
	return fmt.Sprintf("%q: %v", m.K, m.V)
}

type Object []Member

func (obj Object) Index(i int) *Member {
	if i > len(obj) {
		return nil
	}
	return &obj[i]
}

func (obj Object) Get(k string) *Value {
	for i := len(obj) - 1; i >= 0; i-- {
		if obj[i].K == k {
			return obj[i].V
		}
	}

	return nil
}

func (obj Object) String() string {
	str := "{"
	i := 0
	for ; i < len(obj)-1; i++ {
		str += fmt.Sprint(obj[i]) + ", "
	}
	str += fmt.Sprint(obj[i]) + "}"
	return str
}

type Array []*Value

func (arr Array) Index(i int) *Value {
	if i > len(arr) {
		return nil
	}
	return arr[i]
}

func (arr Array) String() string {
	str := "["
	i := 0
	for ; i < len(arr)-1; i++ {
		str += fmt.Sprint(arr[i]) + ", "
	}
	str += fmt.Sprint(arr[i]) + "]"
	return str
}

type Context struct {
	json  string
	pos   int
	width int
}

func (c *Context) parseWhitespace() {
	r := c.next()
	for unicode.IsSpace(r) {
		r = c.next()
	}
	c.backup()
}

func (c *Context) next() (r rune) {
	if c.isAtEnd() {
		c.width = 0
		return EOF
	}
	r, c.width = utf8.DecodeRuneInString(c.json[c.pos:])
	c.pos += c.width
	return r
}

func (c *Context) backup() {
	c.pos -= c.width
}

func (c *Context) peek() rune {
	r := c.next()
	c.backup()
	return r
}

func (c *Context) isAtEnd() bool {
	return c.pos >= len(c.json)
}

func newContext(json string) *Context {
	return &Context{json, 0, 0}
}

type Value struct {
	U    any
	Type Type
}

func (v Value) String() string {
	switch v.Type {
	case TypeString:
		return fmt.Sprintf("%q", v.U)
	case TypeObject:
		return fmt.Sprint(v.U)
	default:
		return fmt.Sprint(v.U)
	}
}

func (v *Value) parse(json string) error {
	c := newContext(json)
	c.parseWhitespace()
	v.Type = TypeNull
	err := v.parseValue(c)
	if err == nil {
		c.parseWhitespace()
		if !c.isAtEnd() {
			err = errPluralRoot
		}
	}
	return err
}

func (v *Value) parseValue(c *Context) error {
	if !c.isAtEnd() {
		switch c.peek() {
		case 'n':
			return v.parseLiteral(c, "null", TypeNull)
		case 't':
			return v.parseLiteral(c, "true", TypeTrue)
		case 'f':
			return v.parseLiteral(c, "false", TypeFalse)
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

func (v *Value) parseObject(c *Context) error {
	ms := Object{}
	c.next()
	c.parseWhitespace()
	if c.peek() == '}' {
		c.next()
		v.Type = TypeObject
		v.U = ms
		return nil
	}

	for {
		m := Member{}

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
			e := &Value{}
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
				v.Type = TypeObject
				v.U = ms
				return nil
			} else {
				return errMissComma
			}
		} else {
			return errMissCurlyBracket
		}
	}
}

func (v *Value) parseArray(c *Context) error {
	arr := Array{}
	c.next()
	c.parseWhitespace()
	if c.peek() == ']' {
		c.next()
		v.Type = TypeArray
		v.U = arr
		return nil
	}
	for {
		if !c.isAtEnd() {
			e := &Value{}
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
				v.Type = TypeArray
				v.U = arr
				return nil
			} else {
				return errMissComma
			}
		} else {
			return errMissSquareBracket
		}
	}
}

func (v *Value) parseStringRaw(c *Context) (string, error) {
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

func (v *Value) parseString(c *Context) error {
	s, err := v.parseStringRaw(c)
	if err != nil {
		return err
	}
	v.U = s
	v.Type = TypeString
	return nil
}

func (v *Value) parseNumber(c *Context) error {
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

	v.U = n
	v.Type = TypeNumber
	return nil
}

func (v *Value) parseLiteral(c *Context, litetal string, typ Type) error {
	if !strings.HasPrefix(c.json[c.pos:], litetal) {
		return errInvaildValue
	}

	c.pos += len(litetal)

	switch typ {
	case TypeTrue, TypeFalse:
		b, _ := strconv.ParseBool(litetal)
		v.U = b
	case TypeNull:
		v.U = "null"
	}

	v.Type = typ

	return nil
}

func (v *Value) Get(k string) *Value {
	if v.Type == TypeObject {
		ms, ok := v.U.(Object)
		if !ok {
			return nil
		}
		return ms.Get(k)
	}

	return nil
}

func (v *Value) Seek(ks ...string) *Value {
	if v.Type == TypeObject {
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

func (v *Value) Append(e ...*Value) *Value {
	if v.Type == TypeArray {
		v.U = append(v.U.(Array), e...)
		return v
	}

	return nil
}

func (v *Value) Set(key string, val *Value) *Value {
	if v.Type == TypeObject {
		m := Member{key, val}
		v.U = append(v.U.(Object), m)
		return v
	}

	return nil
}

func (v *Value) BOOL() bool {
	if v.Type == TypeTrue || v.Type == TypeFalse {
		return v.U.(bool)
	} else {
		return false
	}
}

func (v *Value) NULL() string {
	if v.Type == TypeNull {
		return "null"
	} else {
		return ""
	}
}

func (v *Value) STRING() string {
	if v.Type == TypeString {
		return v.U.(string)
	} else {
		return ""
	}
}

func (v *Value) NUMBER() float64 {
	if v.Type == TypeNumber {
		return v.U.(float64)
	} else {
		return 0
	}
}

func (v *Value) ARRAY() Array {
	if v.Type == TypeArray {
		return v.U.(Array)
	} else {
		return Array{}
	}
}

func (v *Value) OBJECT() Object {
	if v.Type == TypeObject {
		return v.U.(Object)
	} else {
		return Object{}
	}
}

func NewBool(b bool) *Value {
	if b {
		return &Value{b, TypeTrue}
	} else {
		return &Value{b, TypeFalse}
	}
}

func NewString(s string) *Value {
	return &Value{s, TypeString}
}

func NewNumber(n float64) *Value {
	return &Value{n, TypeNumber}
}

func NewArray() *Value {
	return &Value{Array{}, TypeArray}
}

func NewObject() *Value {
	return &Value{Object{}, TypeObject}
}

func NewNull() *Value {
	return &Value{"null", TypeNull}
}

func Parse(data string) (*Value, error) {
	v := &Value{}
	return v, v.parse(data)
}

func Unmarshal(parsed *Value, v any) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errorf("Attempt to unmarshal into a non-pointer")
	}

	return unmarshalValue(parsed, reflect.Indirect(reflect.ValueOf(v)))
}

func unmarshalValue(parsed *Value, v reflect.Value) (err error) {
	switch parsed.Type {
	case TypeTrue, TypeFalse:
		switch v.Kind() {
		case reflect.Bool:
			v.SetBool(parsed.BOOL())
		case reflect.Interface:
			v.Set(reflect.ValueOf(parsed.BOOL()))
		default:
			err = errMismatchType
		}
	case TypeNumber:
		switch v.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			v.SetInt(int64(parsed.NUMBER()))
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			v.SetUint(uint64(parsed.NUMBER()))
		case reflect.Float64, reflect.Float32:
			v.SetFloat(parsed.NUMBER())
		case reflect.Interface:
			v.Set(reflect.ValueOf(parsed.NUMBER()))
		default:
			err = errMismatchType
		}
	case TypeString:
		switch v.Kind() {
		case reflect.String:
			if !v.CanSet() {
				x := ""
				v = reflect.ValueOf(&x).Elem()
			}
			v.SetString(parsed.STRING())
		case reflect.Interface:
			v.Set(reflect.ValueOf(parsed.STRING()))
		default:
			err = errMismatchType
		}
	case TypeArray:
		switch v.Kind() {
		case reflect.Slice:
			l := reflect.MakeSlice(v.Type(), len(parsed.ARRAY()), len(parsed.ARRAY()))
			for i, e := range parsed.ARRAY() {
				if err = unmarshalValue(e, reflect.Indirect(l.Index(i))); err != nil {
					return
				}
			}
			v.Set(l)
		case reflect.Array:
			if v.Len() != len(parsed.ARRAY()) {
				err = fmt.Errorf("array length mismatch: %d vs %d", v.Len(), len(parsed.ARRAY()))
				return
			}
			for i, e := range parsed.ARRAY() {
				elem := v.Index(i)
				if err = unmarshalValue(e, elem); err != nil {
					return
				}
			}
		default:
			err = errMismatchType
		}
	case TypeObject:
		t := v.Type()
		for i := range v.NumField() {
			f := v.Field(i)
			if !f.CanSet() {
				continue
			}
			key := t.Field(i).Tag.Get("json")
			if key == "" {
				continue
			}

			v := parsed.OBJECT().Get(key)
			if v == nil {
				continue
			}
			if err = unmarshalValue(v, reflect.Indirect(f)); err != nil {
				return
			}
		}
	case TypeNull:
	default:
		err = errUnsupportedType(parsed)
	}
	return
}
