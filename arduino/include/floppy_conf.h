/**
 * Floppy configuration
 */

#pragma once

///////////////////////// PINS
#define PIN_DRVSEL 12
#define PIN_DIR 7
#define PIN_STEP 6
#define PIN_TRACK0 5
#define PIN_HEADSEL 4
#define PIN_MOTOR 10
#define PIN_READDATA 8 // Input capture pin for TIMER 1

///////////////////////// GEOMETRY
#define TRACKS 80
#define SECTORS 18
#define SECTOR_SIZE 512

///////////////////////// TIMINGS
#define STEP_DELAY 10           // (ms)
#define STEP_PULSE 1            // (ms)
#define MOTOR_OFF_TIMEOUT 10000 // (ms)
