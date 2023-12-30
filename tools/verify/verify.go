package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/albenik/go-serial"
)

const BAUD_RATE int = 1000000

const READ_TIMEOUT time.Duration = 1000 * time.Millisecond
const READ_TIMEOUT_OP time.Duration = 5000 * time.Millisecond

// Serial commands
const CMD_ACK byte = 'A'
const CMD_ERROR byte = 'E'
const CMD_OK byte = 'O'
const CMD_READ_SECTOR byte = 'R'
const CMD_DATA_BEGIN byte = 'D'
const CMD_HANDSHAKE byte = 'H'
const CMD_INITIALIZE byte = 'I'

const TRACKS byte = 80
const HEADS byte = 2
const SECTORS byte = 18
const SECTOR_SIZE int = 512

// Serial communication
func write_byte(port serial.Port, data byte) error {
	buf := []byte{data}

	n := 0
	var err error

	for n == 0 {
		n, err = port.Write(buf)

		if err != nil {
			return err
		}
	}

	return nil
}

func read_byte(port serial.Port, timeout time.Duration) (byte, error) {
	// Read buffer of 1 byte
	buf := make([]byte, 1)

	n := 0
	// var err error

	start := time.Now()

	for n <= 0 {
		port.SetReadTimeout(10)
		n, _ = port.Read(buf)

		// if err != nil {
		// 	port.SetReadTimeout(-1)
		// 	return 0, err
		// }

		if time.Since(start) > timeout {
			port.SetReadTimeout(-1)
			return 0, errors.New("timeout error")
		}
	}

	port.SetReadTimeout(-1)
	return buf[0], nil
}

func read_bytes(port serial.Port, n_bytes int, timeout time.Duration) ([]byte, error) {
	// Read buffer of n bytes
	buf := make([]byte, n_bytes)

	n := 0
	// var err error
	var read int

	start := time.Now()

	for n < n_bytes {
		port.SetReadTimeout(10)
		read, _ = port.Read(buf[n:])

		// if err != nil {
		// 	port.SetReadTimeout(-1)
		// 	return []byte{}, err
		// }

		if time.Since(start) > timeout {
			port.SetReadTimeout(-1)
			return []byte{}, errors.New("timeout error")
		}

		n += read
	}

	port.SetReadTimeout(-1)
	return buf, nil
}

func do_handshake(port serial.Port) error {

	// Send handshake
	write_byte(port, CMD_HANDSHAKE)

	// Expect handshake back
	res, err := read_byte(port, READ_TIMEOUT)

	if err != nil {
		// fmt.Printf("Read error: %s\n", err)
		return err
	}

	if res != CMD_HANDSHAKE {
		// fmt.Printf("Invalid handshake response: %c\n", res)
		return errors.New("invalid handshake response")
	}

	return nil
}

func connect_and_handshake(name string) (serial.Port, error) {

	mode := &serial.Mode{
		BaudRate: BAUD_RATE,
	}

	// Try to open port
	port, err := serial.Open(name, mode)

	if err != nil {
		return nil, err
	}

	// Wait for Arduino to reset
	time.Sleep(2000 * time.Millisecond)

	// Clear any stuff still in input buffer
	port.ResetInputBuffer()

	// Perform handshake on port
	err = do_handshake(port)

	if err != nil {
		// fmt.Printf("Handshake failed: %s\n", err)
		port.Close()
		return nil, err
	}

	return port, nil
}

func find_arduino() (serial.Port, string, error) {

	// Get serial ports list
	port_names, err := serial.GetPortsList()

	if err != nil {
		return nil, "", err
	}

	// If there are no ports available
	if len(port_names) == 0 {
		return nil, "", errors.New("no serial ports available")
	}

	// Repeat for each available port
	for _, name := range port_names {

		// Try to handhsake with device at this port
		port, err := connect_and_handshake(name)

		// If handshake succesful, use this port
		if err == nil {
			return port, name, nil
		}

		// fmt.Println(err)
	}
	return nil, "", errors.New("unable to find Arduino")

}

func do_initialize(port serial.Port) error {

	// Send initialization command
	write_byte(port, CMD_INITIALIZE)

	var res byte
	var err error

	// Wait for acknowledgement
	res, err = read_byte(port, READ_TIMEOUT)

	if err != nil {
		return err
	}

	if res != CMD_ACK {
		return errors.New("init no ACK")
	}

	// Wait for result
	res, err = read_byte(port, READ_TIMEOUT_OP)

	if err != nil {
		return err
	}

	if res != CMD_OK {
		return errors.New("initialization error")
	}

	return nil
}

