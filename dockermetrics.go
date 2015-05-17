package main

import (
	"flag"
	"log"
	"os"
)

type Sink interface{}

var (
	flagset = flag.NewFlagSet("dockermetrics", flag.ExitOnError)
	flags   = struct {
		Debug          bool
		SyncInterval   int
		DockerEndpoint string
		Formatter      string
	}{}
	formatters map[string]FormatterFactory
)

func init() {
	flagset.BoolVar(&flags.Debug, "debug", false, "Moar logging!")
	flagset.IntVar(&flags.SyncInterval, "sync-interval", 30, "Time between syncing container list")
	flagset.StringVar(&flags.DockerEndpoint, "docker-endpoint", "unix:///run/docker.sock", "Docker endpoint")
	flagset.StringVar(&flags.Formatter, "formatter", "graphite", "Metric formatter")

	formatters = map[string]FormatterFactory{
		"graphite": NewGraphite,
	}
}

func main() {
	log.Println("Starting docker metrics")
	flagset.Parse(os.Args[1:])

	var formatter Formatter

	f := formatters[flags.Formatter]

	if f == nil {
		log.Fatal("Unknown formatter")
	}
	formatter = f()

	client := NewClient(formatter)
	client.Connect()
	client.Sync()
	client.SyncTick()
	client.Watch()
	log.Println("Exiting")
}
