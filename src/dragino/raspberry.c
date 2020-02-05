/*******************************************************************************
 *
 * Copyright (c) 2018 Dragino
 * http://www.dragino.com
 *
 * Copyright (c) 2020 RTrade Technologies (modificiations)
 * https://temporal.cloud
 * 
 *******************************************************************************/

#include <stdio.h>
#include <stdbool.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <string.h>
#include <sys/time.h>
#include <signal.h>
#include <stdlib.h>
#include <sys/ioctl.h>
#include <wiringPi.h>
#include <wiringPiSPI.h>
#include "../../include/dragino/raspberry.h"
#include "../../include/array_len/array_len.h"

/*******************************************************************************
 *
 * Configure these values!
 *
 *******************************************************************************/

/* SX1272 - Raspberry connections */
int ssPin = 6;
int dio0  = 7;
int RST   = 0;

/* Set spreading factor (SF7 - SF12) */
enum sf_t sf = SF7;

/* Sets center frequency - 868.1 Mhz*/
uint32_t  freq = 868100000;

byte hello[32] = "HELLO";

void die(const char *s) {
    perror(s);
    exit(1);
}

void selectreceiver() {
    digitalWrite(ssPin, LOW);
}

void unselectreceiver() {
    digitalWrite(ssPin, HIGH);
}

byte readReg(byte addr) {
    unsigned char spibuf[2];

    selectreceiver();
   
    spibuf[0] = addr & 0x7F;
    spibuf[1] = 0x00;
   
    wiringPiSPIDataRW(CHANNEL, spibuf, 2);
   
    unselectreceiver();

    return spibuf[1];
}

void writeReg(byte addr, byte value) {
    unsigned char spibuf[2];

    spibuf[0] = addr | 0x80;
    spibuf[1] = value;
    
    selectreceiver();

    wiringPiSPIDataRW(CHANNEL, spibuf, 2);

    unselectreceiver();
}

static void opmode (uint8_t mode) {
    writeReg(REG_OPMODE, (readReg(REG_OPMODE) & ~OPMODE_MASK) | mode);
}

static void opmodeLora() {
    uint8_t u = OPMODE_LORA;
  
    /* TBD: sx1276 high freq */
    if (sx1272 == false) {
        u |= 0x8;   
    }

    writeReg(REG_OPMODE, u);
}


void setupLoRa() {
    byte version;
    uint64_t frf;    
    
    digitalWrite(RST, HIGH);
    delay(100);
    digitalWrite(RST, LOW);
    delay(100);

    version = readReg(REG_VERSION);
    if (version == 0x22) {
        /* sx1272 */
        printf("SX1272 detected, starting.\n");
        sx1272 = true;
    } else {
        /* sx1276? */
        digitalWrite(RST, LOW);
        delay(100);
        digitalWrite(RST, HIGH);
        delay(100);
        version = readReg(REG_VERSION);
        if (version == 0x12) {
            /* sx1276 */
            printf("SX1276 detected, starting.\n");
            sx1272 = false;
        } else {
            printf("Unrecognized transceiver.\n");
            /* printf("Version: 0x%x\n",version); */
            exit(1);
        }
    }

    opmode(OPMODE_SLEEP);

    /* set frequency */
    frf = ((uint64_t)freq << 19) / 32000000;
    writeReg(REG_FRF_MSB, (uint8_t)(frf>>16) );
    writeReg(REG_FRF_MID, (uint8_t)(frf>> 8) );
    writeReg(REG_FRF_LSB, (uint8_t)(frf>> 0) );

    /* LoRaWAN public sync word */
    writeReg(REG_SYNC_WORD, 0x34); 

    if (sx1272) {
        if (sf == SF11 || sf == SF12) {
            writeReg(REG_MODEM_CONFIG,0x0B);
        } else {
            writeReg(REG_MODEM_CONFIG,0x0A);
        }
        writeReg(REG_MODEM_CONFIG2,(sf<<4) | 0x04);
    } else {
        if (sf == SF11 || sf == SF12) {
            writeReg(REG_MODEM_CONFIG3,0x0C);
        } else {
            writeReg(REG_MODEM_CONFIG3,0x04);
        }
        writeReg(REG_MODEM_CONFIG,0x72);
        writeReg(REG_MODEM_CONFIG2,(sf<<4) | 0x04);
    }

    if (sf == SF10 || sf == SF11 || sf == SF12) {
        writeReg(REG_SYMB_TIMEOUT_LSB,0x05);
    } else {
        writeReg(REG_SYMB_TIMEOUT_LSB,0x08);
    }

    writeReg(REG_MAX_PAYLOAD_LENGTH,0x80);
    writeReg(REG_PAYLOAD_LENGTH,PAYLOAD_LENGTH);
    writeReg(REG_HOP_PERIOD,0xFF);
    writeReg(REG_FIFO_ADDR_PTR, readReg(REG_FIFO_RX_BASE_AD));
    writeReg(REG_LNA, LNA_MAX_GAIN);
}

boolean receive(char *payload) {
    int i, irqflags;
    /* clear rxDone */
    writeReg(REG_IRQ_FLAGS, 0x40);
    irqflags = readReg(REG_IRQ_FLAGS);

    /*  payload crc: 0x20 */
    if ((irqflags & 0x20) == 0x20) {
        printf("CRC error\n");
        writeReg(REG_IRQ_FLAGS, 0x20);
        return false;
    } else {
        byte currentAddr = readReg(REG_FIFO_RX_CURRENT_ADDR);
        byte receivedCount = readReg(REG_RX_NB_BYTES);
        receivedbytes = receivedCount;

        writeReg(REG_FIFO_ADDR_PTR, currentAddr);

        for (i = 0; i < receivedCount; i++) {
            payload[i] = (char)readReg(REG_FIFO);
        }
    }

    return true;
}

