package gpm

import (
	"github.com/madokast/nopreempt"
)

func GetGId() int64 {
	return nopreempt.GetGID()
}

func EnablePreempt(mp nopreempt.MP) {
	mp.Release()
}

func DisablePreempt() nopreempt.MP {
	return nopreempt.AcquireM()
}
