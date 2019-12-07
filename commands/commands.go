package commands

import (
	"github.com/liclac/eyesemoji/glowglasses"
)

type CommandFn func(gg *glowglasses.GlowGlassesX, args []string) error

type Command struct {
	Name string
	Help string
	Fn   CommandFn
}

var Commands = map[string]*Command{}

func Register(fn CommandFn, name, help string) {
	Commands[name] = &Command{Name: name, Help: help, Fn: fn}
}
