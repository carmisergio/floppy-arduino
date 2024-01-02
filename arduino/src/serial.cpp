/**
 * Serial Interface
 *
 * @author Sergio Carmine <me@sergiocarmi.net>
 * @date 26/12/2023
 */

#include "serial.h"

#define CMD_ACK 'A'
#define CMD_READ_SECTOR 'R'
#define CMD_READ_BLOCKS 'B'
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

uint16_t read_uint16_t()
{
    uint16_t res;

    // Read data
    Serial.readBytes((byte *)&res, sizeof(uint16_t));

    return res;
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
        case CMD_READ_BLOCKS:
            cmd_read_blocks();
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
        // Fill buffer with given data
        // for (int i = 0; i < SECTOR_SIZE; i++)
        // {
        //     buf[i + 1] = 'D';
        // }
        // Send back data
        Serial.write(CMD_OK);
        Serial.write(buf + 1, SECTOR_SIZE);
    }

    Serial.flush();
}

void SerialInterface::cmd_read_blocks()
{
    uint16_t block;
    byte amount;

    // Read command info
    block = read_uint16_t();
    amount = read_byte();

    // block = 35;
    // amount = 2;

    // Tell host that we're doing the read
    Serial.write(CMD_ACK);
    Serial.flush();

    // Perform read
    FloppyError ec = floppy->read_blocks(buf, block, amount);

    // If there was an error
    if (ec != FloppyError::OK)
    {
        Serial.write(CMD_ERROR);
    }
    else
    {
        // Fill buffer with given data
        // for (int i = 0; i < SECTOR_SIZE; i++)
        // {
        //     buf[i + 1] = 'D';
        // }
        // Send back data
        Serial.write(CMD_OK);

        for (byte i = 0; i < amount; i++)
        {
            Serial.write(buf + SECTOR_DATA_SIZE * i + 1, SECTOR_SIZE);
        }
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