void receivepacket() {

    long int SNR;
    int rssicorr;

    if (digitalRead(dio0) == 1) {
        if (receive(message)) {
            /* received a message */
            byte value = readReg(REG_PKT_SNR_VALUE);
            /* The SNR sign bit is 1 */
            if ( value & 0x80 ) {
                /* Invert and divide by 4 */
                value = ( ( ~value + 1 ) & 0xFF ) >> 2;
                SNR = -value;
            }
            else {
                /* Divide by 4 */
                SNR = ( value & 0xFF ) >> 2;
            }
            
            if (sx1272) {
                rssicorr = 139;
            } else {
                rssicorr = 157;
            }

            printf("Packet RSSI: %d, ", readReg(0x1A)-rssicorr);
            printf("RSSI: %d, ", readReg(0x1B)-rssicorr);
            printf("SNR: %li, ", SNR);
            printf("Length: %i", (int)receivedbytes);
            printf("\n");
            printf("Payload: %s\n", message);

        } 
    /* dio0=1 */
    }
}

static void configPower (int8_t pw) {
    if (sx1272 == false) {
        /* no boost used for now */
        if (pw >= 17) {
            pw = 15;
        } else if(pw < 2) {
            pw = 2;
        }
        /* check board type for BOOST pin */
        writeReg(RegPaConfig, (uint8_t)(0x80|(pw&0xf)));
        writeReg(RegPaDac, readReg(RegPaDac)|0x4);

    } else {
        /* set PA config (2-17 dBm using PA_BOOST) */
        if (pw > 17) {
            pw = 17;
        } else if (pw < 2) {
            pw = 2;
        }
        writeReg(RegPaConfig, (uint8_t)(0x80|(pw-2)));
    }
}


static void writeBuf(byte addr, byte *value, byte len) {                                                       
    int i;
    unsigned char spibuf[256];                                                                          
    
    spibuf[0] = addr | 0x80;                                                                            
    
    for (i = 0; i < len; i++) {                                                                         
        spibuf[i + 1] = value[i];                                                                       
    }                                                                                                   
    
    selectreceiver();                                                                                   
    wiringPiSPIDataRW(CHANNEL, spibuf, len + 1);                                                        
    unselectreceiver();                                                                                 
}

void txlora(byte *frame, byte datalen) {
    /* set the IRQ mapping DIO0=TxDone DIO1=NOP DIO2=NOP */
    writeReg(RegDioMapping1, MAP_DIO0_LORA_TXDONE|MAP_DIO1_LORA_NOP|MAP_DIO2_LORA_NOP);
    /* clear all radio IRQ flags */
    writeReg(REG_IRQ_FLAGS, 0xFF);
    /* mask all IRQs but TxDone */
    writeReg(REG_IRQ_FLAGS_MASK, ~IRQ_LORA_TXDONE_MASK);
    /* initialize the payload size and address pointers */
    writeReg(REG_FIFO_TX_BASE_AD, 0x00);
    writeReg(REG_FIFO_ADDR_PTR, 0x00);
    writeReg(REG_PAYLOAD_LENGTH, datalen);
    /* download buffer to the radio FIFO */
    writeBuf(REG_FIFO, frame, datalen);
    /* now we actually start the transmission */
    opmode(OPMODE_TX);
}

/* writeData: helper function that writes data and calculates frame size */
void writeData(byte *frame, bool debug) {
    txlora(frame, strlen((char *)frame));
    if (debug) {
        printf("send: %s\n", frame);
    }
}

/* is used to setup the LoRa transmitter */
int setup(bool sender) {
    if (!wiringPiSetup()) {
        return 1;
    }
    
    pinMode(ssPin, OUTPUT);
    pinMode(dio0, INPUT);
    pinMode(RST, OUTPUT);
    
    if (!wiringPiSPISetup(CHANNEL, 500000)) {
        return 1;
    }
    
    setupLoRa();
    
    if (sender) {
       /* radio init */
        opmodeLora();
        opmode(OPMODE_STANDBY);
        opmode(OPMODE_RX);
        return 0;
    }
    
    opmodeLora();
    /* enter standby mode (required for FIFO loading)) */
    opmode(OPMODE_STANDBY);
    /* set PA ramp-up time 50 uSec */
    writeReg(RegPaRamp, (readReg(RegPaRamp) & 0xF0) | 0x08); 
    /* set radio power */
    configPower(23);
    
    return 0;
}

/* an example of how this library can be used */
int main (int argc, char *argv[]) {
    bool isSender;
    int exitCode;

    if (argc < 2) {
        printf("Usage: argv[1] sender|rec [message]\n");
        exit(1);
    }
    
    switch (strcmp("sender", argv[1])) {
        case 0:
            isSender = false;
            break;
        case 1:
            isSender = true;
            break;
        default:
            perror("invalid strcmp return value");
            exit(1);
    }
    
    exitCode = setup(isSender);
    if (!exitCode) {
        die("invalid exit code: " + exitCode);
    }
    
    if (!isSender) {
        printf("Send packets at SF%i on %.6f Mhz.\n", sf,(double)freq/1000000);
        printf("------------------\n");
        if (argc > 2) {
            strncpy((char *)hello, argv[2], sizeof(hello));
        }
        while(1) {
            writeData(hello, true);
            delay(5000);
        }
    } else {
        printf("Listening at SF%i on %.6f Mhz.\n", sf,(double)freq/1000000);
        printf("------------------\n");
        while(1) {
            receivepacket(); 
            delay(1);
        }
    }
    
    return (0);
}
