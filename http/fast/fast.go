// 🚀 Fast is an Express inspired web framework written in Go.

package fast

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// Version of current package
const Version = "1.0.0"

// Map is a shortcut for map[string]interface{}
type Map map[string]interface{}

// Fast denotes the Fast application.
type Fast struct {
	server   *fasthttp.Server // FastHTTP server
	routes   []*Route         // Route stack
	Settings *Settings        // Fast settings
}

// Settings holds is a struct holding the server settings
type Settings struct {
	// This will spawn multiple Go processes listening on the same port
	Prefork bool // default: false
	// Enable strict routing. When enabled, the router treats "/foo" and "/foo/" as different.
	StrictRouting bool // default: false
	// Enable case sensitivity. When enabled, "/Foo" and "/foo" are different routes.
	CaseSensitive bool // default: false
	// Enables the "Server: value" HTTP header.
	ServerHeader string // default: ""
	// Enables handler values to be immutable even if you return from handler
	Immutable bool // default: false
	// Max body size that the server accepts
	BodyLimit int // default: 4 * 1024 * 1024
	// Folder containing template files
	TemplateFolder string // default: ""
	// Template engine: html, amber, handlebars , mustache or pug
	TemplateEngine func(raw string, bind interface{}) (string, error) // default: nil
	// Extension for the template files
	TemplateExtension string // default: ""
	// The amount of time allowed to read the full request including body.
	ReadTimeout time.Duration // default: unlimited
	// The maximum duration before timing out writes of the response.
	WriteTimeout time.Duration // default: unlimited
	// The maximum amount of time to wait for the next request when keep-alive is enabled.
	IdleTimeout time.Duration // default: unlimited
}

// Group struct
type Group struct {
	prefix string
	app    *Fast
}

// New creates a new Fast named instance.
// You can pass optional settings when creating a new instance.
func New(settings ...*Settings) *Fast {
	// Parse arguments
	for _, v := range os.Args[1:] {
		if v == "-prefork" {
			isPrefork = true
		} else if v == "-child" {
			isChild = true
		}
	}
	// Create app
	app := new(Fast)
	// Create settings
	app.Settings = new(Settings)
	// Set default settings
	app.Settings.Prefork = isPrefork
	app.Settings.BodyLimit = 4 * 1024 * 1024
	// If settings exist, set defaults
	if len(settings) > 0 {
		app.Settings = settings[0] // Set custom settings
		if !app.Settings.Prefork { // Default to -prefork flag if false
			app.Settings.Prefork = isPrefork
		}
		if app.Settings.BodyLimit == 0 { // Default MaxRequestBodySize
			app.Settings.BodyLimit = 4 * 1024 * 1024
		}
		if app.Settings.Immutable { // Replace unsafe conversion funcs
			GetString = getStringImmutable
		}
	}
	return app
}

// Group is used for Routes with common prefix to define a new sub-router with optional middleware.
func (app *Fast) Group(prefix string, handlers ...func(*Ctx)) *Group {
	if len(handlers) > 0 {
		app.registerMethod("USE", prefix, handlers...)
	}
	return &Group{
		prefix: prefix,
		app:    app,
	}
}

// Static struct
type Static struct {
	// Transparently compresses responses if set to true
	// This works differently than the compression middleware
	// The server tries minimizing CPU usage by caching compressed files.
	// It adds ".gz" suffix to the original file name.
	// Optional. Default value false
	Compress bool
	// Enables byte range requests if set to true.
	// Optional. Default value false
	ByteRange bool
	// Enable directory browsing.
	// Optional. Default value false.
	Browse bool
	// Index file for serving a directory.
	// Optional. Default value "index.html".
	Index string
}

// Static registers a new route with path prefix to serve static files from the provided root directory.
func (app *Fast) Static(prefix, root string, config ...Static) *Fast {
	app.registerStatic(prefix, root, config...)
	return app
}

// Use registers a middleware route.
// Middleware matches requests beginning with the provided prefix.
// Providing a prefix is optional, it defaults to "/"
func (app *Fast) Use(args ...interface{}) *Fast {
	var path = ""
	var handlers []func(*Ctx)
	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case string:
			path = arg
		case func(*Ctx):
			handlers = append(handlers, arg)
		default:
			log.Fatalf("Invalid handler: %v", reflect.TypeOf(arg))
		}
	}
	app.registerMethod("USE", path, handlers...)
	return app
}

// Connect http method handler.
func (app *Fast) Connect(path string, handlers ...func(*Ctx)) *Fast {
	app.registerMethod("CONNECT", path, handlers...)
	return app
}

