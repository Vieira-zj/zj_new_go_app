package utils

import (
	"os"
	"syscall"
)

// FileLock is a file lock for process, but not thread or goroutine.
type FileLock struct {
	f *os.File
}

func (l *FileLock) Lock() error {
	f, err := os.OpenFile("/tmp/testlock.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	l.f = f
	return syscall.Flock(int(f.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
}

func (l *FileLock) Unlock() error {
	if l.f != nil {
		defer l.f.Close()
		return syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
	}

	return nil
}
