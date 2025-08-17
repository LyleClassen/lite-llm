package monitor

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type PerformanceMetrics struct {
	CPUUsage            float64
	MemoryUsedMB        int
	MemoryTotalMB       int
	MemoryUsagePercent  float64
	GPUUsage            float64
	GPUMemoryUsedMB     int
	GPUMemoryTotalMB    int
	Timestamp           time.Time
}

func GetPerformanceMetrics() *PerformanceMetrics {
	metrics := &PerformanceMetrics{
		Timestamp: time.Now(),
		GPUUsage:  -1, // -1 indicates unavailable
	}

	// Get CPU usage
	cpuUsage, err := getCPUUsage()
	if err != nil {
		logrus.Warnf("Failed to get CPU usage: %v", err)
	} else {
		metrics.CPUUsage = cpuUsage
	}

	// Get memory usage
	memUsed, memTotal, err := getMemoryUsage()
	if err != nil {
		logrus.Warnf("Failed to get memory usage: %v", err)
	} else {
		metrics.MemoryUsedMB = memUsed
		metrics.MemoryTotalMB = memTotal
		if memTotal > 0 {
			metrics.MemoryUsagePercent = float64(memUsed) / float64(memTotal) * 100
		}
	}

	// Get GPU usage (AMD-specific)
	gpuUsage, gpuMemUsed, gpuMemTotal := getAMDGPUUsage()
	if gpuUsage >= 0 {
		metrics.GPUUsage = gpuUsage
		metrics.GPUMemoryUsedMB = gpuMemUsed
		metrics.GPUMemoryTotalMB = gpuMemTotal
	}

	return metrics
}

func getCPUUsage() (float64, error) {
	// Read /proc/stat for CPU usage
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, err
	}

	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) < 8 || fields[0] != "cpu" {
		return 0, err
	}

	// Parse CPU times
	var idle, total uint64
	for i := 1; i < len(fields); i++ {
		val, err := strconv.ParseUint(fields[i], 10, 64)
		if err != nil {
			return 0, err
		}
		total += val
		if i == 4 { // idle time is the 4th field
			idle = val
		}
	}

	if total == 0 {
		return 0, nil
	}

	// Calculate usage percentage
	usage := float64(total-idle) / float64(total) * 100
	return usage, nil
}

func getMemoryUsage() (int, int, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	var memTotal, memFree, memBuffers, memCached int

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		switch key {
		case "MemTotal":
			memTotal = value / 1024 // Convert KB to MB
		case "MemFree":
			memFree = value / 1024
		case "Buffers":
			memBuffers = value / 1024
		case "Cached":
			memCached = value / 1024
		}
	}

	memUsed := memTotal - memFree - memBuffers - memCached
	return memUsed, memTotal, nil
}

func getAMDGPUUsage() (float64, int, int) {
	// Try to get AMD GPU usage from sysfs
	// This is a simplified implementation - in practice you might want to use
	// tools like radeontop or parse more detailed GPU statistics

	// Check for AMD GPU memory usage
	gpuMemTotal := getAMDGPUMemory("/sys/class/drm/card0/device/mem_info_vram_total")
	gpuMemUsed := getAMDGPUMemory("/sys/class/drm/card0/device/mem_info_vram_used")

	if gpuMemTotal > 0 {
		totalMB := int(gpuMemTotal / (1024 * 1024))
		usedMB := int(gpuMemUsed / (1024 * 1024))
		
		// Try to get GPU usage percentage
		usage := getAMDGPUUsagePercent()
		
		return usage, usedMB, totalMB
	}

	return -1, 0, 0
}

func getAMDGPUMemory(path string) int64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	value, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0
	}

	return value
}

func getAMDGPUUsagePercent() float64 {
	// Try to read GPU usage from sysfs
	// Note: This path may vary depending on the GPU and driver version
	usagePaths := []string{
		"/sys/class/drm/card0/device/gpu_busy_percent",
		"/sys/class/drm/card0/device/busy_percent",
	}

	for _, path := range usagePaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		usage, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
		if err != nil {
			continue
		}

		return usage
	}

	// If we can't get usage directly, estimate based on memory usage
	// This is not accurate but provides some indication
	return -1
}