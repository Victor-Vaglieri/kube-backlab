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
	full := flag.Bool("Full", false, "Stop the whole cluster")

	flag.Parse()

	namespace := strings.ToLower(*project)

	fmt.Printf("\n%s=== KUBE-BACKLAB SHUTDOWN ===%s\n", colorYellow, colorReset)

	writeLog("INFO", fmt.Sprintf("Removing application deployments for project '%s'...", namespace), colorWhite)
	
	// skaffold delete --namespace $Namespace
	cmd := exec.Command("skaffold", "delete", "--namespace", namespace)
	cmd.Env = append(os.Environ(), "PROJECT_PATH=src")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	if *full {
		writeLog("INFO", "Stopping k3d cluster 'dev-cluster'...", colorWhite)
		stopCmd := exec.Command("k3d", "cluster", "stop", "dev-cluster")
		stopCmd.Stdout = os.Stdout
		stopCmd.Stderr = os.Stderr
		stopCmd.Run()
		writeLog("SUCCESS", fmt.Sprintf("Project '%s' removed and Cluster stopped.", namespace), colorGreen)
	} else {
		writeLog("SUCCESS", fmt.Sprintf("Project '%s' removed. Cluster is still running for other projects.", namespace), colorGreen)
		fmt.Printf("%sTIP: To stop the whole cluster, use: go run stop-lab.go -Full%s\n", colorGray, colorReset)
	}

	fmt.Printf("%sTo restart, run: go run start-lab.go -Project %s%s\n\n", colorGray, namespace, colorReset)
}
