/**
 * Floppy controller library
 *
 * @author Sergio Carmine <me@sergiocarmi.net>
 * @date 21/12/2023
 */

#include "floppy.h"

#include "crc.h"
#include "raw_serial.h"

asm volatile("   .equ TIFR1,    0x16\n"  // timer 1 flag register
             "   .equ TIFR2,     0x17\n" // timer 2 flag register
             "   .equ TOV,     0\n"      // overflow flag
             "   .equ OCF,     1\n"      // output compare flag
             "   .equ ICF,     5\n"      // input capture flag
             "   .equ TCCRC1,   0x82\n"  // timer 1 control register C
             "   .equ FOC,     0x80\n"   // force output compare flag
             "   .equ TCNTL1,   0x84\n"  // timer 1 counter (low byte)
             "   .equ ICRL1,    0x86\n"  // timer 1 input capture register (low byte)
             "   .equ OCRL1,    0x88\n"  // timer 1 output compare register (low byte)
             "   .equ TCNT2,    0xB2\n"  // timer 2 current count register
             "   .equ UDR0,   0xC6\n\t"  // USART data register
             "   .equ UCSR0A, 0xC0\n\t"  // USART control register A
             "   .equ UDRE0,  5\n\t"     // USART ready to send bit
);

// ASM Macros
asm volatile(
    //
    // Measure length of a data pulse
    //
    // Inputs:
    //  - r17: last pulse time
    // Outputs:
    //  - r17: this pulse time
    //  - r18: pulse length
    // Clobbers:
    //  - r0
    ".macro MSPULSE\n\t"
    "   0:          sbic    TIFR1, ICF\n\t" // (1/2) Check if there has been an input capture
    "               rjmp    1f\n\t"         // (2) Loop if not
    "               sbic    TIFR2, TOV\n\t" // Check timeout
    "               rjmp    no_pulse\n\t"   // There's no pulse
    "               rjmp    0b\n\t"
    "   1:          lds     r0, ICRL1\n\t"  // (2) Get time of input capture
    "               sbi     TIFR1, ICF\n\t" // (2) Clear input capture happened flag
    "               mov     r18, r0\n\t"    // (1)
    "               sub     r18, r17\n\t"   // (1) Compute pulse length (r17 contains last pulse time)
    "               mov     r17, r0\n\t"    // (1) r17 is now new pulse time
    "               sts     TCNT2, r19\n\t" // Clear timeout
    ".endm\n\t"

    //
    // Jump to label if pulse is not of length
    //
    // Arguments:
    //  - length: pulse to recognize
    //             -> S: Short
    //             -> M: Medium
    //             -> L: Long
    //  - dst: label to jump to if different
    // Inputs:
    //  - r18: pulse length
    //  - r20: minimum medium pulse length
    //  - r21: minimum long pulse length
    ".macro JPDIFF length:req, dst:req\n\t"
    "   .if \\length == S\n"
    "               cp      r18, r20\n\t" // If >= minimum medium pulse, jump
    "               brlo .+2\n\t"
    "               rjmp \\dst\n\t"
    "   .elseif \\length == M\n"
    "               cp      r18, r20\n\t" // If < minimum medium pulse, jump
    "               brge .+2\n\t"
    "               rjmp \\dst\n\t"
    "               cp      r18, r21\n\t" // If >= minimum long pulse, jump
    "               brlo .+2\n\t"
    "               rjmp \\dst\n\t"
    "   .elseif \\length == L\n"
    "               cp      r18, r21\n\t" // If >= minimum medium pulse, jump
    "               brge .+2\n\t"
    "               rjmp \\dst\n\t"
    "   .endif\n\t"
    ".endm\n\t"

    //
    // Jump to different label depending on pulse length
    //
    // Arguments:
    //  - dstS: label to jump to if pulse is Short
    //  - dstM: label to jump to if pulse is Medium
    //  - dstL: label to jump to if pulse is Long
    // Inputs:
    //  - r18: pulse length
    //  - r20: minimum medium pulse length
    //  - r21: minimum long pulse length
    ".macro PSWITCH dstS:req, dstM:req, dstL:req\n\t"
    "               cp      r18, r20\n\t" // If < minimum short pulse
    "               brge    .+2\n\t"
    "               rjmp    \\dstS\n\t"
    "               cp      r18, r21\n\t" // If < minimum long pulse
    "               brge    .+2\n\t"
    "               rjmp    \\dstM\n\t"
    "               rjmp    \\dstL\n\t"
    ".endm\n\t"

    //
    // Send to UART immediately
    // If we can not send it immediately, fail the read
    //
    // Arguments:
    //  - src: register where to get the byte to send
    // Clobbers:
    //  - r0
    ".macro SENDBYTE src:req fail:req\n\t"
    // "                   nop\n\t"
    // "               lds     r0, UCSR0A\n\t" // (2)
    // "               sbrs    r0, UDRE0\n\t" // (2) Wait for send register to be empty
    // "               rjmp    \\fail\n\t"
    "               sts     UDR0, \\src\n\t" // (1) Write byte to uart
    ".endm\n\t"

    //
    // Store bit in current byte
    //
    // Arguments:
    //  - bit: bit value (0 or 1)
    //  - done: label to jump to if this was the last bit of the last byte
    // Inputs/Outputs:
    //  - r23: bit counter
    //  - r24: working byte
    //  - r30:r31: byte counter
    //  - X:  pointer to buffer
    ".macro STOREBIT bit:req, done:req\n\t"
    "               lsl     r24\n\t" // Shift previous bits to the left
    "   .if \\bit == 1\n\t"          // If bit is 1, set LSB
    "               ori     r24, 1\n\t"
    "   .endif\n\t"
    "               dec     r23\n\t"     // Decrement bit counter
    "               brne    .+12\n\t"    // Skip end of byte if bit counter is not 0
                                         // End of byte code
    "               ldi     r23, 8\n\t"  // Reset bit counter
    "               st      X+, r24\n\t" // Store current byte
    "               subi    r30, 1\n\t"  // Decrement byte counter
    "               sbci    r31, 0\n\t"
    "               brpl    .+2\n\t" // Check if byte counter is still positive (more bytes to read)
    "               rjmp    \\done\n\t"
    ".endm\n\t"

    //
    // Store bit in current byte with serial output
    //
    // Arguments:
    //  - bit: bit value (0 or 1)
    //  - done: label to jump to if this was the last bit of the last byte
    //  - sendfail: label to jump to if immediate serial out fails
    // Inputs/Outputs:
    //  - r23: bit counter
    //  - r24: working byte
    //  - r30:r31: byte counter
    //  - X: pointer to buffer
    ".macro STOREBIT2 bit:req, done:req, sendfail:req\n\t"
    "               lsl     r24\n\t" // Shift previous bits to the left
    "   .if \\bit == 1\n\t"          // If bit is 1, set LSB
    "               ori     r24, 1\n\t"
    "   .endif\n\t"
    "               dec     r23\n\t"              // Decrement bit counter
    "               brne    .+12\n\t"             // Skip end of byte if bit counter is not 0
                                                  // End of byte code
    "               ldi     r23, 8\n\t"           // Reset bit counter
    "               SENDBYTE r24, \\sendfail\n\t" // Send byte out to Serial
    "               st      X+, r24\n\t"          // Store current byte
    "               subi    r30, 1\n\t"           // Decrement byte counter
    "               sbci    r31, 0\n\t"
    "               brpl    .+2\n\t" // Check if byte counter is still positive (more bytes to read)
    "               rjmp    \\done\n\t"
    ".endm\n\t");

