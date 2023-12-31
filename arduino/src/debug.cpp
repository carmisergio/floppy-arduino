#include "debug.h"

#ifdef ENABLE_DEBUG
SoftwareSerial debugSerial(DEBUG_TX_PIN, DEBUG_RX_PIN);
#endif