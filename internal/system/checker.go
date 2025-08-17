package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type Checker struct{}

type SystemInfo struct {
	HasDocker     bool
	HasROCm       bool
	HasNVIDIA     bool
	HasAMDGPU     bool
	GPUMemory     int // in MB
	SystemMemory  int // in MB
	GPUModel      string
	GPUType       string // "nvidia", "amd", or "unknown"
	KernelVersion string
}

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) CheckAll() error {
	info, err := c.GetSystemInfo()
	if err != nil {
		return fmt.Errorf("failed to get system info: %w", err)
	}

	if err := c.validateRequirements(info); err != nil {
		return err
	}

	c.printSystemInfo(info)
	return nil
}

func (c *Checker) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// Check Docker
	info.HasDocker = c.checkDocker()

	// Check GPUs
	info.HasNVIDIA, nvidiaModel, nvidiaMemory := c.checkNVIDIAGPU()
	info.HasAMDGPU, amdModel, amdMemory := c.checkAMDGPU()
	
	// Set primary GPU info
	if info.HasNVIDIA {
		info.GPUType = "nvidia"
		info.GPUModel = nvidiaModel
		info.GPUMemory = nvidiaMemory
	} else if info.HasAMDGPU {
		info.GPUType = "amd"
		info.GPUModel = amdModel
		info.GPUMemory = amdMemory
	} else {
		info.GPUType = "unknown"
	}

	// Check ROCm
	info.HasROCm = c.checkROCm()

	// Get system memory
	info.SystemMemory = c.getSystemMemory()

	// Get kernel version
	info.KernelVersion = c.getKernelVersion()

	return info, nil
}

func (c *Checker) checkDocker() bool {
	cmd := exec.Command("docker", "--version")
	err := cmd.Run()
	return err == nil
}

func (c *Checker) checkNVIDIAGPU() (bool, string, int) {
	// Check if NVIDIA GPU is present using lspci
	cmd := exec.Command("lspci", "-v")
	output, err := cmd.Output()
	if err != nil {
		logrus.Warnf("Failed to run lspci: %v", err)
		return false, "", 0
	}

	lines := strings.Split(string(output), "\n")
	var hasNVIDIA bool
	var gpuModel string
	var memory int

	// Look for NVIDIA graphics cards
	nvidiaRegex := regexp.MustCompile(`(?i)NVIDIA.*(GeForce|RTX|GTX|Tesla|Quadro)`)
	
	for i, line := range lines {
		if strings.Contains(line, "VGA compatible controller") && nvidiaRegex.MatchString(line) {
			hasNVIDIA = true
			
			// Extract GPU model
			parts := strings.Split(line, ": ")
			if len(parts) > 1 {
				gpuModel = strings.TrimSpace(parts[1])
			}

			// Look for memory information in subsequent lines
			for j := i + 1; j < len(lines) && j < i+20; j++ {
				if strings.TrimSpace(lines[j]) == "" {
					break
				}
				
				// Look for memory size in various formats
				memRegex := regexp.MustCompile(`(?i)memory.*?(\d+)([MG])B`)
				matches := memRegex.FindStringSubmatch(lines[j])
				if len(matches) >= 3 {
					size, _ := strconv.Atoi(matches[1])
					if matches[2] == "G" {
						memory = size * 1024
					} else {
						memory = size
					}
					break
				}
			}
			break
		}
	}

	// If we couldn't get memory from lspci, try nvidia-smi
	if hasNVIDIA && memory == 0 {
		memory = c.getNVIDIAMemoryFromSMI()
	}

	return hasNVIDIA, gpuModel, memory
}

func (c *Checker) getNVIDIAMemoryFromSMI() int {
	// Try to get GPU memory from nvidia-smi
	cmd := exec.Command("nvidia-smi", "--query-gpu=memory.total", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		// Default assumption for RTX 3070 if we can't detect
		return 8192 // 8GB
	}

	memStr := strings.TrimSpace(string(output))
	if mem, err := strconv.Atoi(memStr); err == nil {
		return mem
	}

	return 8192 // Default
}

