package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/albenik/go-serial"
)

// This device is an example implementation of an in-memory block device

type DeviceExample struct {
	dataset []byte
}

func (d *DeviceExample) ReadAt(p []byte, off uint) error {
	copy(p, d.dataset[off:int(off)+len(p)])
	log.Printf("[DeviceExample] READ offset:%d len:%d\n", off, len(p))
	return nil
}

func (d *DeviceExample) WriteAt(p []byte, off uint) error {
	copy(d.dataset[off:], p)
	log.Printf("[DeviceExample] WRITE offset:%d len:%d\n", off, len(p))
	return nil
}

func (d *DeviceExample) Disconnect() {
	log.Println("[DeviceExample] DISCONNECT")
}

func (d *DeviceExample) Flush() error {
	log.Println("[DeviceExample] FLUSH")
	return nil
}

func (d *DeviceExample) Trim(off, length uint) error {
	log.Printf("[DeviceExample] TRIM offset:%d len:%d\n", off, length)
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s /dev/nbd0\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	// Parse arguments
	conf, conf_res := parse_args()

	// Check result of configuration
	switch conf_res {
	case ConfigERR:
		os.Exit(1)
	case ConfigExitCleanly:
		os.Exit(0)
	}

	var err error
	var port serial.Port
	var name string

	// Find serial port
	if !conf.device.has_value {
		fmt.Println("Trying to find Arduino...")
		port, name, err = find_arduino()
		if err != nil {
			PrtCol("Error: ", ColorRedHI)
			fmt.Println("unable to find Arduino")
			os.Exit(1)
		}
	} else {
		fmt.Println("Connecting to Arduino...")
		port, err = connect_and_handshake(conf.device.value)
		if err != nil {
			PrtCol("Error: ", ColorRedHI)
			fmt.Printf("unable to connect to Arduino on port %s: %s\n", conf.device.value, err)
			os.Exit(1)
		}
		name = conf.device.value
	}

	PrtCol("Connected ", ColorGreenHI)
	fmt.Printf("on port %s\n", name)

	size := uint(1024 * 1024 * 512) // 512M
	deviceExp := &DeviceExample{}
	deviceExp.dataset = make([]byte, size)
	device, err := CreateDevice(args[0], size, deviceExp)
	if err != nil {
		fmt.Printf("Cannot create device: %s\n", err)
		os.Exit(1)
	}
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	go func() {
		if err := device.Connect(); err != nil {
			log.Printf("Buse device stopped with error: %s", err)
		} else {
			log.Println("Buse device stopped gracefully.")
		}
	}()
	<-sig
	// Received SIGTERM, cleanup
	fmt.Println("SIGINT, disconnecting...")
	device.Disconnect()
}
