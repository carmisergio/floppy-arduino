/**
 * Serial Interface
 *
 * @author Sergio Carmine <me@sergiocarmi.net>
 * @date 26/12/2023
 */

#include "serial.h"
#include "raw_serial.h"

#define CMD_ACK 'A'
#define CMD_READ_SECTOR 'R'
#define CMD_DATA_INCOMING 'D'
#define CMD_HANDSHAKE 'H'
#define CMD_INITIALIZE 'I'
#define CMD_ERROR 'E'
#define CMD_OK 'O'

byte read_byte()
{
    // Wait for data to be available
    while (Serial.available() <= 0)
        ;

    byte read = Serial.read();

    return read;
}

SerialInterface::SerialInterface(Floppy *floppy, byte *buf)
{
    this->floppy = floppy;
    this->buf = buf;
}

void SerialInterface::tick()
{
    if (Serial.available() > 0)
    {
        digitalWrite(13, LOW);
        switch (Serial.read())
        {
        case CMD_READ_SECTOR:
            cmd_read_sector();
            break;
        case CMD_HANDSHAKE:
            cmd_handshake();
            break;
        case CMD_INITIALIZE:
            cmd_initialize();
            break;
        }
    }
}

void SerialInterface::cmd_read_sector()
{
    byte cylinder, head, sector;

    // Read sector info
    cylinder = read_byte();
    head = read_byte();
    sector = read_byte();

    // cylinder = 0;
    // head = 0;
    // sector = 1;

    // Tell host that we're doing the read
    Serial.write(CMD_ACK);
    Serial.flush();

    // Perform read
    FloppyError ec = floppy->read_sector(buf, cylinder, head, sector);

    // If there was an error
    if (ec != FloppyError::OK)
    {

        Serial.write(CMD_ERROR);
    }
    else
    {
        // Everything OK
        Serial.write(CMD_OK);
    }

    Serial.flush();
}

void SerialInterface::cmd_initialize()
{

    Serial.write(CMD_ACK);

    // Initialize floppy
    if (floppy->initialize() != FloppyError::OK)
    {
        Serial.write(CMD_ERROR);
        return;
    }

    Serial.write(CMD_OK);
}

void SerialInterface::cmd_handshake()
{
    // Respond to handshake
    Serial.write(CMD_HANDSHAKE);
    Serial.flush();
}

void serial_send_byte_immediate(byte c)
{
    // Wait for data register to be empty
    while (!(UCSR0A & (1 << UDRE0)))
        ;

    // Write data
    UDR0 = c;
}