Floppy::Floppy()
{

    initialized = false;
    cur_track = 0;
    motor_on = false;
    save_last_op_time();
}

void Floppy::select_drive(bool selected)
{
    digitalWrite(PIN_DRVSEL, !selected);
}

void Floppy::set_direction(bool direction)
{
    digitalWrite(PIN_DIR, !direction);
}

void Floppy::set_head(unsigned short head)
{
    digitalWrite(PIN_HEADSEL, (head == 0));
}

void Floppy::set_motor_state(bool state)
{
    // Only perform operation if requested state is different from current
    if (motor_on != state)
    {
        motor_on = state;
        digitalWrite(PIN_MOTOR, !state);

        if (state)
            delay(100); // Small delay to allow motor to spin up
    }
}

void Floppy::step(unsigned short n)
{
    // Do N steps
    for (unsigned short i = 0; i < n; i++)
    {
        digitalWrite(PIN_STEP, LOW);
        delay(STEP_PULSE);
        digitalWrite(PIN_STEP, HIGH);
        delay(STEP_DELAY);
    }
}

bool Floppy::go_to_track_0()
{

    select_drive(true);
    set_direction(false);

    // Wait for drive to recognize selection
    delay(100);

    // Step until drive reports track 0
    unsigned int cnt = 0;
    while (digitalRead(PIN_TRACK0) && cnt < TRACKS)
    {
        step(1);
        cnt++;
    }

    select_drive(false);

    // Operation was succesful if track 0 was reached
    // in at most 80 steps
    return cnt < TRACKS;
}

