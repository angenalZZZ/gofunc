package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nats-io/nats.go"

	"github.com/dop251/goja"

	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/js"
	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
)

var (
	//flagMsgLimit = flag.Int("c", 100000000, "the nats-Limits for pending messages for this subscription")
	//flagBytesLimit = flag.Int("d", 4096, "the nats-Limits for a message's bytes for this subscription")
	flagConfig = flag.String("c", "natsql.yaml", "sets config file")
	flagTest   = flag.String("t", "", "sets json file and run SQL test")
	flagAddr   = flag.String("a", "", "the NatS-Server address")
	flagName   = flag.String("name", "", "the NatS-Subscription name prefix [required]")
	flagToken  = flag.String("token", "", "the NatS-Token auth string [required]")
	flagCred   = flag.String("cred", "", "the NatS-Cred file")
	flagCert   = flag.String("cert", "", "the NatS-TLS cert file")
	flagKey    = flag.String("key", "", "the NatS-TLS key file")
)

var (
	isTest = false
	// js test json data file
	jsonFile string
	// js runtime and register
	jsr *js.GoRuntime
)

// Your Arguments.
func initArgs() {
	flag.Usage = func() {
		fmt.Printf(" Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

// Check Arguments And Init Config.
func checkArgs() {
	if *flagConfig != "" {
		configFile = *flagConfig
	}

	if err := initConfig(); err != nil {
		panic(err)
	}

	if *flagAddr != "" {
		configInfo.Nats.Addr = *flagAddr
	}
	if *flagToken != "" {
		configInfo.Nats.Token = *flagToken
	}
	if *flagCred != "" {
		configInfo.Nats.Cred = *flagCred
	}
	if *flagCert != "" {
		configInfo.Nats.Cert = *flagCert
	}
	if *flagKey != "" {
		configInfo.Nats.Key = *flagKey
	}

	if *flagTest != "" {
		jsonFile = *flagTest
	}
	if jsonFile != "" {
		isTest = true
	}
	if isTest {
		configInfo.Log.Level = "debug"
	}

	if log.Log == nil {
		log.Log = log.Init(configInfo.Log)
	}
	if nat.Log == nil {
		nat.Log = log.Log
	}
	js.RunLogTimeFormat = configInfo.Log.TimeFormat

	// 全局订阅前缀:subject
	if *flagName != "" {
		subject = *flagName
	}
	if subject == "" {
		subject = configInfo.Nats.Subscribe
	}
	if subject == "" {
		panic("the subscription name prefix can't be empty.")
	}

	if cacheDir == "" {
		cacheDir = filepath.Join(data.CurrentDir, subject)
	}
	if f.PathExists(cacheDir) == false {
		panic("the cache disk directory is not found.")
	}

	if configInfo.Nats.Amount < 1 {
		configInfo.Nats.Amount = -1
	}
	if configInfo.Nats.Bulk < 1 {
		configInfo.Nats.Bulk = 1
	}
	if configInfo.Nats.Amount > 0 && configInfo.Nats.Amount < configInfo.Nats.Bulk {
		configInfo.Nats.Amount = configInfo.Nats.Bulk
	}
	if configInfo.Nats.Interval < 1 {
		configInfo.Nats.Interval = 1
	}

	nat.Log.Debug().Msgf("configuration complete")
}

// Create a subscriber for Client Connect.
func natClientConnect(isGlobal bool, subj string) (conn *nats.Conn) {
	var err error

	// NatS
	if isGlobal {
		nat.Subject = subject + subj
		nat.Conn, err = nat.New(subject, configInfo.Nats.Addr, configInfo.Nats.Cred, configInfo.Nats.Token, configInfo.Nats.Cert, configInfo.Nats.Key)
		if err != nil {
			nat.Log.Error().Msgf("[nats] failed connect to server: %v\n", err)
			os.Exit(1)
		}
		return nat.Conn
	}

	conn, err = nat.New(subj, configInfo.Nats.Addr, configInfo.Nats.Cred, configInfo.Nats.Token, configInfo.Nats.Cert, configInfo.Nats.Key)
	if err != nil {
		nat.Log.Error().Msgf("[nats] failed connect to server: %v\n", err)
		os.Exit(1)
	}
	return
}

// Init Subscribers And Handlers.
func createHandlers() {
	if handlers == nil {
		handlers = make([]*handler, 0)
	}

	stopHandlers() // Stop Subscribers And Handlers.

	if jsr == nil {
		p := js.GoRuntimeParam{
			CacheDir: cacheDir,
			DbType:   configInfo.Db.Type,
			DbConn:   configInfo.Db.Conn,
		}
		_ = f.Mkdir(p.CacheDir)
		jsr = js.NewRuntime(&p)
	}
	defer jsr.Clear()

	if _, err := jsr.RunString(configInfo.Nats.Script); err != nil {
		return
	}
	self := jsr.Runtime.Get("subscribe")
	if self == nil {
		return
	}
	objs, ok := self.Export().([]interface{})
	if !ok {
		return
	}

	handlers = make([]*handler, 0)

	for _, obj := range objs {
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			return
		}

		var itemName, itemSpec, itemSubj, itemDir string
		if itemName, ok = objMap["name"].(string); !ok || itemName == "" {
			return
		}
		if itemSpec, ok = objMap["spec"].(string); !ok {
			return
		}
		if itemSpec == "+" {
			itemSubj = subject + itemName
		} else {
			itemSubj = itemName
		}
		itemFunc, ok := objMap["func"].(func(goja.FunctionCall) goja.Value)
		if !ok {
			return
		}
		res := itemFunc(goja.FunctionCall{This: jsr.ToValue(obj)})
		if res == nil || res.String() == "" {
			itemDir = filepath.Join(cacheDir, itemName)
		} else {
			itemDir = filepath.Join(cacheDir, res.String())
		}

		h := new(handler)
		h.jsFile = filepath.Join(itemDir, "natsql.js")
		h.isScriptMod()
		if err := h.doScriptMod(); err != nil {
			return
		}

		// js global variable
		jso := make(map[string]interface{})
		jso["config"] = configInfo
		jso["configDir"] = itemDir
		jso["configJs"] = h.jsFile
		h.jso = jso

		// js runtime and register param
		jsp := js.GoRuntimeParam{
			CacheDir:   filepath.Join(cacheDir, itemName),
			DbType:     configInfo.Db.Type,
			DbConn:     configInfo.Db.Conn,
			NatSubject: itemSubj,
		}
		h.jsp = &jsp

		// natS subscriber
		if !isTest {
			// Create a subscriber for Client Connect
			conn := natClientConnect(false, itemSubj)
			h.jsp.NatConn = conn

			sub := nat.NewSubscriberFastCache(conn, itemSubj, itemDir)
			sub.Hand = h.Handle

			ctx, wait := f.ContextWithWait(context.TODO())
			h.Context, h.sub = ctx, sub
			// Run natS subscriber
			go sub.Run(wait)
		}

		handlers = append(handlers, h)
	}
}

// Stop Subscribers And Handlers.
func stopHandlers() {
	for _, h := range handlers {
		if h.sub != nil && h.sub.Running {
			f.DoneContext(h.Context)
			h.sub.Stop()
			h.Stop()
		}
	}
}

// Run script test.
func runTest() {
	if !isTest {
		return
	}

	createHandlers() // Init Subscribers And Handlers.

	for _, h := range handlers {
		itemDir := h.jso["configDir"].(string)
		filename := filepath.Join(itemDir, jsonFile)
		item, err := f.ReadFile(filename)
		if err != nil {
			panic(fmt.Errorf("test json file %q not opened: %s", jsonFile, err.Error()))
		}

		list, err := data.ListData(item)
		if err != nil {
			panic(err)
		}
		if len(list) == 0 {
			panic(fmt.Errorf("test json file can't be empty"))
		}

		nat.Log.Debug().Msgf("test json file: %q %d records\r\n", jsonFile, len(list))

		if subject == "" {
			subject = "test"
		}
		if err = h.Handle(list); err != nil {
			panic(err)
		}
	}

	nat.Log.Debug().Msg("test finished.")
	os.Exit(0)
}
