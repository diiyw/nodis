//go:build !wasm

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/storage"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var CLI struct {
	Addr    string `arg:"" default:":6380" usage:"nodis server address"`
	Storage string `arg:"" default:"memory" usage:"select storage: memory, pebble"`
}

func main() {
	_ = kong.Parse(&CLI)
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	opts := *nodis.DefaultOptions
	switch CLI.Storage {
	case "pebble":
		opts.Storage = storage.NewPebble("data", nil)
	case "memory":
	default:
		return fmt.Errorf("unknown storage: %s", CLI.Storage)
	}
	n, err := nodis.OpenWithError(&opts)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	servingDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = n.Close()
		case <-servingDone:
		}
	}()
	err = n.Serve(CLI.Addr)
	close(servingDone)
	closeErr := n.Close()
	if closeErr != nil {
		return closeErr
	}
	if errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
