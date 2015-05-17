package main

import (
	"github.com/fsouza/go-dockerclient"
)

type Formatter interface {
	Fmt(*docker.Stats, *docker.Container, string)
}

type FormatterFactory func() Formatter
