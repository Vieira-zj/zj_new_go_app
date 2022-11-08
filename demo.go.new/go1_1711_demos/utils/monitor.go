package utils

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
)

var (
	pMonitor     ProcessMonitor
	pMonitorOnce sync.Once
)

// ProcessMonitor monitors app process resource usage, like cpu, mem, and number of goroutines.
type ProcessMonitor struct {
	client *process.Process
}

func NewProcessMonitor() (ProcessMonitor, error) {
	var err error
	pMonitorOnce.Do(func() {
		p, perr := process.NewProcess(int32(os.Getpid()))
		if perr != nil {
			err = perr
		}
		pMonitor = ProcessMonitor{
			client: p,
		}
	})
	return pMonitor, err
}

func (m ProcessMonitor) GetCpuPercent(ctx context.Context) (float64, error) {
	return m.client.PercentWithContext(ctx, time.Second)
}

// GetCpuPercentOfSingleCore 获取 cpu 时间占单个核心的比例。
func (m ProcessMonitor) GetCpuPercentOfSingleCore(ctx context.Context) (float64, error) {
	percent, err := m.GetCpuPercent(ctx)
	if err != nil {
		return -1, err
	}
	return percent / float64(runtime.NumCPU()), nil
}

// GetMemoryPercent 获取进程占用内存的比例。
func (m ProcessMonitor) GetMemoryPercent(ctx context.Context) (float32, error) {
	return m.client.MemoryPercentWithContext(ctx)
}

// GetNumOfSysThreads 获取创建的线程数。
func (m ProcessMonitor) GetNumOfSysThreads() int {
	return pprof.Lookup("threadcreate").Count()
}

func (m ProcessMonitor) GetNumOfGoroutine() int {
	return runtime.NumGoroutine()
}

func (m ProcessMonitor) GetCpuPercentOfSingleCoreInContainer(ctx context.Context) (float64, error) {
	cpuPeriod, err := readUint("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
	if err != nil {
		return -1, err
	}
	cpuQuota, err := readUint("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
	if err != nil {
		return -1, err
	}
	cpuNum := float64(cpuPeriod) / float64(cpuQuota)

	cpuPercent, err := m.client.PercentWithContext(ctx, time.Second)
	if err != nil {
		return -1, err
	}

	return cpuPercent / cpuNum, nil
}

func (m ProcessMonitor) GetMemoryPercentInContainer(ctx context.Context) (float32, error) {
	memLimit, err := readUint("/sys/fs/cgroup/memory/memory.limit_in_bytes")
	if err != nil {
		return -1, err
	}

	memInfo, err := m.client.MemoryInfoWithContext(ctx)
	if err != nil {
		return -1, err
	}

	return float32(memInfo.RSS*100) / float32(memLimit), nil
}

func (m ProcessMonitor) PrettyString(ctx context.Context) (string, error) {
	cpu, err := m.GetCpuPercentOfSingleCore(ctx)
	if err != nil {
		return "", err
	}
	mem, err := m.GetMemoryPercent(ctx)
	if err != nil {
		return "", err
	}
	threads := m.GetNumOfSysThreads()
	goroutines := m.GetNumOfGoroutine()

	s := fmt.Sprintf("cpu=%.2f%%,mem=%.2f%%,threads_count=%d,goroutine_count=%d", cpu, mem, threads, goroutines)
	return s, nil
}

// helper

func readUint(path string) (uint64, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, nil
	}

	s := strings.TrimSpace(string(b))
	return parseUint(s, 10, 64)
}

func parseUint(s string, base, bitSize int) (uint64, error) {
	v, err := strconv.ParseUint(s, base, bitSize)
	if err != nil {
		intValue, intErr := strconv.ParseInt(s, base, bitSize)
		// 1. Handle negative values greater than MinInt64 (and)
		// 2. Handle negative values lesser than MinInt64
		if intErr == nil && intValue < 0 {
			return 0, nil
		} else if intErr != nil && intErr.(*strconv.NumError).Err == strconv.ErrRange && intValue < 0 {
			return 0, nil
		}
		return 0, err
	}
	return v, nil
}
