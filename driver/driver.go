package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/albenik/go-serial"
)

type DeviceExample struct {
	port    serial.Port
	dataset []byte
}

func (d *DeviceExample) ReadAt(p []byte, off uint) error {

	log.Printf("[DeviceExample] READ offset:%d len:%d\n", off, len(p))

	blocks_to_read := uint(len(p) / int(SECTOR_SIZE))
	start_block := off / SECTOR_SIZE

	data, err := read_blocks(d.port, start_block, start_block+blocks_to_read-1, 5, false)

	if err != nil {
		return errors.New("read error")
	}

	// Copy data
	copy(p, data)

	return nil
}

func (d *DeviceExample) WriteAt(p []byte, off uint) error {
	// copy(d.dataset[off:], p)
	// log.Printf("[DeviceExample] WRITE offset:%d len:%d\n", off, len(p))
	// return nil
	return errors.New("write not supported")
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

func main() {
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

	// Connection to arduino
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

	// Initialize drive
	err = do_initialize(port)

	if err != nil {
		PrtCol("Error: ", ColorRedHI)
		fmt.Println("drive initalization failed!")
		port.Close()
		os.Exit(2)
	}

	size := uint(512 * 2880) // 512M
	deviceExp := &DeviceExample{}
	deviceExp.port = port
	device, err := CreateDevice(conf.nbd_device.value, size, deviceExp)
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
	port.Close()
}
