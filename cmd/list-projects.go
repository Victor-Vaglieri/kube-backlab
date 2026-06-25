package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

func main() {
	fmt.Printf("\n%s=== KUBE-BACKLAB ACTIVE PROJECTS ===%s\n", colorCyan, colorReset)

	// kubectl get deployments -A --no-headers
	cmd := exec.Command("kubectl", "get", "deployments", "-A", "--no-headers")
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeLog("ERROR", fmt.Sprintf("Failed to list deployments: %v", err), colorRed)
		os.Exit(1)
	}

	lines := strings.Split(string(output), "\n")
	var projectsList []string
	seen := make(map[string]bool)

	for _, line := range lines {
		if strings.Contains(line, "hello-app") {
			parts := regexp.MustCompile(`\s+`).Split(strings.TrimSpace(line), -1)
			if len(parts) > 0 {
				ns := parts[0]
				if !seen[ns] {
					projectsList = append(projectsList, ns)
					seen[ns] = true
				}
			}
		}
	}

	if len(projectsList) > 0 {
		writeLog("INFO", fmt.Sprintf("Found %d active project(s):", len(projectsList)), colorWhite)
		fmt.Println("")

		for _, pName := range projectsList {
			if pName == "" {
				continue
			}

			// kubectl get pods -n $pName --no-headers
			podCmd := exec.Command("kubectl", "get", "pods", "-n", pName, "--no-headers")
			podOutput, _ := podCmd.CombinedOutput()
			podLines := strings.Split(strings.TrimSpace(string(podOutput)), "\n")

			runningCount := 0
			totalCount := 0
			if len(podLines) > 0 && podLines[0] != "" {
				totalCount = len(podLines)
				for _, pLine := range podLines {
					if strings.Contains(pLine, "Running") {
						runningCount++
					}
				}
			}

			pHost := pName + ".dev.local"
			if pName == "dev" {
				pHost = "hello.dev.local"
			}

			hColor := colorYellow
			if runningCount == totalCount && totalCount > 0 {
				hColor = colorGreen
			}

			fmt.Printf(" Project: ")
			fmt.Printf("%s%-15s%s", colorCyan, pName, colorReset)
			fmt.Printf("%s[%d/%d Pods] %s", hColor, runningCount, totalCount, colorReset)
			fmt.Printf("-> %shttp://%s:8080%s\n", colorGray, pHost, colorReset)
		}
	} else {
		writeLog("WARN", "No active projects found in the cluster.", colorYellow)
	}

	fmt.Printf("\n%sTip: Use 'go run start-lab.go -Project <name>' to create a new one.%s\n\n", colorGray, colorReset)
}
