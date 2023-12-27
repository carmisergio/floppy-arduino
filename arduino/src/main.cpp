#include "Arduino.h"

// Floppy controller
#include "floppy.h"

// Serial Interface
#include "serial.h"

byte buf[515];

Floppy floppy;
SerialInterface serial_interface(&floppy, buf);

void setup()
{
    // Init serial
    Serial.begin(BAUD_RATE);

    // Setup floppy
    floppy.setup();
}

void loop()
{
    serial_interface.tick();
    floppy.auto_motor_off();
}
