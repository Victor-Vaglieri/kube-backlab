package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorWhite  = "\033[37m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorReset  = "\033[0m"
)

func writeLog(level, message, color string) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("%s[%s] [%s] %s%s\n", color, timestamp, level, message, colorReset)
}

func runInfrastructureChecks(hostHeader, endpoint string) bool {
	fmt.Printf("\n%s--- [ INFRASTRUCTURE VALIDATION ] ---%s\n", colorCyan, colorReset)

	// 1. Gateway Check
	conn, err := net.DialTimeout("tcp", "localhost:8080", 2*time.Second)
	if err != nil {
		writeLog("ERROR", "Gateway is down. Ensure k3d/skaffold is running.", colorRed)
		return false
	}
	writeLog("SUCCESS", "Gateway (localhost:8080) is reachable.", colorGreen)
	conn.Close()

	// 2. Ingress Routing Check
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Host = hostHeader

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		writeLog("ERROR", fmt.Sprintf("Ingress routing failed: %v", err), colorRed)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 302 || resp.StatusCode == 308 {
		writeLog("SUCCESS", fmt.Sprintf("Ingress routing for %s is functional (Response: %s).", hostHeader, resp.Status), colorGreen)
	} else {
		writeLog("ERROR", fmt.Sprintf("Ingress routing failed. Expected 200/302/308, but received: %s", resp.Status), colorRed)
		return false
	}

	return true
}

func runProjectTests(projectPath string) {
	fmt.Printf("\n%s--- [ PROJECT TEST EXECUTION ] ---%s\n", colorCyan, colorReset)

	testPaths := []string{
		filepath.Join(projectPath, "tests"),
		filepath.Join(projectPath, "src/test"),
		filepath.Join(projectPath, "src/tests"),
	}

	srcPath := filepath.Join(projectPath, "src")

	// 1. Run npm test if package.json exists
	if _, err := os.Stat(filepath.Join(srcPath, "package.json")); err == nil {
		writeLog("INFO", "Node.js project detected. Running npm test...", colorWhite)
		cmd := exec.Command("npm", "test", "--prefix", srcPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	foundCustom := false
	for _, path := range testPaths {
		if _, err := os.Stat(path); err == nil {
			files, _ := os.ReadDir(path)
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".ps1") {
					foundCustom = true
					fullPath := filepath.Join(path, file.Name())
					writeLog("INFO", fmt.Sprintf("Running custom test: %s", file.Name()), colorWhite)
					cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", fullPath)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Run()
				}
			}
		}
	}

	if !foundCustom {
		if _, err := os.Stat(filepath.Join(srcPath, "package.json")); err != nil {
			writeLog("WARN", "No tests found in standard locations.", colorYellow)
		}
	}
}

func main() {
	projectPath := flag.String("ProjectPath", ".", "Path to the project")
	project := flag.String("Project", "", "Project name")
	hostHeader := flag.String("HostHeader", "hello.dev.local", "Host header for ingress check")
	endpoint := flag.String("Endpoint", "http://localhost:8080", "Endpoint to check")

	flag.Parse()

	if *project != "" {
		p := strings.ToLower(*project)
		os.Setenv("PROJECT_NAME", p) // Garante que scripts filhos vejam o projeto correto
		if p == "dev" {
			*hostHeader = "hello.dev.local"
		} else {
			*hostHeader = p + ".dev.local"
		}
	}

	if runInfrastructureChecks(*hostHeader, *endpoint) {
		runProjectTests(*projectPath)
	}

	fmt.Printf("\n%s--- [ PROCESS COMPLETE ] ---%s\n\n", colorCyan, colorReset)
}
