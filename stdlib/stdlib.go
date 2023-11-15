package stdlib

/*
#include<stddef.h>
#include<stdlib.h>
#include<stdio.h>

typedef void* memory;

memory mymalloc(size_t sz, char debug) {
	memory m;
	if (debug) {
		printf("%s:%d malloc %llu bytes ", __FILE__, __LINE__, (unsigned long long int)sz);
	}
	m = (memory)malloc(sz);
	if (debug) {
		printf("at 0x%llX\n", (unsigned long long int)m);
		fflush(stdout);
	}
	return m;
}

void myfree(memory m, char debug) {
	if (debug) {
		printf("%s:%d free memory 0x%llX\n", __FILE__, __LINE__, (unsigned long long int)m);
		fflush(stdout);
	}
	free((void*)m);
}
*/
import "C"
import (
	"fmt"

	"github.com/madokast/direct/config"
)

type Memory struct {
	Address uintptr // no gc scan
	Size    uint
}

const vebose = config.VEBOSE

func Malloc(size uint) Memory {
	var debug C.char = 0
	if vebose {
		debug = 1
	}
	return Memory{Address: uintptr(C.mymalloc(C.size_t(size), debug)), Size: size}
}

func (m Memory) Free() {
	var debug C.char = 0
	if vebose {
		debug = 1
	}
	C.myfree(C.memory(m.Address), debug)
}

func (m Memory) ToString() string {
	return fmt.Sprintf("addr:0x%X, size:%dB", m.Address, m.Size)
}
