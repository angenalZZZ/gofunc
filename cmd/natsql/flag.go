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
	isTest   = false
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

	jsonFile = *flagTest
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
	subject = *flagName
	if subject == "" {
		subject = configInfo.Nats.Subscribe
	}
	if subject == "" {
		panic("the subscription name prefix can't be empty.")
	}

	cacheDir = filepath.Join(data.CurrentDir, subject)
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

// Init Subscribers
func configSubscribers() {
	if handlers == nil {
		handlers = make([]*handler, 0)
	}

	for _, s := range handlers {
		if s.sub != nil && s.sub.Running {
			f.DoneContext(s.Context)
			s.sub.Close()
		}
	}

	if jsr == nil {
		p := js.GoRuntimeParam{
			DbType: configInfo.Db.Type,
			DbConn: configInfo.Db.Conn,
		}
		jsr = js.NewRuntime(&p)
	}
	defer jsr.Clear()

	_, err := jsr.RunString(configInfo.Nats.Script)
	if err != nil {
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

		var itemSubj = subject
		var itemName, itemSpec, itemDir string
		if itemName, ok = objMap["name"].(string); !ok || itemName == "" {
			return
		}
		if itemSpec, ok = objMap["spec"].(string); !ok {
			return
		}
		if itemSpec == "+" {
			itemSubj = itemSubj + itemName
		} else {
			itemSubj = itemName
		}
		if itemFunc, ok := objMap["func"].(func(goja.FunctionCall) goja.Value); !ok {
			return
		} else {
			res := itemFunc(goja.FunctionCall{This: jsr.ToValue(obj)})
			if res == nil || res.String() == "" {
				itemDir = filepath.Join(cacheDir, itemName)
			} else {
				itemDir = filepath.Join(cacheDir, res.String())
			}
		}

		item := new(handler)
		ctx, wait := f.ContextWithWait(context.TODO())

		// Create a subscriber for Client Connect.
		conn := natClientConnect(false, itemSubj)
		sub := nat.NewSubscriberFastCache(conn, itemSubj, itemDir)
		sub.Hand = item.Handle

		// js global variable
		jso := make(map[string]interface{})
		jso["config"] = configInfo
		item.jso = jso

		// js runtime and register
		p1 := js.GoRuntimeParam{
			DbType:     configInfo.Db.Type,
			DbConn:     configInfo.Db.Conn,
			NatConn:    conn,
			NatSubject: itemSubj,
		}
		item.jsr = js.NewRuntime(&p1)

		// natS subscriber
		item.Context, item.sub = ctx, sub
		handlers = append(handlers, item)

		// Run natS subscriber.
		go sub.Run(wait)
	}
}

func runTest(hd *handler) {
	// Check Script
	if err := hd.CheckJs(configInfo.Nats.Script); err != nil {
		panic(err)
	}

	if !isTest {
		return
	}

	item, err := f.ReadFile(jsonFile)
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
	if err = hd.Handle(list); err != nil {
		panic(err)
	}

	nat.Log.Debug().Msg("test finished.")
	os.Exit(0)
}
