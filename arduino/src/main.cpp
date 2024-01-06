#include "Arduino.h"

// Floppy controller
#include "floppy.h"

// Serial Interface
#include "serial.h"

#include "debug.h"

byte buf[SECTOR_DATA_SIZE * MAX_READ_BLOCKS_AMOUNT];

Floppy floppy;
SerialInterface serial_interface(&floppy, buf);

void setup()
{
    // Init debug
    debug_init();

    // Init serial
    Serial.begin(BAUD_RATE);

    // Setup floppy
    floppy.setup();

    debug_print("READY!\n");

    floppy.write_data();
}

void loop()
{
    serial_interface.tick();
    floppy.auto_motor_off();
}
