// 🚀 Fast is an Express inspired web framework written in Go.

package fast

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

// Ctx represents the Context which hold the HTTP request and response.
// It has methods for the request query string, parameters, body, HTTP headers and so on.
type Ctx struct {
	C      *fasthttp.RequestCtx // Reference to *fasthttp.RequestCtx
	app    *Fast                // Reference to *Fast
	route  *Route               // Reference to *Route
	index  int                  // Index of the current stack
	method string               // HTTP method
	path   string               // HTTP path
	values []string             // Route parameter values
	err    error                // Contains error if catched
}

// Range struct
type Range struct {
	Type   string
	Ranges []struct {
		Start int
		End   int
	}
}

// Cookie struct
type Cookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	Expires  time.Time
	Secure   bool
	HTTPOnly bool
	SameSite string
}

// Ctx pool
var poolCtx = sync.Pool{
	New: func() interface{} {
		return new(Ctx)
	},
}

// Acquire Ctx from pool
func acquireCtx(fctx *fasthttp.RequestCtx) *Ctx {
	ctx := poolCtx.Get().(*Ctx)
	ctx.index = -1
	ctx.path = getString(fctx.URI().Path())
	ctx.method = getString(fctx.Request.Header.Method())
	ctx.C = fctx
	return ctx
}

// Return Ctx to pool
func releaseCtx(ctx *Ctx) {
	ctx.route = nil
	ctx.values = nil
	ctx.C = nil
	ctx.err = nil
	poolCtx.Put(ctx)
}

// Accepts checks if the specified extensions or content types are acceptable.
func (ctx *Ctx) Accepts(offers ...string) (offer string) {
	if len(offers) == 0 {
		return ""
	}
	h := ctx.Get("Accept")
	if h == "" {
		return offers[0]
	}

	specs := strings.Split(h, ",")
	for _, value := range offers {
		mimeType := getMIME(value)
		for _, spec := range specs {
			spec = strings.TrimSpace(spec)
			if strings.HasPrefix(spec, "*/*") {
				return value
			}

			if strings.HasPrefix(spec, mimeType) {
				return value
			}

			if strings.Contains(spec, "/*") {
				if strings.HasPrefix(spec, strings.Split(mimeType, "/")[0]) {
					return value
				}
			}
		}
	}
	return ""
}

// AcceptsCharsets checks if the specified charset is acceptable.
func (ctx *Ctx) AcceptsCharsets(offers ...string) (offer string) {
	if len(offers) == 0 {
		return ""
	}

	h := ctx.Get("Accept-Charset")
	if h == "" {
		return offers[0]
	}

	specs := strings.Split(h, ",")
	for _, value := range offers {
		for _, spec := range specs {

			spec = strings.TrimSpace(spec)
			if strings.HasPrefix(spec, "*") {
				return value
			}
			if strings.HasPrefix(spec, value) {
				return value
			}
		}
	}
	return ""
}

// AcceptsEncodings checks if the specified encoding is acceptable.
func (ctx *Ctx) AcceptsEncodings(offers ...string) (offer string) {
	if len(offers) == 0 {
		return ""
	}

	h := ctx.Get("Accept-Encoding")
	if h == "" {
		return offers[0]
	}

	specs := strings.Split(h, ",")
	for _, value := range offers {
		for _, spec := range specs {
			spec = strings.TrimSpace(spec)
			if strings.HasPrefix(spec, "*") {
				return value
			}
			if strings.HasPrefix(spec, value) {
				return value
			}
		}
	}
	return ""
}

// AcceptsLanguages checks if the specified language is acceptable.
func (ctx *Ctx) AcceptsLanguages(offers ...string) (offer string) {
	if len(offers) == 0 {
		return ""
	}
	h := ctx.Get("Accept-Language")
	if h == "" {
		return offers[0]
	}

	specs := strings.Split(h, ",")
	for _, value := range offers {
		for _, spec := range specs {
			spec = strings.TrimSpace(spec)
			if strings.HasPrefix(spec, "*") {
				return value
			}
			if strings.HasPrefix(spec, value) {
				return value
			}
		}
	}
	return ""
}

