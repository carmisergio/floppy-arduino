package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/albenik/go-serial"
)

const BAUD_RATE int = 115200

const READ_TIMEOUT time.Duration = 1000 * time.Millisecond
const READ_TIMEOUT_OP time.Duration = 5000 * time.Millisecond

const READ_RETRIES byte = 5

// Serial commands
const CMD_ACK byte = 'A'
const CMD_ERROR byte = 'E'
const CMD_OK byte = 'O'
const CMD_READ_SECTOR byte = 'R'
const CMD_HANDSHAKE byte = 'H'
const CMD_INITIALIZE byte = 'I'

const TRACKS byte = 80
const HEADS byte = 2
const SECTORS byte = 18
const SECTOR_SIZE uint = 512

const N_BLOCKS uint = uint(TRACKS) * uint(HEADS) * uint(SECTORS)

// const N_BLOCKS uint = 36

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

func read_bytes(port serial.Port, n_bytes uint, timeout time.Duration) ([]byte, error) {
	// Read buffer of n bytes
	buf := make([]byte, n_bytes)

	n := uint(0)
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

		n += uint(read)
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
	time.Sleep(500 * time.Millisecond)

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

func find_arduino() (serial.Port, error) {

	// Get serial ports list
	port_names, err := serial.GetPortsList()

	if err != nil {
		return nil, err
	}

	// If there are no ports available
	if len(port_names) == 0 {
		return nil, errors.New("no serial ports available")
	}

	// Repeat for each available port
	for _, name := range port_names {

		// Try to handhsake with device at this port
		port, err := connect_and_handshake(name)

		// If handshake succesful, use this port
		if err == nil {
			return port, nil
		}

		// fmt.Println(err)
	}

	return nil, errors.New("unable to find Arduino")
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

	// Read result
	res, err = read_byte(port, READ_TIMEOUT_OP)

	if err != nil {
		return []byte{}, err
	}

	if res != CMD_OK {
		return []byte{}, errors.New("floppy read error")
	}

	// If result is OK, read data
	var buf []byte
	buf, err = read_bytes(port, SECTOR_SIZE, READ_TIMEOUT)

	if err != nil {
		return []byte{}, err
	}

	return buf, nil
}

func LBAtoCHS(address uint) (byte, byte, byte) {
	var cylinder, head, sector byte

	// Calulcate results
	cylinder = byte(address / (uint(SECTORS) * uint(HEADS)))
	head = byte((address / uint(SECTORS)) % uint(HEADS))
	sector = byte(address%uint(SECTORS)) + 1 // First sector is 1

	return cylinder, head, sector
}

// Read sector with logical block addressing
func read_block(port serial.Port, address uint, retries byte) ([]byte, error) {

	var data []byte
	var err error

	// Convert LBA to CHS addressing
	cylinder, head, sector := LBAtoCHS(address)

	// Always do at least 1 try
	retries++

	// fmt.Printf("Reading C=%d, H=%d, S=%d\n", cylinder, head, sector)

	for retries > 0 {
		data, err = do_read_sector(port, cylinder, head, sector)

		if err == nil {
			return data, nil
		}

		retries--
	}

	return data, err
}

func read_all_blocks(port serial.Port) ([]byte, error) {
	// Initialize buffer
	blocks := make([]byte, N_BLOCKS*SECTOR_SIZE)

	for i := uint(0); i < N_BLOCKS; i++ {

		fmt.Printf("Block %d/%d\n", i, N_BLOCKS-1)

		block, err := read_block(port, i, READ_RETRIES)

		if err != nil {
			return []byte{}, err
		}

		// Copy read block in file buffer
		start := i * SECTOR_SIZE
		end := start + SECTOR_SIZE
		copy(blocks[start:end], block)
	}

	return blocks, nil
}

func main() {

	var err error
	var port serial.Port

	// Find serial port
	port, err = find_arduino()
	if err != nil {
		fmt.Println("Unable to find Arduino!")
		os.Exit(1)
	}
	fmt.Println("Connected!")

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
	fmt.Println("Initializing drive...")
	err = do_initialize(port)

	if err != nil {
		fmt.Println("Drive initalization failed!")
		port.Close()
		os.Exit(2)
	}

	fmt.Println("Drive initialized!")

	fmt.Println("Reading disk...")

	// Try reading a few sectors
	var data []byte
	data, err = read_all_blocks(port)

	if err != nil {
		fmt.Printf("Read error: %s", err)
	}

	port.Close()

	// Write data to disk
	outf, _ := os.Create("disk.img")
	outf.Write(data)
	outf.Close()
}
