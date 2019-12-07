package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/liclac/eyesemoji/commands"
	"github.com/liclac/eyesemoji/glowglasses"
	"github.com/mattn/go-shellwords"
	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	"github.com/muka/go-bluetooth/bluez/profile/gatt"
	"github.com/peterh/liner"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var fVerbose = pflag.BoolP("verbose", "v", false, "Spam less")
var fQuiet = pflag.BoolP("quiet", "q", false, "Spam more")

func Init() error {
	pflag.Parse()
	if *fVerbose {
		log.SetLevel(log.TraceLevel)
	} else if !*fQuiet {
		log.SetLevel(log.DebugLevel)
	}
	return nil
}

func FindGlasses(adp *adapter.Adapter1) (*device.Device1, error) {
	log.Debug("Enabling discovery...")
	if err := adp.StartDiscovery(); err != nil {
		return nil, errors.Wrap(err, "enabling discovery")
	}
	defer adp.StopDiscovery()

	for {
		devs, err := adp.GetDevices()
		if err != nil {
			return nil, errors.Wrap(err, "listing devices")
		}
		for _, dev := range devs {
			log.WithFields(log.Fields{
				"uuids": dev.Properties.UUIDs,
				"addr":  dev.Properties.Address,
			}).Debugf("Found: %s", dev.Properties.Name)
			for _, uuid := range dev.Properties.UUIDs {
				if uuid == "0000fff0-0000-1000-8000-00805f9b34fb" {
					return dev, nil
				}
			}
		}
		log.Info("Waiting for device to appear...")
		time.Sleep(100 * time.Millisecond)
	}
}

func FindCharacteristic(dev *device.Device1) (*gatt.GattCharacteristic1, error) {
	for {
		chars, err := dev.GetCharacteristics()
		if err != nil {
			return nil, errors.Wrap(err, "listing characteristics")
		}
		for _, ch := range chars {
			log.WithFields(log.Fields{
				"flags": ch.Properties.Flags,
				"value": ch.Properties.Value,
			}).Debugf("Characteristic: %s", ch.Properties.UUID)
			return ch, nil
		}
		log.Info("Waiting for GATT services to appear...")
		time.Sleep(100 * time.Millisecond)
	}
}

func RunREPL(gg *glowglasses.GlowGlassesX) error {
	st := liner.NewLiner()
	st.SetMultiLineMode(true)
	for {
		input, err := st.Prompt(gg.Device.Properties.Name + "> ")
		if err != nil {
			return err
		}
		if err := Eval(gg, input); err != nil {
			log.Error(err)
			continue
		}
	}
}

func Eval(gg *glowglasses.GlowGlassesX, input string) error {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return nil
	}
	words, err := shellwords.Parse(input)
	if err != nil {
		return err
	}
	cmd, ok := commands.Commands[words[0]]
	if !ok {
		return errors.Errorf("unrecognised command, try 'help'")
	}
	return cmd.Fn(gg, words[1:])
}

func Main() error {
	if err := Init(); err != nil {
		return err
	}

	adp, err := api.GetDefaultAdapter()
	if err != nil {
		return errors.Wrap(err, "couldn't get default adapter")
	}
	log.WithFields(log.Fields{
		"addr":      adp.Properties.Address,
		"addr_type": adp.Properties.AddressType,
		"name":      adp.Properties.Name,
		"alias":     adp.Properties.Alias,
	}).Info("Found Adapter!")

	log.Info("Finding glasses...")
	dev, err := FindGlasses(adp)
	if err != nil {
		return errors.Wrap(err, "couldn't find glasses")
	}

	log.WithFields(log.Fields{
		"addr": dev.Properties.Address,
	}).Infof("Connecting to: %s", dev.Properties.Name)
	if err := dev.Connect(); err != nil {
		return errors.Wrap(err, "connecting")
	}

	log.Info("Inspecting GATT services...")
	ch, err := FindCharacteristic(dev)
	if err := dev.Connect(); err != nil {
		return errors.Wrap(err, "listing gatt")
	}

	gg := glowglasses.New(dev, ch)
	return RunREPL(gg)
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