// Append the specified value to the HTTP response header field.
// If the header is not already set, it creates the header with the specified value.
func (ctx *Ctx) Append(field string, values ...string) {
	if len(values) == 0 {
		return
	}
	h := getString(ctx.C.Response.Header.Peek(field))
	for i := range values {
		if h == "" {
			h += values[i]
		} else {
			h += ", " + values[i]
		}
	}
	ctx.Set(field, h)
}

// Attachment sets the HTTP response Content-Disposition header field to attachment.
func (ctx *Ctx) Attachment(name ...string) {
	if len(name) > 0 {
		filename := filepath.Base(name[0])
		ctx.Type(filepath.Ext(filename))
		ctx.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		return
	}
	ctx.Set("Content-Disposition", "attachment")
}

// BaseURL returns (protocol + host).
func (ctx *Ctx) BaseURL() string {
	return ctx.Protocol() + "://" + ctx.Hostname()
}

// Body contains the raw body submitted in a POST request.
// If a key is provided, it returns the form value
func (ctx *Ctx) Body(key ...string) string {
	// Return request body
	if len(key) == 0 {
		return getString(ctx.C.Request.Body())
	}
	// Return post value by key
	if len(key) > 0 {
		return getString(ctx.C.Request.PostArgs().Peek(key[0]))
	}
	return ""
}

// BodyParser binds the request body to a struct.
// It supports decoding the following content types based on the Content-Type header:
// application/json, application/xml, application/x-www-form-urlencoded, multipart/form-data
func (ctx *Ctx) BodyParser(out interface{}) error {
	// Query Params
	ct := getString(ctx.C.Request.Header.ContentType())
	// application/json
	if strings.HasPrefix(ct, "application/json") {
		return jsoniter.Unmarshal(ctx.C.Request.Body(), out)
	}
	// application/xml text/xml
	if strings.HasPrefix(ct, "application/xml") || strings.HasPrefix(ct, "text/xml") {
		return xml.Unmarshal(ctx.C.Request.Body(), out)
	}
	// application/x-www-form-urlencoded
	if strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		data, err := url.ParseQuery(getString(ctx.C.PostBody()))
		if err != nil {
			return err
		}
		return schemaDecoder.Decode(out, data)
	}
	// multipart/form-data
	if strings.HasPrefix(ct, "multipart/form-data") {
		data, err := ctx.C.MultipartForm()
		if err != nil {
			return err
		}
		return schemaDecoder.Decode(out, data.Value)
	}
	return fmt.Errorf("BodyParser: cannot parse content-type: %v", ct)
}

// ClearCookie expires a specific cookie by key.
// If no key is provided it expires all cookies.
func (ctx *Ctx) ClearCookie(key ...string) {
	if len(key) > 0 {
		for i := range key {
			ctx.C.Response.Header.DelClientCookie(key[i])
		}
		return
	}
	//ctx.C.Response.Header.DelAllCookies()
	ctx.C.Request.Header.VisitAllCookie(func(k, v []byte) {
		ctx.C.Response.Header.DelClientCookie(getString(k))
	})
}

// Cookie sets a cookie by passing a cookie struct
func (ctx *Ctx) Cookie(cookie *Cookie) {
	c := &fasthttp.Cookie{}
	c.SetKey(cookie.Name)
	c.SetValue(cookie.Value)
	c.SetPath(cookie.Path)
	c.SetDomain(cookie.Domain)
	c.SetExpire(cookie.Expires)
	c.SetSecure(cookie.Secure)
	if cookie.Secure {
		// Secure must be paired with SameSite=None
		c.SetSameSite(fasthttp.CookieSameSiteNoneMode)
	}
	c.SetHTTPOnly(cookie.HTTPOnly)
	switch strings.ToLower(cookie.SameSite) {
	case "lax":
		c.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	case "strict":
		c.SetSameSite(fasthttp.CookieSameSiteStrictMode)
	case "none":
		c.SetSameSite(fasthttp.CookieSameSiteNoneMode)
		// Secure must be paired with SameSite=None
		c.SetSecure(true)
	default:
		c.SetSameSite(fasthttp.CookieSameSiteDisabled)
	}
	ctx.C.Response.Header.SetCookie(c)
}

// Cookies is used for getting a cookie value by key
func (ctx *Ctx) Cookies(key ...string) (value string) {
	if len(key) == 0 {
		return ctx.Get("Cookie")
	}
	return getString(ctx.C.Request.Header.Cookie(key[0]))
}

