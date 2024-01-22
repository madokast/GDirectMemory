package managed_memory

import (
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"gitlab.grandhoo.com/rock/storage/storage2/utils/test"
	"testing"
)

func TestNewStringFactory(t *testing.T) {
	memory := New(1 * MB)
	defer memory.Free()

	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	factory := NewStringFactory(&concurrentMemory)
	defer factory.Destroy()

	s1, err := factory.CreateFromGoString("hello")
	test.PanicErr(err)
	defer s1.Free(&concurrentMemory)

	logger.Info(s1.Length())
	logger.Info(s1.AsGoString())
	logger.Info(s1.CopyToGoString())
	logger.Info(s1)

	test.Assert(s1.CopyToGoString() == "hello")
}

func TestString_Hashcode(t *testing.T) {
	memory := New(1 * MB)
	defer memory.Free()

	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	factory := NewStringFactory(&concurrentMemory)
	defer factory.Destroy()

	s1, err := factory.CreateFromGoString("hello")
	test.PanicErr(err)
	defer s1.Free(&concurrentMemory)

	test.Assert(s1.CopyToGoString() == "hello")

	logger.Info(s1.Hashcode())
}