byte Floppy::read_data(byte *buffer, unsigned int n)
{

    // Setup timer 1 for pulse timing (prescaler = 1, input capture on falling edge)
    TCCR1A = 0;
    TCCR1B = bit(CS10);
    TCCR1C = 0;

    // Setup timer 2 for exiting in case no pulses are received
    TCCR2A = 0;
    TCCR2B = bit(CS22) | bit(CS21) | bit(CS20); // Prescaler 1024 -> roughly 1 ms period

    // Select drive
    select_drive(true);

    byte ec; // Error code

    asm volatile(
        // Pulse lengths
        "               ldi     r20, 40\n\t" // Minimum medium pulse
        "               ldi     r21, 56\n\t" // Minimum long pulse

        // Default error code = 0 -> OK
        "               eor     %[ec], %[ec]\n\t"
        "               ldi     r19, 0\n\t"     // Used for resetting timeout
        "               sts     TCNT2, r19\n\t" // Clear timeout
        "               sbi     TIFR2, TOV\n\t" // Clear timeout flag

        // Find sync sequence (80 zeroes -> 80 S pulses)
        "   syncstart:\n\t"
        "               ldi     r22, 80\n\t" // Number of short pulses in sync sequence
        "   synclp:\n\t"
        "               MSPULSE\n\t"              // Measure pulse length
        "               JPDIFF  S, syncstart\n\t" // If pulse was not short, go back to start
        "               dec     r22\n\t"          // Count pulse
        "               brne    synclp\n\t"       // If counter is still not 0, continue reading pulses

        //  Initial sync sequence found, consume remaining zeroes
        "   remzlp:\n\t"
        "               MSPULSE\n\t"                              // Measure pulse length
        "               PSWITCH remzlp, markstart, syncstart\n\t" // S -> continue consuming, M -> start reading sync mark, L -> start from beginning

        // Now we're synced to the initial sync sequence, we have to find the sync mark
        // 0xA1A1A1 with special clock bit missing
        // Pulse sequence is LMLMSLMLMSLMLM
        "   markstart:"
        "               MSPULSE\n\t" // Read all sync sequence and go back to start if any anomaly found
        "               JPDIFF      L, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      L, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      S, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      L, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      L, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      S, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      L, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      L, syncstart\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart\n\t"

        // HOORRAY! We're synced! We can start reading some actual data
        "               ldi         r23, 8\n\t" // Initialize bit counter
        "   rdo:        MSPULSE\n\t"            // Read Odd
        "               PSWITCH     rdoS, rdoM, rdoL\n\t"
        "   rdoS:       STOREBIT    1, rddone\n\t"        // Store 1 bit
        "               rjmp        rdo\n\t"              // Go to read odd
        "   rdoM:       STOREBIT    0, rddone\n\t"        // Store 0 bit
        "               rjmp        rde\n\t"              // Go to read even
        "   rdoL:       STOREBIT    0, rddone\n\t"        // Store 0 bit
        "               STOREBIT    1, rddone\n\t"        // Store 1 bit
        "               rjmp        rdo\n\t"              // Go to read odd
        "   rde:        MSPULSE\n\t"                      // Read Even
        "               PSWITCH     rdeS, rdeM, rdeM\n\t" // Long pulse should never happen here, consider it Medium
        "   rdeS:       STOREBIT    0, rddone\n\t"        // Store 0 bit
        "               rjmp        rde\n\t"              // Go to read even
        "   rdeM:       STOREBIT    0, rddone\n\t"        // Store 0 bit
        "               STOREBIT    1, rddone\n\t"        // Store 1 bit
        "               rjmp        rdo\n\t"              // Go to read odd

        "   no_pulse:   ldi         r17, 1\n\t"     // Use r17 because its value is not needed anymore
        "               mov         %[ec], r17\n\t" // No pulse in timeout: ec = 1 -> NO_PULSE
        "               rjmp        rddone\n\t"

        // Read done!
        "   rddone: \n\t"
        : [ec] "=r"(ec)
        : [outbuf] "x"(buffer), "z"(n - 1) // -1 because STOREBIT exits when r30:r31 is NEGATIVE
        : "r0", "r17", "r18", "r20", "r21", "r22", "r23", "r24", "r19");

    select_drive(false);

    // Stop timers
    TCCR1B = 0;
    TCCR2B = 0;

    return ec;
}