// Download transfers the file from path as an attachment.
// Typically, browsers will prompt the user for download.
// By default, the Content-Disposition header filename= parameter is the filepath (this typically appears in the browser dialog).
// Override this default with the filename parameter.
func (ctx *Ctx) Download(file string, name ...string) {
	filename := filepath.Base(file)

	if len(name) > 0 {
		filename = name[0]
	}

	ctx.Set("Content-Disposition", "attachment; filename="+filename)
	ctx.SendFile(file)
}

// Error contains the error information passed via the Next(err) method.
func (ctx *Ctx) Error() error {
	return ctx.err
}

// Format performs content-negotiation on the Accept HTTP header.
// It uses Accepts to select a proper format.
// If the header is not specified or there is no proper format, text/plain is used.
func (ctx *Ctx) Format(body interface{}) {
	var b string
	accept := ctx.Accepts("html", "json")

	switch val := body.(type) {
	case string:
		b = val
	case []byte:
		b = getString(val)
	default:
		if t, ok := val.(fmt.Stringer); ok {
			b = t.String()
		} else {
			b = fmt.Sprintf("%v", val)
		}
	}
	switch accept {
	case "html":
		ctx.SendString(b)
	case "json":
		if err := ctx.JSON(body); err != nil {
			log.Println("Format: error serializing json ", err)
		}
	default:
		ctx.SendString(b)
	}
}

// FormFile returns the first file by key from a MultipartForm.
func (ctx *Ctx) FormFile(key string) (*multipart.FileHeader, error) {
	return ctx.C.FormFile(key)
}

// FormValue returns the first value by key from a MultipartForm.
func (ctx *Ctx) FormValue(key string) (value string) {
	return getString(ctx.C.FormValue(key))
}

// Fresh is not implemented yet, pull requests are welcome!
func (ctx *Ctx) Fresh() bool {
	return false
}

// Get returns the HTTP request header specified by field.
// Field names are case-insensitive
func (ctx *Ctx) Get(key string) (value string) {
	if key == "referrer" {
		key = "referer"
	}
	return getString(ctx.C.Request.Header.Peek(key))
}

// Hostname contains the hostname derived from the Host HTTP header.
func (ctx *Ctx) Hostname() string {
	return getString(ctx.C.URI().Host())
}

// IP returns the remote IP address of the request.
func (ctx *Ctx) IP() string {
	return ctx.C.RemoteIP().String()
}

// IPs returns an string slice of IP addresses specified in the X-Forwarded-For request header.
func (ctx *Ctx) IPs() []string {
	ips := strings.Split(ctx.Get("X-Forwarded-For"), ",")
	for i := range ips {
		ips[i] = strings.TrimSpace(ips[i])
	}
	return ips
}

// Is returns the matching content type,
// if the incoming request’s Content-Type HTTP header field matches the MIME type specified by the type parameter
func (ctx *Ctx) Is(extension string) (match bool) {
	if extension[0] != '.' {
		extension = "." + extension
	}

	items, _ := mime.ExtensionsByType(ctx.Get("Content-Type"))
	if len(items) > 0 {
		for _, item := range items {
			if item == extension {
				return true
			}
		}
	}
	return
}

// JSON converts any interface or string to JSON using Jsoniter.
// This method also sets the content header to application/json.
func (ctx *Ctx) JSON(json interface{}) error {
	// Get stream from pool
	stream := jsonParser.BorrowStream(nil)
	defer jsonParser.ReturnStream(stream)
	// Write struct to stream
	stream.WriteVal(&json)
	// Check for errors
	if stream.Error != nil {
		return stream.Error
	}
	// Set http headers
	ctx.C.Response.Header.SetContentType("application/json")
	ctx.C.Response.SetBodyString(getString(stream.Buffer()))
	// Success!
	return nil
}

