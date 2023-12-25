#include "Arduino.h"

// Floppy controller
#include "floppy.h"

Floppy floppy;

void print_hex(byte *buf, unsigned int n)
{
    for (byte *i = buf; i < buf + n; i++)
    {
        Serial.print(*i, HEX);
        Serial.print(" ");
    }
    Serial.println();
    Serial.flush();
}

void setup()
{
    Serial.begin(4000000);

    // Setup Floppy
    if (floppy.initialize() != FloppyError::OK)
        Serial.println("Floppy initialization error!");
}

byte buf[515];
void loop()
{

    if (floppy.initialized)
    {
        for (int i = 0; i < 79; i += 1)
        {
            floppy.seek(i);

            for (int k = 0; k <= 1; k++)
            {

                for (int j = 1; j <= 18; j++)
                {

                    Serial.print("## Track: ");
                    Serial.print(i);
                    Serial.print(", Head: ");
                    Serial.print(k);
                    Serial.print(", Sector: ");
                    Serial.println(j);

                    FloppyError fc = floppy.read_sector(buf, i, k, j);

                    switch (fc)
                    {
                    case OK:
                        // print_hex(buf, 515);
                        break;
                    case SEEK_ERROR:
                        Serial.println("Seek error!");
                        break;
                    case SECTOR_NOT_FOUND:
                        Serial.println("Sector not found!");
                        break;
                    default:
                        Serial.println("Other error!");
                    }
                }
            }
        }

        // delay(10000);
    }
}