byte Floppy::read_data_serialout(byte *buffer, unsigned int n)
{

    // Setup timer 1 for pulse timing (prescaler = 1, input capture on falling edge)
    TCCR1A = 0;
    TCCR1B = bit(CS10);
    TCCR1C = 0;

    // Setup timer 2 for exiting in case no pulses are received
    TCCR2A = 0;
    TCCR2B = bit(CS22) | bit(CS21) | bit(CS20); // Prescaler 1024 -> roughly 1 ms period

    // Select drive
    select_drive(true);

    byte ec; // Error code
    int bytes_remaining = 0;

    int bytes_as_int = n;

    asm volatile(

        // Pulse lengths
        "               ldi     r20, 40\n\t" // Minimum medium pulse
        "               ldi     r21, 56\n\t" // Minimum long pulse

        // Default error code = 0 -> OK
        "               eor     %[ec], %[ec]\n\t"
        "               ldi     r19, 0\n\t"     // Used for resetting timeout
        "               sts     TCNT2, r19\n\t" // Clear timeout
        "               sbi     TIFR2, TOV\n\t" // Clear timeout flag

        // Find sync sequence (80 zeroes -> 80 S pulses)
        "   syncstart2:\n\t"
        "               ldi     r22, 80\n\t" // Number of short pulses in sync sequence
        "   synclp2:\n\t"
        "               MSPULSE\n\t"               // Measure pulse length
        "               JPDIFF  S, syncstart2\n\t" // If pulse was not short, go back to start
        "               dec     r22\n\t"           // Count pulse
        "               brne    synclp2\n\t"       // If counter is still not 0, continue reading pulses

        //  Initial sync sequence found, consume remaining zeroes
        "   remzlp2:\n\t"
        "               MSPULSE\n\t"                                 // Measure pulse length
        "               PSWITCH remzlp2, markstart2, syncstart2\n\t" // S -> continue consuming, M -> start reading sync mark, L -> start from beginning

        // Now we're synced to the initial sync sequence, we have to find the sync mark
        // 0xA1A1A1 with special clock bit missing
        // Pulse sequence is LMLMSLMLMSLMLM
        "   markstart2:"
        "               MSPULSE\n\t" // Read all sync sequence and go back to start if any anomaly found
        "               JPDIFF      L, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      L, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      S, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      L, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      L, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      S, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      L, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      L, syncstart2\n\t"
        "               MSPULSE\n\t"
        "               JPDIFF      M, syncstart2\n\t"

        // HOORRAY! We're synced! We can start reading some actual data
        "               ldi         r23, 8\n\t" // Initialize bit counter
        "   rdo2:       MSPULSE\n\t"            // Read Odd
        "               PSWITCH     rdoS2, rdoM2, rdoL2\n\t"
        "   rdoS2:      STOREBIT2   1, rddone2, senderr\n\t"  // Store 1 bit
        "               rjmp        rdo2\n\t"                 // Go to read odd
        "   rdoM2:      STOREBIT2   0, rddone2, senderr\n\t"  // Store 0 bit
        "               rjmp        rde2\n\t"                 // Go to read even
        "   rdoL2:      STOREBIT2   0, rddone2, senderr\n\t"  // Store 0 bit
        "               STOREBIT2   1, rddone2, senderr\n\t"  // Store 1 bit
        "               rjmp        rdo2\n\t"                 // Go to read odd
        "   rde2:       MSPULSE\n\t"                          // Read Even
        "               PSWITCH     rdeS2, rdeM2, rdeM2\n\t"  // Long pulse should never happen here, consider it Medium
        "   rdeS2:      STOREBIT2    0, rddone2, senderr\n\t" // Store 0 bit
        "               rjmp        rde2\n\t"                 // Go to read even
        "   rdeM2:      STOREBIT2    0, rddone2, senderr\n\t" // Store 0 bit
        "               STOREBIT2    1, rddone2, senderr\n\t" // Store 1 bit
        "               rjmp        rdo2\n\t"                 // Go to read odd

        "   no_pulse2:  ldi         r17, 1\n\t"     // Use r17 because its value is not needed anymore
        "               mov         %[ec], r17\n\t" // No pulse in timeout: ec = 1 -> NO_PULSE
        "               rjmp        rddone2\n\t"

        "   senderr:    ldi         r17, 2\n\t"     // Use r17 because its value is not needed anymore
        "               mov         %[ec], r17\n\t" // No pulse in timeout: ec = 1 -> NO_PULSE

        // Read done!
        "   rddone2: \n\t" // Re increment byte counter by 1 (so when all bytes are read it will be 0)
        // "   rddone:     adiw        r30, 1\n\t" // Re increment byte counter by 1 (so when all bytes are read it will be 0)
        // "               ldi         r30, 0\n\t"
        // "               ldi         r31, 0\n\t"
        //
        : [ec] "=r"(ec), [bytes_rem] "=z"(bytes_remaining)
        // : [ec] "=r"(ec)
        : [outbuf] "x"(buffer), "z"(bytes_as_int - 1) // -1 because STOREBIT exits when r30:r31 is NEGATIVE
        : "r0", "r17", "r18", "r20", "r21", "r22", "r23", "r24", "r19");

    select_drive(false);

    // Stop timers
    TCCR1B = 0;
    TCCR2B = 0;

    // Send remaining bytes
    for (; bytes_remaining > 0; bytes_remaining--)
    {
        serial_send_byte_immediate(0x00);
    }

    return ec;
}

