package f

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	// https://github.com/andrewstuart/goq
	// eg. type example struct { Title string `html:"h1"` }
	// All important settings
	html2TagName = "html"
	html2Prefix  = '!'
	html2Ignore  = "!ignore"

	// All "Reason" fields within HtmlCannotUnmarshalErr will be constants and part of this list
	html2NonPointer         = "non-pointer value"
	html2NilValue           = "destination argument is nil"
	html2ArrayLenMismatch   = "array length does not match document elements found"
	html2CustomUnmarshalErr = "a custom unmarshal implementation threw an error"
	html2TypeConversionErr  = "a type conversion error occurred"
	html2MapKeyUnmarshalErr = "error unmarshal a map key"
	html2MissingValSelector = "at least one value selector must be passed to use as map index"
)

// NewHtmlDecoder returns a new decoder given an io.Reader
func NewHtmlDecoder(r io.Reader) *HTMLDecoder {
	d := &HTMLDecoder{}
	d.doc, d.err = goquery.NewDocumentFromReader(r)
	return d
}

// NewHtmlSelection is a quick utility function to get a goquery.Selection from a
// slice of *html.Node. Useful for performing unmarshal, since the decision
// was made to use []*html.Node for maximum flexibility.
func NewHtmlSelection(nodes []*html.Node) *goquery.Selection {
	sel := &goquery.Selection{}
	return sel.AddNodes(nodes...)
}

// HtmlUnmarshal takes a byte slice and a destination pointer to any
// interface{}, and unmarshal the document into the destination based on the
// rules above. Any error returned here will likely be of type
// HtmlCannotUnmarshalErr, though an initial goquery error will pass through directly.
func HtmlUnmarshal(bs []byte, v interface{}) error {
	d, err := goquery.NewDocumentFromReader(bytes.NewReader(bs))

	if err != nil {
		return err
	}

	return HtmlUnmarshalSelection(d.Selection, v)
}

// HTMLDecoder implements the same API you will see in encoding/xml and
// encoding/json except that we do not currently support proper streaming
// decoding as it is not supported by goquery upstream.
type HTMLDecoder struct {
	err error
	doc *goquery.Document
}

// Decode will unmarshal the contents of the decoder when given an instance of
// an annotated type as its argument. It will return any errors encountered
// during either parsing the document or unmarshal into the given object.
func (d *HTMLDecoder) Decode(dest interface{}) error {
	if d.err != nil {
		return d.err
	}
	if d.doc == nil {
		return &HtmlCannotUnmarshalErr{
			Reason: "resulting document was nil",
		}
	}

	return HtmlUnmarshalSelection(d.doc.Selection, dest)
}

// UnmarshalHTMLer allows for custom implementations of unmarshal logic
type UnmarshalHTMLer interface {
	UnmarshalHTML([]*html.Node) error
}

// reflectUnmarshalHTMLer is stolen mostly from pkg/encoding/json/decode.go and removed some
// cases (handling `null`) that go doesn't need to handle.
func reflectUnmarshalHTMLer(v reflect.Value) (UnmarshalHTMLer, reflect.Value) {
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (e.Elem().Kind() == reflect.Ptr) {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.IsNil() {
			v.Set(reflect.New(TypeElem(v.Type())))
		}
		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(UnmarshalHTMLer); ok {
				return u, reflect.Value{}
			}
		}
		v = v.Elem()
	}
	return nil, v
}

// HtmlCannotUnmarshalErr represents an error returned by the goquery HtmlUnmarshal
// and helps consumers in programmatically diagnosing the cause of their error.
type HtmlCannotUnmarshalErr struct {
	Err      error
	Val      string
	FldOrIdx interface{}

	V      reflect.Value
	Reason string
}

// This type is a mid-level abstraction to help understand the error printing logic
type html2ErrorChain struct {
	chain []*HtmlCannotUnmarshalErr
	val   string
	tail  error
}

