package main

/*
#include <stdio.h>
void hello() {
	printf("hello world\n");
}
*/
import "C"

func main() {
	C.hello()
}
