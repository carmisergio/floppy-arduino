#pragma once

#define DEBUG_TX_PIN 4
#define DEBUG_RX_PIN 3
#define DEBUG_BAUD 115200

#include "SoftwareSerial.h"

// Comment this line to disable debug
// #define ENABLE_DEBUG

#ifdef ENABLE_DEBUG

extern SoftwareSerial debugSerial;

// If debug is enabled
#define debug_init() (debugSerial.begin(DEBUG_BAUD))
#define debug_print(msg) (debugSerial.print((msg)))

#else

#define debug_init()
#define debug_print(msg)

#endif