// JSONP sends a JSON response with JSONP support.
// This method is identical to JSON, except that it opts-in to JSONP callback support.
// By default, the callback name is simply callback.
func (ctx *Ctx) JSONP(json interface{}, callback ...string) error {
	// Get stream from pool
	stream := jsonParser.BorrowStream(nil)
	defer jsonParser.ReturnStream(stream)
	// Write struct to stream
	stream.WriteVal(&json)
	// Check for errors
	if stream.Error != nil {
		return stream.Error
	}

	str := "callback("
	if len(callback) > 0 {
		str = callback[0] + "("
	}
	str += getString(stream.Buffer()) + ");"

	ctx.Set("X-Content-Type-Options", "nosniff")
	ctx.C.Response.Header.SetContentType("application/javascript")
	ctx.C.Response.SetBodyString(str)
	return nil
}

// Links joins the links followed by the property to populate the response’s Link HTTP header field.
func (ctx *Ctx) Links(link ...string) {
	h := ""
	for i, l := range link {
		if i%2 == 0 {
			h += "<" + l + ">"
		} else {
			h += `; rel="` + l + `",`
		}
	}

	if len(link) > 0 {
		h = strings.TrimSuffix(h, ",")
		ctx.Set("Link", h)
	}
}

// Locals makes it possible to pass interface{} values under string keys scoped to the request
// and therefore available to all following routes that match the request.
func (ctx *Ctx) Locals(key string, value ...interface{}) (val interface{}) {
	if len(value) == 0 {
		return ctx.C.UserValue(key)
	}
	ctx.C.SetUserValue(key, value[0])
	return value[0]
}

// Location sets the response Location HTTP header to the specified path parameter.
func (ctx *Ctx) Location(path string) {
	ctx.Set("Location", path)
}

// Method contains a string corresponding to the HTTP method of the request: GET, POST, PUT and so on.
func (ctx *Ctx) Method(override ...string) string {
	if len(override) > 0 {
		ctx.method = override[0]
	}
	return ctx.method
}

// MultipartForm parse form entries from binary.
// This returns a map[string][]string, so given a key the value will be a string slice.
func (ctx *Ctx) MultipartForm() (*multipart.Form, error) {
	return ctx.C.MultipartForm()
}

// Next executes the next method in the stack that matches the current route.
// You can pass an optional error for custom error handling.
func (ctx *Ctx) Next(err ...error) {
	ctx.route = nil
	ctx.values = nil
	if len(err) > 0 {
		ctx.err = err[0]
	}
	ctx.app.nextRoute(ctx)
}

// OriginalURL contains the original request URL.
func (ctx *Ctx) OriginalURL() string {
	return getString(ctx.C.Request.Header.RequestURI())
}

// Params is used to get the route parameters.
// Defaults to empty string "", if the param doesn't exist.
func (ctx *Ctx) Params(key string) (value string) {
	if ctx.route.Params == nil {
		return
	}
	for i := 0; i < len(ctx.route.Params); i++ {
		if (ctx.route.Params)[i] == key {
			return ctx.values[i]
		}
	}
	return
}

// Path returns the path part of the request URL.
// Optionally, you could override the path.
func (ctx *Ctx) Path(override ...string) string {
	if len(override) > 0 {
		// Non strict routing
		if !ctx.app.Settings.StrictRouting && len(override[0]) > 1 {
			override[0] = strings.TrimRight(override[0], "/")
		}
		// Not case sensitive
		if !ctx.app.Settings.CaseSensitive {
			override[0] = strings.ToLower(override[0])
		}
		ctx.path = override[0]
	}
	return ctx.path
}

// Protocol contains the request protocol string: http or https for TLS requests.
func (ctx *Ctx) Protocol() string {
	if ctx.C.IsTLS() {
		return "https"
	}
	return "http"
}

// Query returns the query string parameter in the url.
func (ctx *Ctx) Query(key string) (value string) {
	return getString(ctx.C.QueryArgs().Peek(key))
}

// Range returns a struct containing the type and a slice of ranges.
func (ctx *Ctx) Range(size int) (rangeData Range, err error) {
	rangeStr := string(ctx.C.Request.Header.Peek("Range"))
	if rangeStr == "" || !strings.Contains(rangeStr, "=") {
		return rangeData, fmt.Errorf("malformed range header string")
	}
	data := strings.Split(rangeStr, "=")
	rangeData.Type = data[0]
	arr := strings.Split(data[1], ",")
	for i := 0; i < len(arr); i++ {
		item := strings.Split(arr[i], "-")
		if len(item) == 1 {
			return rangeData, fmt.Errorf("malformed range header string")
		}
		start, startErr := strconv.Atoi(item[0])
		end, endErr := strconv.Atoi(item[1])
		if startErr != nil { // -nnn
			start = size - end
			end = size - 1
		} else if endErr != nil { // nnn-
			end = size - 1
		}
		if end > size-1 { // limit last-byte-pos to current length
			end = size - 1
		}
		if start > end || start < 0 {
			continue
		}
		rangeData.Ranges = append(rangeData.Ranges, struct {
			Start int
			End   int
		}{
			start,
			end,
		})
	}
	if len(rangeData.Ranges) < 1 {
		return rangeData, fmt.Errorf("unsatisfiable range")
	}
	return rangeData, nil
}

