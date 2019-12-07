package commands

import (
	"fmt"

	"github.com/liclac/eyesemoji/glowglasses"
)

func init() {
	Register(On, "on", "Turn the glasses on")
	Register(Off, "off", "Turn the glasses off")
	Register(Help, "help", "Show this message")
}

func On(gg *glowglasses.GlowGlassesX, args []string) error {
	return gg.On()
}

func Off(gg *glowglasses.GlowGlassesX, args []string) error {
	return gg.Off()
}

func Help(gg *glowglasses.GlowGlassesX, args []string) error {
	for _, cmd := range Commands {
		fmt.Printf("  %-7s %s\n", cmd.Name, cmd.Help)
	}
	return nil
}
