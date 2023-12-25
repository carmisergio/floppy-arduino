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
    OK,
};

// Floppy controller class
class Floppy
{
    // Floppy state
    unsigned short cur_track;

    // Set floppy drive selection
    static void select_drive(bool selected);

    // Set direction
    static void set_direction(bool direction);

    // Set head
    static void set_head(unsigned short head);

    // Step n steps
    static void step(unsigned short n);

    // Find tarck 0
    static bool go_to_track_0();

public:
    bool initialized;

    static byte read_data(byte *buffer, unsigned int n);

    // Constructor
    Floppy();

    // Initialize floppy
    FloppyError initialize();

    // Seek to track
    FloppyError seek(unsigned short track);

    // Read sector
    FloppyError read_sector(byte *buffer, unsigned short cylinder, unsigned short head, unsigned short sector);
};
