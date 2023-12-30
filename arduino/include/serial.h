/**
 * Serial Interface
 *
 * @author Sergio Carmine <me@sergiocarmi.net>
 * @date 26/12/2023
 */

#pragma once

#include "floppy.h"

#define BAUD_RATE 1000000

class SerialInterface
{

    Floppy *floppy; // Pointer to the floppy class
    byte *buf;      // Pointer to buffer for reads and writes

    // Read sector command handler
    void cmd_read_sector();

    // Handshake command handler
    void cmd_handshake();

    // Drive initialization command
    void cmd_initialize();

public:
    // Constructor
    SerialInterface(Floppy *floppy, byte *buf);

    // Handle messages from Serial
    void tick();
};