// tPath returns the type path in the same string format one might use to access
// the nested value in go code. This should hopefully help make debugging easier.
func (e html2ErrorChain) tPath() string {
	nest := ""

	for _, err := range e.chain {
		if err.FldOrIdx != nil {
			switch nesting := err.FldOrIdx.(type) {
			case string:
				switch err.V.Type().Kind() {
				case reflect.Map:
					nest += fmt.Sprintf("[%q]", nesting)
				case reflect.Struct:
					nest += fmt.Sprintf(".%s", nesting)
				}
			case int:
				nest += fmt.Sprintf("[%d]", nesting)
			case *int:
				nest += fmt.Sprintf("[%d]", *nesting)
			default:
				fmt.Printf("err.FldOrIdx = %#v\n", err.FldOrIdx)
				nest += fmt.Sprintf("[%v]", nesting)
			}
		}
	}

	return nest
}

func (e html2ErrorChain) last() *HtmlCannotUnmarshalErr {
	return e.chain[len(e.chain)-1]
}

// Error gives a human-readable error message for debugging purposes.
func (e html2ErrorChain) Error() string {
	last := e.last()

	// Avoid panic if we cannot get a type name for the Value
	t := "unknown: invalid value"
	if last.V.IsValid() {
		t = last.V.Type().String()
	}

	msg := "could not unmarshal "

	if e.val != "" {
		msg += fmt.Sprintf("value %q ", e.val)
	}

	msg += fmt.Sprintf(
		"into '%s%s' (type %s): %s",
		e.chain[0].V.Type(),
		e.tPath(),
		t,
		last.Reason,
	)

	// If a generic error was reported elsewhere, report its message last
	if e.tail != nil {
		msg = msg + ": " + e.tail.Error()
	}

	return msg
}

// Traverse e.Err, printing hopefully helpful type info until there are no more
// chained errors.
func (e *HtmlCannotUnmarshalErr) unwind() *html2ErrorChain {
	str := &html2ErrorChain{chain: []*HtmlCannotUnmarshalErr{}}
	for {
		str.chain = append(str.chain, e)

		if e.Val != "" {
			str.val = e.Val
		}

		// Terminal error was of type *HtmlCannotUnmarshalErr and had no children
		if e.Err == nil {
			return str
		}

		if e2, ok := e.Err.(*HtmlCannotUnmarshalErr); ok {
			e = e2
			continue
		}

		// Child error was not a *HtmlCannotUnmarshalErr; print its message
		str.tail = e.Err
		return str
	}
}

func (e *HtmlCannotUnmarshalErr) Error() string {
	return e.unwind().Error()
}

type html2ValFunc func(*goquery.Selection) string

type html2QueryTag string

func (tag html2QueryTag) preprocess(s *goquery.Selection) *goquery.Selection {
	arr := strings.Split(string(tag), ",")
	var offset int
	for len(arr)-1 > offset && arr[offset][0] == html2Prefix {
		m := arr[offset][1:]
		v := reflect.ValueOf(s).MethodByName(m)
		if !v.IsValid() {
			return s
		}

		result := v.Call(nil)

		if sel, ok := result[0].Interface().(*goquery.Selection); ok {
			s = sel
		}
		offset++
	}
	return s
}

func (tag html2QueryTag) selector(which int) string {
	arr := strings.Split(string(tag), ",")
	if which > len(arr)-1 {
		return ""
	}
	var offset int
	for len(arr) > offset && arr[offset][0] == html2Prefix {
		offset++
	}
	return arr[which+offset]
}

var (
	html2TextVal html2ValFunc = func(s *goquery.Selection) string {
		return strings.TrimSpace(s.Text())
	}
	html2Val = func(s *goquery.Selection) string {
		str, _ := s.Html()
		return strings.TrimSpace(str)
	}

	html2vfMut   = sync.Mutex{}
	html2vfCache = map[html2QueryTag]html2ValFunc{}
)

func html2AttrFunc(attr string) html2ValFunc {
	return func(s *goquery.Selection) string {
		str, _ := s.Attr(attr)
		return str
	}
}

