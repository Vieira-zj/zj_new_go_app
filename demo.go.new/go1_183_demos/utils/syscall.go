package utils

import (
	"fmt"
	"syscall"
)

func GetParentProcessId() int {
	return syscall.Getppid()
}

func KillProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

func PrintAppUsage() error {
	// syscall: getrusage
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return err
	}

	fmt.Printf("User CPU time used: %+v \n", usage.Utime)
	fmt.Printf("System CPU time used: %+v \n", usage.Stime)
	fmt.Printf("Maximum resident set size: %v \n", usage.Maxrss)
	fmt.Printf("Integral shared memory size: %v \n", usage.Ixrss)
	fmt.Printf("Integral unshared data size: %v \n", usage.Idrss)
	fmt.Printf("Integral unshared stack size: %v \n", usage.Isrss)
	fmt.Printf("Page reclaims (soft page faults): %v\n", usage.Minflt)
	fmt.Printf("Page faults (hard page faults): %v\n", usage.Majflt)
	fmt.Printf("Swaps: %v\n", usage.Nswap)
	fmt.Printf("Block input operations: %v\n", usage.Inblock)
	fmt.Printf("Block output operations: %v\n", usage.Oublock)
	fmt.Printf("IPC messages sent: %v\n", usage.Msgsnd)
	fmt.Printf("IPC messages received: %v\n", usage.Msgrcv)
	fmt.Printf("Signals received: %v\n", usage.Nsignals)
	fmt.Printf("Voluntary context switches: %v\n", usage.Nvcsw)
	fmt.Printf("Involuntary context switches: %v\n", usage.Nivcsw)

	return nil
}