void Floppy::save_last_op_time()
{
    last_op_time = millis();
}

void Floppy::setup()
{
    pinMode(PIN_DRVSEL, OUTPUT);
    pinMode(PIN_DIR, OUTPUT);
    pinMode(PIN_STEP, OUTPUT);
    pinMode(PIN_HEADSEL, OUTPUT);
    pinMode(PIN_MOTOR, OUTPUT);

    pinMode(PIN_TRACK0, INPUT_PULLUP);
    pinMode(PIN_READDATA, INPUT_PULLUP);

    motor_on = true;
    set_motor_state(false);
    select_drive(false);
    save_last_op_time();
}

FloppyError Floppy::initialize()
{

    // Go to track 0
    if (!go_to_track_0())
        return FloppyError::TRACK0_NOT_FOUND;

    initialized = true;
    cur_track = 0;

    return FloppyError::OK;
}

FloppyError Floppy::seek(byte track)
{
    // Check if drive initialized
    if (!initialized)
        return FloppyError::NOT_INITIALIZED;

    // Check if track number is in range
    if (track >= TRACKS)
        return FloppyError::TRACK_OUT_OF_RANGE;

    // See if we need to go forward or backward
    bool dir = track > cur_track;

    // Select drive
    select_drive(true);

    // Set direction
    set_direction(dir);

    // STEP
    step(abs((short)track - (short)cur_track));

    // Release drive
    select_drive(false);

    cur_track = track;

    return FloppyError::OK;
}

