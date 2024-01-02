package main

import (
	"fmt"
	"os"
	"strconv"
)

// Arguments
const ARG_DEVICE string = "--device"
const ARG_DEVICE_SHORT string = "-d"
const ARG_START_TK string = "--start-track"
const ARG_START_TK_SHORT string = "-s"
const ARG_END_TK string = "--end-track"
const ARG_END_TK_SHORT string = "-e"
const ARG_RETRIES string = "--retries"
const ARG_RETRIES_SHORT string = "-r"
const ARG_HELP string = "--help"
const ARG_HELP_SHORT string = "-h"

// Messages
const MSG_HELP string = "Usage: verify [OPTIONS]\nOptions: \n \t-d --device: Serial port of Arduino\n \t-s --start-track: Track to start verification from\n \t-e --end-track: Track to end verification on\n \t-r --retires: Number of read retries\n \t-h --help: Display this message"
const MSG_OPT_VALUE_MISSING string = "verify: missing option value"
const MSG_OPT_VALUE_INVALID string = "verify: invalid option value"
const MSG_BAD_OPTION string = "verify: bad option"
const MSG_TRY_HELP string = "Try 'verify --help' for more information"

// Deafaults
const DEFAULT_MAX_RETRIES uint = 0

type OptionalString struct {
	value     string
	has_value bool
}

type OptionalByte struct {
	value     byte
	has_value bool
}

type OptionalUint struct {
	value     uint
	has_value bool
}

type Config struct {
	device      OptionalString
	start_track OptionalByte
	end_track   OptionalByte
	max_retries OptionalUint
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

		} else if args[i] == ARG_START_TK || args[i] == ARG_START_TK_SHORT {
			// --start-track or -s

			// Try to consume option
			// Check if there is a value to consume
			i++
			if i < len(args) {

				value, err := strconv.ParseInt(args[i], 10, 8)

				if err == nil {
					conf.start_track.value = byte(value)
					conf.start_track.has_value = true
				} else {
					fmt.Println(MSG_OPT_VALUE_INVALID)
					fmt.Println(MSG_TRY_HELP)
					return conf, ConfigERR
				}

			} else {
				fmt.Println(MSG_OPT_VALUE_MISSING)
				fmt.Println(MSG_TRY_HELP)
				return conf, ConfigERR
			}
		} else if args[i] == ARG_END_TK || args[i] == ARG_END_TK_SHORT {
			// --end-track or -e

			// Try to consume option
			// Check if there is a value to consume
			i++
			if i < len(args) {

				value, err := strconv.ParseInt(args[i], 10, 8)

				if err == nil {
					conf.end_track.value = byte(value)
					conf.end_track.has_value = true
				} else {
					fmt.Println(MSG_OPT_VALUE_INVALID)
					fmt.Println(MSG_TRY_HELP)
					return conf, ConfigERR
				}

			} else {
				fmt.Println(MSG_OPT_VALUE_MISSING)
				fmt.Println(MSG_TRY_HELP)
				return conf, ConfigERR
			}
		} else if args[i] == ARG_RETRIES || args[i] == ARG_RETRIES_SHORT {
			// --retries or -r

			// Try to consume option
			// Check if there is a value to consume
			i++
			if i < len(args) {

				value, err := strconv.ParseUint(args[i], 10, 32)

				if err == nil {
					conf.max_retries.value = uint(value)
					conf.max_retries.has_value = true
				} else {
					fmt.Println(MSG_OPT_VALUE_INVALID)
					fmt.Println(MSG_TRY_HELP)
					return conf, ConfigERR
				}

			} else {
				fmt.Println(MSG_OPT_VALUE_MISSING)
				fmt.Println(MSG_TRY_HELP)
				return conf, ConfigERR
			}
		} else {
			fmt.Println(MSG_BAD_OPTION)
			fmt.Println(MSG_TRY_HELP)
			return conf, ConfigERR
		}
	}

	// Handle defaults
	if !conf.max_retries.has_value {
		conf.max_retries.value = DEFAULT_MAX_RETRIES
		conf.max_retries.has_value = true
	}

	return conf, ConfigOK
}
