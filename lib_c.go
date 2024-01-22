//go:build cgo

package managed_memory

/*
#include<stddef.h> // size_t
#include<stdlib.h>

//#define log
#ifdef log
#include<stdio.h>
#endif

typedef void* memory;
typedef unsigned long long int ull;

memory lib_malloc(size_t sz) {
	memory m;
	m = (memory)malloc(sz);
#ifdef log
    printf("%s:%d malloc %llu bytes at 0x%llX\n", __FILE__, __LINE__, (ull)sz, (ull)m);
    fflush(stdout);
#endif
	return m;
}

memory lib_calloc(size_t sz) {
	memory m;
	m = (memory)calloc(sz, sizeof(char));
#ifdef log
    printf("%s:%d calloc %llu bytes at 0x%llX\n", __FILE__, __LINE__, (ull)sz, (ull)m);
    fflush(stdout);
#endif
	return m;
}

void lib_free(memory m) {
	free((void*)m);
#ifdef log
    printf("%s:%d free memory 0x%llX\n", __FILE__, __LINE__, (ull)m);
    fflush(stdout);
#endif
}
*/
import "C"
import (
	"fmt"
	"gitlab.grandhoo.com/rock/storage/internal/logger"
)

func libMalloc(size SizeType) pointer {
	ptr := pointer(uintptr(C.lib_malloc(C.size_t(uint(size)))))
	if ptr.IsNull() {
		logger.Panic(fmt.Sprintf("cannot alloc %d bytes (%s)", size, humanFriendlyMemorySize(size)))
	}
	return ptr
}

func libCalloc(size SizeType) pointer {
	ptr := pointer(uintptr(C.lib_calloc(C.size_t(uint(size)))))
	if ptr.IsNull() {
		logger.Panic(fmt.Sprintf("cannot alloc %d bytes (%s)", size, humanFriendlyMemorySize(size)))
	}
	return ptr
}

func libFree(ptr pointer) {
	C.lib_free(C.memory(uintptr(ptr)))
}
