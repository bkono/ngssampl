package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carlmjohnson/exitcode"
	"github.com/carlmjohnson/flagext"
	nats "github.com/nats-io/nats.go"
)

const (
	AppName = "ngssampl"
	subject = "sample.event"
)

func main() {
	exitcode.Exit(CLI(os.Args[1:]))
}

func CLI(args []string) error {
	var app appEnv
	err := app.ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	if err = app.Exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

func (app *appEnv) ParseArgs(args []string) error {
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	fl.BoolVar(&app.pub, "pub", false, "continuous publish mode")
	fl.BoolVar(&app.sub, "sub", false, "subscribe to the test topic")
	fl.StringVar(&app.creds, "creds", "", "path to user credentials for nats")

	app.Logger = log.New(os.Stderr, AppName+" ", log.LstdFlags)
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), `ngssampl - 

Usage:

	ngssampl [options]

Options:
`)
		fl.PrintDefaults()
		fmt.Fprintln(fl.Output(), "")
	}
	if err := fl.Parse(args); err != nil {
		return err
	}
	if err := flagext.ParseEnv(fl, AppName); err != nil {
		return err
	}

	if !app.pub && !app.sub {
		return fmt.Errorf("at least one of pub or sub must be set")
	}

	if app.creds == "" {
		return fmt.Errorf("nats creds are required, please set the creds value to the credentials path")
	}

	return nil
}

type appEnv struct {
	pub   bool
	sub   bool
	creds string
	*log.Logger
}

func (app *appEnv) Exec() error {
	app.Println("starting")
	nc, err := nats.Connect("nats://connect.ngs.global", nats.UserCredentials(app.creds))
	if err != nil {
		return err
	}

	defer func() { app.Println("done") }()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if app.sub {
		app.Println("-sub provided. starting listener")
		nc.Subscribe(subject, func(msg *nats.Msg) {
			rcvd := time.Now().UnixNano() / 1000000
			sent := int64(binary.LittleEndian.Uint64(msg.Data))
			app.Printf("---> received ( %v ) sent ( %v ) - diff %v ms", rcvd, sent, rcvd-sent)
		})
		nc.Flush()
	}

	if app.pub {
		app.Println("-pub provided. starting publish loop")
		ticker := time.NewTicker(5 * time.Second)

		go func() {
			for {
				select {
				case <-sigs:
					app.Println("exit signal received, breaking")

					return
				case <-ticker.C:
					b := make([]byte, 8)
					binary.LittleEndian.PutUint64(b, uint64(time.Now().UnixNano()/1000000))
					nc.Publish(subject, b)
				}
			}
		}()
	}

	<-sigs
	app.Println("exiting Exec")

	return err
}