func (tag html2QueryTag) valFunc() html2ValFunc {
	html2vfMut.Lock()
	defer html2vfMut.Unlock()

	if fn := html2vfCache[tag]; fn != nil {
		return fn
	}

	srcArr := strings.Split(string(tag), ",")
	if len(srcArr) < 2 {
		html2vfCache[tag] = html2TextVal
		return html2TextVal
	}

	src := srcArr[1]

	var f html2ValFunc
	switch {
	case src[0] == '[':
		// [someattr] will return value of .Attr("someattr")
		attr := src[1 : len(src)-1]
		f = html2AttrFunc(attr)
	case src == "html":
		f = html2Val
	case src == "text":
		f = html2TextVal
	default:
		f = html2TextVal
	}

	html2vfCache[tag] = f
	return f
}

// popVal should allow us to handle arbitrarily nested maps as well as the
// cleanly handling the possible of map[literal]literal by just delegating
// back to `html2UnmarshalByType`.
func (tag html2QueryTag) popVal() html2QueryTag {
	arr := strings.Split(string(tag), ",")
	if len(arr) < 2 {
		return tag
	}
	newA := []string{arr[0]}
	newA = append(newA, arr[2:]...)

	return html2QueryTag(strings.Join(newA, ","))
}

func html2WrapUnErr(err error, v reflect.Value) error {
	if err == nil {
		return nil
	}

	return &HtmlCannotUnmarshalErr{
		V:      v,
		Reason: html2CustomUnmarshalErr,
		Err:    err,
	}
}

// HtmlUnmarshalSelection will unmarshal a goquery.Selection into an interface
// appropriately and with goquery tags.
func HtmlUnmarshalSelection(s *goquery.Selection, face interface{}) error {
	v := reflect.ValueOf(face)

	// Must come before v.IsNil() else IsNil panics on NonPointer value
	if v.Kind() != reflect.Ptr {
		return &HtmlCannotUnmarshalErr{V: v, Reason: html2NonPointer}
	}

	if face == nil || v.IsNil() {
		return &HtmlCannotUnmarshalErr{V: v, Reason: html2NilValue}
	}

	u, v := reflectUnmarshalHTMLer(v)

	if u != nil {
		return html2WrapUnErr(u.UnmarshalHTML(s.Nodes), v)
	}

	return html2UnmarshalByType(s, v, "")
}

func html2UnmarshalByType(s *goquery.Selection, v reflect.Value, tag html2QueryTag) error {
	u, v := reflectUnmarshalHTMLer(v)

	if u != nil {
		return html2WrapUnErr(u.UnmarshalHTML(s.Nodes), v)
	}

	// Handle special cases where we can just set the value directly
	switch val := v.Interface().(type) {
	case []*html.Node:
		val = append(val, s.Nodes...)
		v.Set(reflect.ValueOf(val))
		return nil
	}

	t := v.Type()

	switch t.Kind() {
	case reflect.Struct:
		return html2UnmarshalStruct(s, v)
	case reflect.Slice:
		return html2UnmarshalSlice(s, v, tag)
	case reflect.Array:
		return html2UnmarshalArray(s, v, tag)
	case reflect.Map:
		return html2UnmarshalMap(s, v, tag)
	default:
		vf := tag.valFunc()
		str := vf(s)
		err := html2UnmarshalLiteral(str, v)
		if err != nil {
			return &HtmlCannotUnmarshalErr{
				V:      v,
				Reason: html2TypeConversionErr,
				Err:    err,
				Val:    str,
			}
		}
		return nil
	}
}

func html2UnmarshalLiteral(s string, v reflect.Value) error {
	t := v.Type()

	switch t.Kind() {
	case reflect.Interface:
		if t.NumMethod() == 0 {
			// For empty interfaces, just set to a string
			nv := reflect.New(reflect.TypeOf(s)).Elem()
			nv.Set(reflect.ValueOf(s))
			v.Set(nv)
		}
	case reflect.Bool:
		i, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(i)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(i)
	case reflect.Float32, reflect.Float64:
		i, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.SetFloat(i)
	case reflect.String:
		v.SetString(s)
	}
	return nil
}