func (c *Checker) checkAMDGPU() (bool, string, int) {
	// Check if AMD GPU is present using lspci
	cmd := exec.Command("lspci", "-v")
	output, err := cmd.Output()
	if err != nil {
		logrus.Warnf("Failed to run lspci: %v", err)
		return false, "", 0
	}

	lines := strings.Split(string(output), "\n")
	var hasAMD bool
	var gpuModel string
	var memory int

	// Look for AMD/ATI graphics cards
	amdRegex := regexp.MustCompile(`(?i)(AMD|ATI).*(Radeon|RX|Ellesmere|Polaris)`)
	
	for i, line := range lines {
		if strings.Contains(line, "VGA compatible controller") && amdRegex.MatchString(line) {
			hasAMD = true
			
			// Extract GPU model
			parts := strings.Split(line, ": ")
			if len(parts) > 1 {
				gpuModel = strings.TrimSpace(parts[1])
			}

			// Look for memory information in subsequent lines
			for j := i + 1; j < len(lines) && j < i+20; j++ {
				if strings.TrimSpace(lines[j]) == "" {
					break
				}
				
				// Look for memory size in various formats
				memRegex := regexp.MustCompile(`(?i)memory.*?(\d+)([MG])B`)
				matches := memRegex.FindStringSubmatch(lines[j])
				if len(matches) >= 3 {
					size, _ := strconv.Atoi(matches[1])
					if matches[2] == "G" {
						memory = size * 1024
					} else {
						memory = size
					}
					break
				}
			}
			break
		}
	}

	// If we couldn't get memory from lspci, try alternative methods
	if hasAMD && memory == 0 {
		memory = c.getGPUMemoryAlternative()
	}

	return hasAMD, gpuModel, memory
}

func (c *Checker) getGPUMemoryAlternative() int {
	// Try to get GPU memory from /sys/class/drm
	files, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return 0
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "card") && !strings.Contains(file.Name(), "-") {
			memPath := fmt.Sprintf("/sys/class/drm/%s/device/mem_info_vram_total", file.Name())
			if data, err := os.ReadFile(memPath); err == nil {
				if size, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
					return int(size / (1024 * 1024)) // Convert bytes to MB
				}
			}
		}
	}

	// Default assumption for RX 570/580 if we can't detect
	return 8192 // 8GB
}

func (c *Checker) checkROCm() bool {
	// Check if ROCm is installed
	paths := []string{
		"/opt/rocm/bin/rocminfo",
		"/usr/bin/rocminfo",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			// Try running rocminfo
			cmd := exec.Command(path)
			err := cmd.Run()
			return err == nil
		}
	}

	// Check if ROCm modules are loaded
	cmd := exec.Command("lsmod")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "amdgpu")
}

func (c *Checker) getSystemMemory() int {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.Atoi(fields[1])
				if err == nil {
					return kb / 1024 // Convert KB to MB
				}
			}
		}
	}
	return 0
}

func (c *Checker) getKernelVersion() string {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func (c *Checker) validateRequirements(info *SystemInfo) error {
	var errors []string

	if !info.HasDocker {
		errors = append(errors, "Docker is not installed or not accessible")
	}

	if !info.HasNVIDIA && !info.HasAMDGPU {
		errors = append(errors, "No supported GPU detected (NVIDIA or AMD)")
	}

	if info.GPUMemory < 6144 { // 6GB minimum
		errors = append(errors, fmt.Sprintf("GPU memory (%dMB) is below recommended minimum (6GB)", info.GPUMemory))
	}

	if info.SystemMemory < 8192 { // 8GB minimum
		errors = append(errors, fmt.Sprintf("System memory (%dMB) is below recommended minimum (8GB)", info.SystemMemory))
	}

	if info.GPUType == "amd" && !info.HasROCm {
		logrus.Warn("ROCm not detected - AMD GPU acceleration may not work properly")
		logrus.Warn("Install ROCm for optimal performance: https://rocm.docs.amd.com/projects/install-on-linux/en/latest/")
	}
	
	if info.GPUType == "nvidia" {
		// Check for NVIDIA Container Toolkit
		cmd := exec.Command("docker", "run", "--rm", "--gpus", "all", "nvidia/cuda:11.0-base", "nvidia-smi")
		err := cmd.Run()
		if err != nil {
			logrus.Warn("NVIDIA Container Toolkit not detected - GPU acceleration may not work properly")
			logrus.Warn("Install NVIDIA Container Toolkit: https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("system requirements not met:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

func (c *Checker) printSystemInfo(info *SystemInfo) {
	logrus.Info("=== System Information ===")
	logrus.Infof("Kernel Version: %s", info.KernelVersion)
	logrus.Infof("Docker: %v", info.HasDocker)
	logrus.Infof("GPU Type: %s", info.GPUType)
	
	if info.HasNVIDIA {
		logrus.Infof("NVIDIA GPU: %v", info.HasNVIDIA)
		logrus.Infof("GPU Model: %s", info.GPUModel)
		logrus.Infof("GPU Memory: %d MB", info.GPUMemory)
	}
	
	if info.HasAMDGPU {
		logrus.Infof("AMD GPU: %v", info.HasAMDGPU)
		logrus.Infof("GPU Model: %s", info.GPUModel)
		logrus.Infof("GPU Memory: %d MB", info.GPUMemory)
		logrus.Infof("ROCm: %v", info.HasROCm)
	}
	
	logrus.Infof("System Memory: %d MB", info.SystemMemory)
	logrus.Info("=========================")
}