// Redirect to the URL derived from the specified path, with specified status.
// If status is not specified, status defaults to 302 Found
func (ctx *Ctx) Redirect(path string, status ...int) {
	code := 302
	if len(status) > 0 {
		code = status[0]
	}

	ctx.Set("Location", path)
	ctx.C.Response.SetStatusCode(code)
}

// Render a template with data and sends a text/html response.
// We support the following engines: html, amber, handlebars, mustache, pug
func (ctx *Ctx) Render(file string, bind interface{}) error {
	var err error
	var raw []byte
	var html string

	if ctx.app.Settings.TemplateFolder != "" {
		file = filepath.Join(ctx.app.Settings.TemplateFolder, file)
	}
	if ctx.app.Settings.TemplateExtension != "" {
		file = file + ctx.app.Settings.TemplateExtension
	}
	if raw, err = ioutil.ReadFile(filepath.Clean(file)); err != nil {
		return err
	}
	if ctx.app.Settings.TemplateEngine != nil {
		// Custom template engine
		// https://github.com/valyala/quicktemplate
		if html, err = ctx.app.Settings.TemplateEngine(getString(raw), bind); err != nil {
			return err
		}
	} else {
		// Default template engine
		// https://golang.org/pkg/text/template/
		var buf bytes.Buffer
		var tmpl *template.Template

		if tmpl, err = template.New("").Parse(getString(raw)); err != nil {
			return err
		}
		if err = tmpl.Execute(&buf, bind); err != nil {
			return err
		}
		html = buf.String()
	}
	ctx.Set("Content-Type", "text/html")
	ctx.SendString(html)
	return err
}

// Route returns the matched Route struct.
func (ctx *Ctx) Route() *Route {
	return ctx.route
}

// SaveFile saves any multipart file to disk.
func (ctx *Ctx) SaveFile(file *multipart.FileHeader, path string) error {
	return fasthttp.SaveMultipartFile(file, path)
}

// Secure returns a boolean property, that is true, if a TLS connection is established.
func (ctx *Ctx) Secure() bool {
	return ctx.C.IsTLS()
}

// Send sets the HTTP response body. The Send body can be of any type.
func (ctx *Ctx) Send(bodies ...interface{}) {
	if len(bodies) > 0 {
		ctx.C.Response.SetBodyString("")
	}
	for i := range bodies {
		switch body := bodies[i].(type) {
		case string:
			ctx.C.Response.AppendBodyString(body)
		case []byte:
			ctx.C.Response.AppendBody(body) // .AppendBodyString(getString(body))
		default:
			if t, ok := body.(fmt.Stringer); ok {
				ctx.C.Response.AppendBodyString(t.String())
			} else {
				ctx.C.Response.AppendBodyString(fmt.Sprintf("%v", body))
			}
		}
	}
}

// SendBytes sets the HTTP response body for []byte types
// This means no type assertion, recommended for faster performance
func (ctx *Ctx) SendBytes(body []byte) {
	ctx.C.Response.SetBodyString(getString(body))
}

// SendFile transfers the file from the given path.
// The file is compressed by default
// Sets the Content-Type response HTTP header field based on the filenames extension.
func (ctx *Ctx) SendFile(file string, noCompression ...bool) {
	// Disable gzip
	if len(noCompression) > 0 && noCompression[0] {
		fasthttp.ServeFileUncompressed(ctx.C, file)
		return
	}
	fasthttp.ServeFile(ctx.C, file)
}