func html2UnmarshalStruct(s *goquery.Selection, v reflect.Value) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		tag := html2QueryTag(t.Field(i).Tag.Get(html2TagName))

		if tag == html2Ignore {
			continue
		}

		// If tag is empty and the object doesn't implement Unmarshaler, skip
		if tag == "" {
			if u, _ := reflectUnmarshalHTMLer(v.Field(i)); u == nil {
				continue
			}
		}

		sel := tag.preprocess(s)
		if tag != "" {
			selStr := tag.selector(0)
			sel = sel.Find(selStr)
		}

		err := html2UnmarshalByType(sel, v.Field(i), tag)
		if err != nil {
			return &HtmlCannotUnmarshalErr{
				Reason:   html2TypeConversionErr,
				Err:      err,
				V:        v,
				FldOrIdx: t.Field(i).Name,
			}
		}
	}
	return nil
}

func html2UnmarshalArray(s *goquery.Selection, v reflect.Value, tag html2QueryTag) error {
	if v.Type().Len() != len(s.Nodes) {
		return &HtmlCannotUnmarshalErr{
			Reason: html2ArrayLenMismatch,
			V:      v,
		}
	}

	for i := 0; i < v.Type().Len(); i++ {
		err := html2UnmarshalByType(s.Eq(i), v.Index(i), tag)
		if err != nil {
			return &HtmlCannotUnmarshalErr{
				Reason:   html2TypeConversionErr,
				Err:      err,
				V:        v,
				FldOrIdx: i,
			}
		}
	}

	return nil
}

func html2UnmarshalSlice(s *goquery.Selection, v reflect.Value, tag html2QueryTag) error {
	slice := v
	eleT := v.Type().Elem()

	for i := 0; i < s.Length(); i++ {
		newV := reflect.New(TypeElem(eleT))

		err := html2UnmarshalByType(s.Eq(i), newV, tag)

		if err != nil {
			return &HtmlCannotUnmarshalErr{
				Reason:   html2TypeConversionErr,
				Err:      err,
				V:        v,
				FldOrIdx: i,
			}
		}

		if eleT.Kind() != reflect.Ptr {
			newV = newV.Elem()
		}

		v = reflect.Append(v, newV)
	}

	slice.Set(v)
	return nil
}

func html2ChildrenUntilMatch(s *goquery.Selection, sel string) *goquery.Selection {
	orig := s
	s = s.Children()
	for s.Length() != 0 && s.Filter(sel).Length() == 0 {
		s = s.Children()
	}
	if s.Length() == 0 {
		return orig
	}
	return s.Filter(sel)
}

func html2UnmarshalMap(s *goquery.Selection, v reflect.Value, tag html2QueryTag) error {
	// Make new map here because indirect for some Reason doesn't help us out
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	keyT, eleT := v.Type().Key(), v.Type().Elem()

	if tag.selector(1) == "" {
		// We need minimum one value selector to determine the map key
		return &HtmlCannotUnmarshalErr{
			Reason: html2MissingValSelector,
			V:      v,
		}
	}

	valTag := tag

	// Find children at the same level that match the given selector
	s = html2ChildrenUntilMatch(s, tag.selector(1))
	// Then augment the selector we will pass down to the next unmarshal step
	valTag = valTag.popVal()

	var err error
	s.EachWithBreak(func(_ int, subS *goquery.Selection) bool {
		newK, newV := reflect.New(TypeElem(keyT)), reflect.New(TypeElem(eleT))

		err = html2UnmarshalByType(subS, newK, tag)
		if err != nil {
			err = &HtmlCannotUnmarshalErr{
				Reason:   html2MapKeyUnmarshalErr,
				V:        v,
				Err:      err,
				FldOrIdx: newK.Interface(),
				Val:      valTag.valFunc()(subS),
			}
			return false
		}

		err = html2UnmarshalByType(subS, newV, valTag)
		if err != nil {
			return false
		}

		if eleT.Kind() != reflect.Ptr {
			newV = newV.Elem()
		}
		if keyT.Kind() != reflect.Ptr {
			newK = newK.Elem()
		}

		v.SetMapIndex(newK, newV)

		return true
	})

	if err != nil {
		return &HtmlCannotUnmarshalErr{
			Reason: html2TypeConversionErr,
			Err:    err,
			V:      v,
		}
	}

	return nil
}
