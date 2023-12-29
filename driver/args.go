package main

import (
	"fmt"
	"os"
)

// Arguments
const ARG_DEVICE string = "--device"
const ARG_DEVICE_SHORT string = "-d"
const ARG_RETRIES string = "--retries"
const ARG_RETRIES_SHORT string = "-r"
const ARG_HELP string = "--help"
const ARG_HELP_SHORT string = "-h"

// Messages
const MSG_HELP string = "Usage: driver [OPTIONS] NBD_DEVICE\nOptions: \n \t-d --device: Serial port of Arduino\n \t-r --retires: Number of read retries\n \t-h --help: Display this message"
const MSG_NBD_DEVICE_MISSING string = "driver: missing nbd device"
const MSG_OPT_VALUE_MISSING string = "driver: missing option value"
const MSG_OPT_VALUE_INVALID string = "driver: invalid option value"
const MSG_BAD_OPTION string = "driver: bad option"
const MSG_TRY_HELP string = "Try 'driver --help' for more information"

// Deafaults
const DEFAULT_MAX_RETRIES uint = 5

type OptionalString struct {
	value     string
	has_value bool
}

// type OptionalByte struct {
// 	value     byte
// 	has_value bool
// }

type OptionalUint struct {
	value     uint
	has_value bool
}

type Config struct {
	device      OptionalString
	max_retries OptionalUint
	nbd_device  OptionalString
}

type ConfigResult byte

const (
	ConfigOK          = 0
	ConfigERR         = 1
	ConfigExitCleanly = 2
)

func parse_args() (Config, ConfigResult) {

	var conf Config

	// Remove this program name form args
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {

		if args[i] == ARG_HELP || args[i] == ARG_HELP_SHORT {
			// --help or -h
			fmt.Println(MSG_HELP)
			return conf, ConfigExitCleanly

		} else if args[i] == ARG_DEVICE || args[i] == ARG_DEVICE_SHORT {
			// --device or -d

			// Try to consume option
			// Check if there is a value to consume
			i++
			if i < len(args) {
				conf.device.value = args[i]
				conf.device.has_value = true
			} else {
				fmt.Println(MSG_OPT_VALUE_MISSING)
				fmt.Println(MSG_TRY_HELP)
				return conf, ConfigERR
			}

		} else {
			conf.nbd_device.value = args[i]
			conf.nbd_device.has_value = true
		}
	}

	// Handle defaults
	if !conf.max_retries.has_value {
		conf.max_retries.value = DEFAULT_MAX_RETRIES
		conf.max_retries.has_value = true
	}

	// Check required parameters
	if !conf.nbd_device.has_value {
		fmt.Println(MSG_NBD_DEVICE_MISSING)
		fmt.Println(MSG_TRY_HELP)
		return conf, ConfigERR
	}

	return conf, ConfigOK
}