// Put http method handler.
func (app *Fast) Put(path string, handlers ...func(*Ctx)) *Fast {
	app.registerMethod("PUT", path, handlers...)
	return app
}

// Post http method handler.
func (app *Fast) Post(path string, handlers ...func(*Ctx)) *Fast {
	app.registerMethod("POST", path, handlers...)
	return app
}

// Delete http method handler.
func (app *Fast) Delete(path string, handlers ...func(*Ctx)) *Fast {
	app.registerMethod("DELETE", path, handlers...)
	return app
}

// Head http method handler.
func (app *Fast) Head(path string, handlers ...func(*Ctx)) *Fast {
	app.registerMethod("HEAD", path, handlers...)
	return app
}

// Patch http method handler.
func (app *Fast) Patch(path string, handlers ...func(*Ctx)) *Fast {
	app.registerMethod("PATCH", path, handlers...)
	return app
}

// Options http method handler.
func (app *Fast) Options(path string, handlers ...func(*Ctx)) *Fast {
	app.registerMethod("OPTIONS", path, handlers...)
	return app
}

// Trace http method handler.
func (app *Fast) Trace(path string, handlers ...func(*Ctx)) *Fast {
	app.registerMethod("TRACE", path, handlers...)
	return app
}

// Get http method handler.
func (app *Fast) Get(path string, handlers ...func(*Ctx)) *Fast {
	app.registerMethod("GET", path, handlers...)
	return app
}

// All matches all HTTP methods and complete paths
func (app *Fast) All(path string, handlers ...func(*Ctx)) *Fast {
	app.registerMethod("ALL", path, handlers...)
	return app
}

// Group is used for Routes with common prefix to define a new sub-router with optional middleware.
func (grp *Group) Group(prefix string, handlers ...func(*Ctx)) *Group {
	prefix = groupPaths(grp.prefix, prefix)
	if len(handlers) > 0 {
		grp.app.registerMethod("USE", prefix, handlers...)
	}
	return &Group{
		prefix: prefix,
		app:    grp.app,
	}
}

// Static : https://fiber.wiki/application#static
func (grp *Group) Static(prefix, root string, config ...Static) *Group {
	prefix = groupPaths(grp.prefix, prefix)
	grp.app.registerStatic(prefix, root, config...)
	return grp
}