void write_filler_data(unsigned int amount)
{
    for (unsigned int i = 0; i < amount; i++)
    {
        serial_send_byte_immediate(0x00);
    }
}

FloppyError Floppy::find_sector(byte cylinder, byte head, byte sector)
{
    FloppyError ec;

    // Select head
    set_head(head);

    // Turn on motor
    set_motor_state(true);

    // Seek to track
    ec = seek(cylinder);

    Serial.flush();
    noInterrupts();

    if (ec != FloppyError::OK)
    {
        // interrupts();
        return ec;
    }

    unsigned short attempts = 50;
    byte tmpbuf[7];

    // Read data
    while (attempts > 0)
    {

        // Try to read Address Block
        if (read_data(tmpbuf, 7) > 0)
        {
            attempts--;
            continue;
        }

        // Check if what we read was an address block
        if (tmpbuf[0] == 0xFE)
        {
            if (tmpbuf[1] != cylinder)
            {
                cur_track = tmpbuf[1]; // Save our current position
                                       // Serial.print("Seek error: ");
                                       // Serial.println(tmpbuf[1]);
                // interrupts();
                return FloppyError::SEEK_ERROR;
            }

            if (tmpbuf[2] == head && tmpbuf[3] == sector)
            {
                // Check CRC
                if (calc_crc(tmpbuf, 5) == 256u * tmpbuf[5] + tmpbuf[6])
                    break;
            }
        }

        attempts--;
    }

    if (attempts == 0)
    {
        // write_filler_data(SECTOR_SIZE + 3);
        // interrupts();
        // Serial.println("Sector not found!");
        return FloppyError::SECTOR_NOT_FOUND;
    }

    return FloppyError::OK;
}

FloppyError Floppy::read_sector(byte *buffer, byte cylinder, byte head, byte sector)
{
    FloppyError ec;

    save_last_op_time();

    // Find sector
    ec = find_sector(cylinder, head, sector);

    if (ec != FloppyError::OK)
    {
        // write_filler_data(SECTOR_SIZE + 3);
        interrupts();
        return ec;
    }

    serial_send_byte_immediate('D');

    // We've read the correct address block
    // Read actual data
    byte err = read_data_serialout(buffer, SECTOR_SIZE + 3);
    // if (byte ec = read_data(buffer, SECTOR_SIZE + 3) > 0)
    if (err > 0)
    {
        interrupts();
        // Serial.print("\nEC: ");
        // Serial.println(err);
        return FloppyError::NO_PULSE;
    }

    // Data mark is not correct
    if (buffer[0] != 0xFB)
    {
        interrupts();
        // Serial.println("Incorrect data mark");
        return FloppyError::INCORRECT_DATA_MARK;
    }

    // Check CRC
    if (calc_crc(buffer, 513) != 256u * buffer[513] + buffer[514])
    {

        interrupts();
        // Serial.println("CRC Error");
        return FloppyError::CRC;
    }

    interrupts();

    return FloppyError::OK;
}

void Floppy::auto_motor_off()
{
    if (motor_on && (millis() - last_op_time) > MOTOR_OFF_TIMEOUT)
    {
        set_motor_state(false);
    }
}

void Floppy::test_serial()
{
    Serial.flush();
    noInterrupts();
    for (int i = 0; i < 100; i++)
    {
        serial_send_byte_immediate('$');
    }
    interrupts();
}
