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
#include "./src/dragino/raspberry.c"

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
        printf("invalid exit code: %d\n", exitCode);
        exit(1);
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