func do_read_sector(port serial.Port, cylinder byte, head byte, sector byte) ([]byte, error) {

	var res byte
	var err error

	// Send read sector command
	write_byte(port, CMD_READ_SECTOR)
	write_byte(port, cylinder)
	write_byte(port, head)
	write_byte(port, sector)

	// Expect ACK
	res, err = read_byte(port, READ_TIMEOUT)

	if err != nil {
		return []byte{}, err
	}

	if res != CMD_ACK {
		return []byte{}, errors.New("no ACK")
	}

	// Read first result
	res, err = read_byte(port, READ_TIMEOUT_OP)

	if err != nil {
		return []byte{}, err
	}

	if res != CMD_DATA_BEGIN {
		return []byte{}, errors.New("floppy read error")
	}

	// If arduino tells us to expect data, read data
	var buf []byte
	buf, err = read_bytes(port, SECTOR_SIZE+3, READ_TIMEOUT)

	if err != nil {
		return []byte{}, err
	}

	// Read second result
	res, err = read_byte(port, READ_TIMEOUT)

	if err != nil {
		return []byte{}, err
	}

	if res != CMD_OK {
		return []byte{}, errors.New("floppy read error")
	}

	return buf[1 : SECTOR_SIZE+1], nil
}

func print_table_header() {

	sectorspace := int(SECTORS) * 3

	fmt.Println()
	fmt.Print("   ")
	for head := byte(0); head < HEADS; head++ {
		for space := 0; space < (sectorspace-6)/2; space++ {
			fmt.Print(" ")
		}
		PrtCol(fmt.Sprintf("HEAD %d", head), ColorWhiteBold)
		for space := 0; space < (sectorspace-6)/2; space++ {
			fmt.Print(" ")
		}
	}
	fmt.Println()
	fmt.Print("   ")
	for head := byte(0); head < HEADS; head++ {
		for sector := byte(1); sector <= SECTORS; sector++ {
			fmt.Printf("%3d", sector)
		}
	}
	fmt.Println()
}

func verify_sector_retries(port serial.Port, cylinder byte, head byte, sector byte, retries uint) (uint, error) {

	var err error

	tries := uint(0)

	for tries <= retries {
		_, err = do_read_sector(port, cylinder, head, sector)

		if err == nil {
			return tries, nil
		}

		tries++
	}

	return tries, err
}

func do_verify(port serial.Port, start_track OptionalByte, end_track OptionalByte, max_retries OptionalUint) (uint, uint, uint) {

	var track byte
	var head byte
	var sector byte

	good := uint(0)
	bad := uint(0)
	degraded := uint(0)

	if !start_track.has_value {
		start_track.value = 0
	}
	if !end_track.has_value {
		end_track.value = TRACKS - 1
	}

	print_table_header()

	for track = start_track.value; track <= end_track.value; track++ {

		fmt.Printf("%-2d ", track)

		for head = 0; head < HEADS; head++ {
			for sector = 1; sector <= SECTORS; sector++ {

				tries, err := verify_sector_retries(port, track, head, sector, max_retries.value)

				if err != nil {
					PrtCol(" E ", ColorBgRed)
					bad++
				} else {
					if tries == 0 {
						PrtCol(" S ", ColorBgGreen)
						good++
					} else {
						PrtCol(fmt.Sprintf("%3d", tries), ColorBgYellow)
						degraded++
					}
				}

			}
		}

		fmt.Println()
	}

	return good, bad, degraded
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

	// CTRL-C handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig != nil {
				port.Close()
				fmt.Println("Exiting...")
				os.Exit(0)
			}
		}
	}()

	// Initialize drive
	err = do_initialize(port)

	if err != nil {
		PrtCol("Error: ", ColorRedHI)
		fmt.Println("drive initalization failed!")
		port.Close()
		os.Exit(2)
	}

	// Do disk verification
	fmt.Println("Veifying disk...")
	good, bad, degraded := do_verify(port, conf.start_track, conf.end_track, conf.max_retries)

	PrtCol("Done!\n", ColorGreenHI)
	fmt.Printf("%d sectors ", good)
	PrtCol("good", ColorGreenHI)
	fmt.Printf(", %d sectors ", bad)
	PrtCol("bad", ColorRedHI)

	if conf.max_retries.value > 0 {
		fmt.Printf(", %d sectors ", degraded)
		PrtCol("degraded", ColorYellowHI)
	}

	fmt.Println()

	port.Close()
}
