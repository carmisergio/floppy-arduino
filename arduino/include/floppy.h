/**
 * Floppy controller library
 *
 * @author Sergio Carmine <me@sergiocarmi.net>
 * @date 21/12/2023
 */

#pragma once

#include "Arduino.h"

#include "floppy_conf.h"

// Floppy controller error codes
enum FloppyError
{
    TRACK_OUT_OF_RANGE,
    TRACK0_NOT_FOUND,
    NOT_INITIALIZED,
    SEEK_ERROR,
    SECTOR_NOT_FOUND,
    INCORRECT_DATA_MARK,
    NO_PULSE,
    CRC,
    OK,
};

// Floppy controller class
class Floppy
{
    // Floppy state
    unsigned short cur_track;
    bool motor_on;
    unsigned long last_op_time;

    // Set floppy drive selection
    static void select_drive(bool selected);

    // Set direction
    static void set_direction(bool direction);

    // Set head
    static void set_head(unsigned short head);

    // Set motor on
    void set_motor_state(bool state);

    // Step n steps
    static void step(unsigned short n);

    // Find tarck 0
    static bool go_to_track_0();

    // Save last operation time
    void save_last_op_time();

public:
    bool initialized;

    byte read_data(byte *buffer, unsigned int n);
    byte read_data_serialout(byte *buffer, unsigned int n);

    // Constructor
    Floppy();

    void setup();

    // Initialize floppy
    FloppyError initialize();

    // Seek to track
    FloppyError seek(byte track);

    // Find sector
    FloppyError find_sector(byte cylinder, byte head, byte sector);

    // Read sector
    FloppyError read_sector(byte *buffer, byte cylinder, byte head, byte sector);

    // Run automatic motor off routines
    void auto_motor_off();

    void test_serial();
};