// Use registers a middleware route.
// Middleware matches requests beginning with the provided prefix.
// Providing a prefix is optional, it defaults to "/"
func (grp *Group) Use(args ...interface{}) *Group {
	var path = ""
	var handlers []func(*Ctx)
	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case string:
			path = arg
		case func(*Ctx):
			handlers = append(handlers, arg)
		default:
			log.Fatalf("Invalid Use() arguments, must be (prefix, handler) or (handler)")
		}
	}
	grp.app.registerMethod("USE", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// Connect http method handler.
func (grp *Group) Connect(path string, handlers ...func(*Ctx)) *Group {
	grp.app.registerMethod("CONNECT", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// Put http method handler.
func (grp *Group) Put(path string, handlers ...func(*Ctx)) *Group {
	grp.app.registerMethod("PUT", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// Post http method handler.
func (grp *Group) Post(path string, handlers ...func(*Ctx)) *Group {
	grp.app.registerMethod("POST", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// Delete http method handler.
func (grp *Group) Delete(path string, handlers ...func(*Ctx)) *Group {
	grp.app.registerMethod("DELETE", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// Head http method handler.
func (grp *Group) Head(path string, handlers ...func(*Ctx)) *Group {
	grp.app.registerMethod("HEAD", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// Patch http method handler.
func (grp *Group) Patch(path string, handlers ...func(*Ctx)) *Group {
	grp.app.registerMethod("PATCH", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// Options http method handler.
func (grp *Group) Options(path string, handlers ...func(*Ctx)) *Group {
	grp.app.registerMethod("OPTIONS", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// Trace http method handler.
func (grp *Group) Trace(path string, handlers ...func(*Ctx)) *Group {
	grp.app.registerMethod("TRACE", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// Get http method handler.
func (grp *Group) Get(path string, handlers ...func(*Ctx)) *Group {
	grp.app.registerMethod("GET", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// All matches all HTTP methods and complete paths
func (grp *Group) All(path string, handlers ...func(*Ctx)) *Group {
	grp.app.registerMethod("ALL", groupPaths(grp.prefix, path), handlers...)
	return grp
}

// Listen serves HTTP requests from the given addr or port.
// You can pass an optional *tls.Config to enable TLS.
func (app *Fast) Listen(address interface{}, tlsconfig ...*tls.Config) error {
	addr, ok := address.(string)
	if !ok {
		port, ok := address.(int)
		if !ok {
			port = 80
		}
		addr = strconv.Itoa(port)
	}
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}
	// Create fasthttp server
	app.server = app.newServer()
	// Print listening message
	if !isChild {
		fmt.Printf("Fast v%s listening on %s\n", Version, addr)
	}
	var ln net.Listener
	var err error
	// Prefork enabled
	if app.Settings.Prefork && runtime.NumCPU() > 1 {
		if ln, err = app.prefork(addr); err != nil {
			return err
		}
	} else {
		if ln, err = net.Listen("tcp4", addr); err != nil {
			return err
		}
	}

	// TLS config
	if len(tlsconfig) > 0 {
		ln = tls.NewListener(ln, tlsconfig[0])
	}
	return app.server.Serve(ln)
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
// Shutdown works by first closing all open listeners and then waiting indefinitely for all connections to return to idle and then shut down.
//
// When Shutdown is called, Serve, ListenAndServe, and ListenAndServeTLS immediately return nil.
// Make sure the program doesn't exit and waits instead for Shutdown to return.
//
// Shutdown does not close keepalive connections so its recommended to set ReadTimeout to something else than 0.
func (app *Fast) Shutdown() error {
	return app.server.Shutdown()
}

// Test is used for internal debugging by passing a *http.Request.
// Timeout is optional and defaults to 200ms, -1 to 10s.
func (app *Fast) Test(request *http.Request, msTimeout ...int) (*http.Response, error) {
	timeout := 200
	if len(msTimeout) > 0 {
		timeout = msTimeout[0]
	}
	if timeout < 0 {
		timeout = 10000
	}
	// Dump raw http request
	dump, err := httputil.DumpRequest(request, true)
	if err != nil {
		return nil, err
	}
	// Setup server
	app.server = app.newServer()
	// Create conn
	conn := new(testConn)
	// Write raw http request
	if _, err = conn.r.Write(dump); err != nil {
		return nil, err
	}
	// Serve conn to server
	channel := make(chan error)
	go func() {
		channel <- app.server.ServeConn(conn)
	}()
	// Wait for callback
	select {
	case err := <-channel:
		if err != nil {
			return nil, err
		}
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return nil, fmt.Errorf("timeout error")
	}
	// Read response
	buffer := bufio.NewReader(&conn.w)
	// Convert raw http response to *http.Response
	resp, err := http.ReadResponse(buffer, request)
	if err != nil {
		return nil, err
	}
	// Return *http.Response
	return resp, nil
}

// Sharding: https://www.nginx.com/blog/socket-sharding-nginx-release-1-9-1/
func (app *Fast) prefork(address string) (ln net.Listener, err error) {
	// Master proc
	if !isChild {
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			return ln, err
		}
		tcplistener, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return ln, err
		}
		fl, err := tcplistener.File()
		if err != nil {
			return ln, err
		}
		files := []*os.File{fl}
		childs := make([]*exec.Cmd, runtime.NumCPU()/2)
		// #nosec G204
		for i := range childs {
			childs[i] = exec.Command(os.Args[0], append(os.Args[1:], "-prefork", "-child")...)
			childs[i].Stdout = os.Stdout
			childs[i].Stderr = os.Stderr
			childs[i].ExtraFiles = files
			if err := childs[i].Start(); err != nil {
				return ln, err
			}
		}

		for _, child := range childs {
			if err := child.Wait(); err != nil {
				return ln, err
			}
		}
		os.Exit(0)
	} else {
		// 1 core per child
		runtime.GOMAXPROCS(1)
		ln, err = net.FileListener(os.NewFile(3, ""))
	}
	return ln, err
}

func (app *Fast) newServer() *fasthttp.Server {
	return &fasthttp.Server{
		Handler:               app.handler,
		Name:                  app.Settings.ServerHeader,
		MaxRequestBodySize:    app.Settings.BodyLimit,
		NoDefaultServerHeader: app.Settings.ServerHeader == "",
		ReadTimeout:           app.Settings.ReadTimeout,
		WriteTimeout:          app.Settings.WriteTimeout,
		IdleTimeout:           app.Settings.IdleTimeout,
		LogAllErrors:          false,
		ErrorHandler: func(ctx *fasthttp.RequestCtx, err error) {
			if err.Error() == "body size exceeds the given limit" {
				ctx.Response.SetStatusCode(413)
				ctx.Response.SetBodyString("Request Entity Too Large")
			} else {
				ctx.Response.SetStatusCode(400)
				ctx.Response.SetBodyString("Bad Request")
			}
		},
	}
}
