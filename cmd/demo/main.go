package main

import (
	"context"

	"github.com/Sn0wo2/caelum"
)

func main() {
	log, err := caelum.New(caelum.Config{
		Filter: "debug",
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := log.Close(); err != nil {
			panic(err)
		}
	}()

	if err := log.Info("Hello, caelum!"); err != nil {
		panic(err)
	}

	span, err := log.CreateSpanContext(context.Background(), caelum.Event{
		Name: new("login"), Fields: map[string]any{"user": "guest"},
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := span.Close(); err != nil {
			panic(err)
		}
	}()

	if err := span.Record(map[string]any{"user": "Me0wo"}); err != nil {
		panic(err)
	}

	if err := log.DebugContext(span.Context(context.Background()), "User logged in"); err != nil {
		panic(err)
	}

	if err := log.Warn("disk almost full"); err != nil {
		panic(err)
	}

	if err := log.Error("request failed"); err != nil {
		panic(err)
	}

	if err := log.SetFilter("warn"); err != nil {
		panic(err)
	}

	if err := log.Info("this is filtered out"); err != nil {
		panic(err)
	}

	if err := log.Warn("only warnings and above now"); err != nil {
		panic(err)
	}
}
