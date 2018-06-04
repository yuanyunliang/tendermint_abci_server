package main

import (
	"os"
	"fmt"
	"io"
	"bufio"
	"errors"
	"github.com/tendermint/tmlibs/log"
	"github.com/tendermint/abci/types"
	"github.com/tendermint/abci/server"
	//"github.com/tendermint/abci/example/dummy"
	"abci_server/example/dummy"
	cmn "github.com/tendermint/tmlibs/common"

)

// client is a global variable so it can be reused by the console
var (
	//client abcicli.Client
	logger log.Logger
)

// flags
var (
	// global
	flagAddress  = "tcp://0.0.0.0:46658" // Address of application socket
	flagAbci     = "socket"              // Either socket or grpc
	//flagVerbose  = false                 // for the println output
	flagLogLevel = "debug"               // for the logger
	flagPersist string
)

func main() {
	if err := initLogger(); err !=nil {
		fmt.Errorf("init loger failed, %s", err)
	}

	go func() {
		// err := runCounter()
		err := runDummy()
		if err != nil {
			fmt.Printf("run dummy error: %s\n", err)
			os.Exit(1)
		}
	}()

	runConsole()

}

func initLogger() error {
	if logger == nil {
		allowLevel, err := log.AllowLevel(flagLogLevel)
		if err != nil {
			return err
		}

		logFile, err := os.Create("logs/abci.log")
		if err != nil {
			fmt.Println("ABCI log file init error:", err)
		}
		multiWriter := io.MultiWriter(logFile, os.Stdout)

		// logger = log.NewFilter(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), allowLevel)
		logger = log.NewFilter(log.NewTMLogger(log.NewSyncWriter(multiWriter)), allowLevel)
	}
	return nil
}

func runDummy() error {

	// Create the application - in memory or persisted to disk
	var app types.Application
	if flagPersist == "" {
		app = dummy.NewDummyApplication()
	} else {
		app = dummy.NewPersistentDummyApplication(flagPersist)
		app.(*dummy.PersistentDummyApplication).SetLogger(logger.With("module", "dummy"))
	}

	// Start the listener
	srv, err := server.NewServer(flagAddress, flagAbci, app)
	if err != nil {
		return err
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		return err
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})
	return nil
}

func runConsole() error {
	for {
		fmt.Printf("> ")
		bufReader := bufio.NewReader(os.Stdin)
		line, more, err := bufReader.ReadLine()
		if more {
			return errors.New("input is too long")
		} else if err != nil {
			return err
		}

		fmt.Println("ABCI Server,", line)
	}
}