// SendStatus sets the HTTP status code and if the response body is empty,
// it sets the correct status message in the body.
func (ctx *Ctx) SendStatus(status int) {
	ctx.C.Response.SetStatusCode(status)
	// Only set status body when there is no response body
	if len(ctx.C.Response.Body()) == 0 {
		ctx.C.Response.SetBodyString(statusMessages[status])
	}
}

// SendString sets the HTTP response body for string types
// This means no type assertion, recommended for faster performance
func (ctx *Ctx) SendString(body string) {
	ctx.C.Response.SetBodyString(body)
}

// Set sets the response’s HTTP header field to the specified key, value.
func (ctx *Ctx) Set(key string, val string) {
	ctx.C.Response.Header.Set(key, val)
}

// Subdomains returns a string slive of subdomains in the domain name of the request.
// The subdomain offset, which defaults to 2, is used for determining the beginning of the subdomain segments.
func (ctx *Ctx) Subdomains(offset ...int) []string {
	o := 2
	if len(offset) > 0 {
		o = offset[0]
	}
	subdomains := strings.Split(ctx.Hostname(), ".")
	subdomains = subdomains[:len(subdomains)-o]
	return subdomains
}

// Stale is not implemented yet, pull requests are welcome!
func (ctx *Ctx) Stale() bool {
	return !ctx.Fresh()
}

// Status sets the HTTP status for the response.
// This method is chainable.
func (ctx *Ctx) Status(status int) *Ctx {
	ctx.C.Response.SetStatusCode(status)
	return ctx
}

// Token gets token from request(query url, post args, header authorization).
func (ctx *Ctx) Token(key ...string) string {
	k := "token"
	if len(key) > 0 {
		k = key[0]
	}
	switch ctx.method {
	case "POST", "PUT":
		if ctx.C.Request.PostArgs().Has(k) {
			return getString(ctx.C.Request.PostArgs().Peek(k))
		}
	}
	if token := ctx.C.Request.Header.Peek("Authorization"); token != nil && len(token) > 8 {
		t := strings.Split(getString(token), " ")
		return t[len(t)-1]
	}
	if ctx.C.QueryArgs().Has(k) {
		return getString(ctx.C.QueryArgs().Peek(k))
	}
	return ""
}

// Type sets the Content-Type HTTP header to the MIME type specified by the file extension.
func (ctx *Ctx) Type(ext string) *Ctx {
	ctx.C.Response.Header.SetContentType(getMIME(ext))
	return ctx
}

// Vary adds the given header field to the Vary response header.
// This will append the header, if not already listed, otherwise leaves it listed in the current location.
func (ctx *Ctx) Vary(fields ...string) {
	if len(fields) == 0 {
		return
	}

	h := getString(ctx.C.Response.Header.Peek("Vary"))
	for i := range fields {
		if h == "" {
			h += fields[i]
		} else {
			h += ", " + fields[i]
		}
	}

	ctx.Set("Vary", h)
}

// Write appends any input to the HTTP body response.
func (ctx *Ctx) Write(bodies ...interface{}) {
	for i := range bodies {
		switch body := bodies[i].(type) {
		case string:
			ctx.C.Response.AppendBodyString(body)
		case []byte:
			ctx.C.Response.AppendBody(body) // .AppendBodyString(getString(body))
		default:
			if t, ok := body.(fmt.Stringer); ok {
				ctx.C.Response.AppendBodyString(t.String())
			} else {
				ctx.C.Response.AppendBodyString(fmt.Sprintf("%v", body))
			}
		}
	}
}

// XHR returns a Boolean property, that is true, if the request’s X-Requested-With header field is XMLHttpRequest,
// indicating that the request was issued by a client library (such as jQuery).
func (ctx *Ctx) XHR() bool {
	return ctx.Get("X-Requested-With") == "XMLHttpRequest"
}

// XSSProtection X-XSS-Protection...
func (ctx *Ctx) XSSProtection() {
	ctx.Set("X-Content-Type-Options", "nosniff")
	ctx.Set("X-Frame-Options", "SAMEORIGIN") // or DENY
	ctx.Set("X-XSS-Protection", "1; mode=block")
	if ctx.C.IsTLS() {
		ctx.Set("Strict-Transport-Security", "max-age=31536000")
	}
	// Also consider adding Content-Security-Policy headers
	// c.Header("Content-Security-Policy", "script-src 'self' https://cdnjs.cloudflare.com")
}
