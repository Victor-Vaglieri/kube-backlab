package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	project := flag.String("Project", "dev", "Project name")
	watch := flag.Bool("Watch", false, "Watch logs")

	flag.Parse()

	namespace := strings.ToLower(*project)

	fmt.Printf("\n%s=== KUBE-BACKLAB DIAGNOSTIC LOGS ===%s\n", colorCyan, colorReset)
	fmt.Printf("%sTarget Project: %s%s\n", colorGray, namespace, colorReset)

	// kubectl get pods -A --no-headers
	cmd := exec.Command("kubectl", "get", "pods", "-A", "--no-headers")
	output, _ := cmd.CombinedOutput()
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	var failingPods []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "Running") && !strings.Contains(line, "Completed") {
			failingPods = append(failingPods, line)
		}
	}

	if len(failingPods) > 0 {
		writeLog("ERROR", "Found failing pods in the cluster!", colorRed)
		for _, pod := range failingPods {
			fmt.Printf(" > %s%s%s\n", colorRed, pod, colorReset)
		}

		fmt.Printf("\n%s--- [ LAST ERROR LOGS ] ---%s\n", colorYellow, colorReset)
		for _, podLine := range failingPods {
			parts := strings.Fields(podLine)
			if len(parts) >= 2 {
				ns := parts[0]
				name := parts[1]
				fmt.Printf("%s[%s/%s] Recent Logs:%s\n", colorGray, ns, name, colorReset)
				logCmd := exec.Command("kubectl", "logs", name, "-n", ns, "--tail=10")
				logOutput, _ := logCmd.CombinedOutput()
				fmt.Printf("%s\n", string(logOutput))
				fmt.Printf("%s------------------------%s\n", colorGray, colorReset)
			}
		}
	} else {
		writeLog("SUCCESS", "All pods are healthy (Running/Completed).", colorGreen)
	}

	fmt.Printf("\n%s--- [ APPLICATION LOGS (%s) ] ---%s\n", colorCyan, namespace, colorReset)
	appPodCmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "-l", "app=hello-app", "-o", "name")
	appPodOutput, _ := appPodCmd.Output()
	appPod := strings.TrimSpace(string(appPodOutput))
	if appPod != "" {
		// handle multiple pods if returned, take first
		pods := strings.Split(appPod, "\n")
		targetPod := strings.TrimPrefix(pods[0], "pod/")
		
		var logCmd *exec.Cmd
		if *watch {
			writeLog("INFO", "Streaming application logs... (Ctrl+C to stop)", colorWhite)
			logCmd = exec.Command("kubectl", "logs", "-n", namespace, targetPod, "-f")
		} else {
			logCmd = exec.Command("kubectl", "logs", "-n", namespace, targetPod, "--tail=20")
		}
		logCmd.Stdout = os.Stdout
		logCmd.Stderr = os.Stderr
		logCmd.Run()
	} else {
		writeLog("WARN", fmt.Sprintf("No application pod found in '%s' namespace.", namespace), colorYellow)
	}

	fmt.Printf("\n%s--- [ INFRASTRUCTURE LOGS (infra) ] ---%s\n", colorCyan, colorReset)
	dbPodCmd := exec.Command("kubectl", "get", "pods", "-n", "infra", "-l", "app.kubernetes.io/name=postgresql", "-o", "name")
	dbPodOutput, _ := dbPodCmd.Output()
	dbPod := strings.TrimSpace(string(dbPodOutput))
	if dbPod != "" {
		pods := strings.Split(dbPod, "\n")
		targetPod := strings.TrimPrefix(pods[0], "pod/")
		writeLog("INFO", "Last 5 DB logs:", colorWhite)
		logCmd := exec.Command("kubectl", "logs", "-n", "infra", targetPod, "--tail=5")
		logOutput, _ := logCmd.CombinedOutput()
		fmt.Printf("%s\n", string(logOutput))
	} else {
		writeLog("WARN", "No DB pod found in 'infra' namespace.", colorYellow)
	}

	fmt.Printf("\n%sTip: Use 'k9s' for a better interactive experience.%s\n\n", colorGray, colorReset)
}
