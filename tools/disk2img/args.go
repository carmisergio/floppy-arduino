package main

import (
	"fmt"
	"os"
	"strconv"
)

// Arguments
const ARG_DEVICE string = "--device"
const ARG_DEVICE_SHORT string = "-d"
const ARG_START_BLK string = "--start-block"
const ARG_START_BLK_SHORT string = "-s"
const ARG_END_BLK string = "--end-block"
const ARG_END_BLK_SHORT string = "-e"
const ARG_RETRIES string = "--retries"
const ARG_RETRIES_SHORT string = "-r"
const ARG_IGNORE_ERRORS string = "--ignore-errors"
const ARG_IGNORE_ERRORS_SHORT string = "-i"
const ARG_HELP string = "--help"
const ARG_HELP_SHORT string = "-h"

// Messages
const MSG_HELP string = "Usage: disk2img [OPTIONS] OUT_FILE\nOptions: \n \t-d --device: Serial port of Arduino\n \t-s --start-block: First block to read\n \t-e --end-block: Last block to read\n \t-r --retires: Number of read retries\n \t-i --ignore-errors: Ignore read errors\n \t-h --help: Display this message"
const MSG_OUT_FILE_MISSING string = "disk2img: missing out file"
const MSG_OPT_VALUE_MISSING string = "disk2img: missing option value"
const MSG_OPT_VALUE_INVALID string = "disk2img: invalid option value"
const MSG_BAD_OPTION string = "disk2img: bad option"
const MSG_TRY_HELP string = "Try 'disk2img --help' for more information"

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
	device        OptionalString
	start_block   OptionalUint
	end_block     OptionalUint
	max_retries   OptionalUint
	ignore_errors bool
	out_file      OptionalString
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

		} else if args[i] == ARG_START_BLK || args[i] == ARG_START_BLK_SHORT {
			// --start-block or -s

			// Try to consume option
			// Check if there is a value to consume
			i++
			if i < len(args) {

				value, err := strconv.ParseInt(args[i], 10, 32)

				if err == nil {
					conf.start_block.value = uint(value)
					conf.start_block.has_value = true
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
		} else if args[i] == ARG_END_BLK || args[i] == ARG_END_BLK_SHORT {
			// --end-block or -e

			// Try to consume option
			// Check if there is a value to consume
			i++
			if i < len(args) {

				value, err := strconv.ParseInt(args[i], 10, 32)

				if err == nil {
					conf.end_block.value = uint(value)
					conf.end_block.has_value = true
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
		} else if args[i] == ARG_IGNORE_ERRORS || args[i] == ARG_IGNORE_ERRORS_SHORT {
			// --ignore-errrors or -i

			conf.ignore_errors = true
		} else {
			conf.out_file.value = args[i]
			conf.out_file.has_value = true
		}
	}

	// Handle defaults
	if !conf.max_retries.has_value {
		conf.max_retries.value = DEFAULT_MAX_RETRIES
		conf.max_retries.has_value = true
	}

	// Check required parameters
	if !conf.out_file.has_value {
		fmt.Println(MSG_OUT_FILE_MISSING)
		fmt.Println(MSG_TRY_HELP)
		return conf, ConfigERR
	}

	return conf, ConfigOK
}
