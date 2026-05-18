package main

import (
	"flag"
	"fmt"
	"math/rand"
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
	flag.Parse()

	namespace := strings.ToLower(*project)

	fmt.Printf("\n%s=== KUBE-BACKLAB FAILURE SIMULATOR ===%s\n", colorRed, colorReset)
	fmt.Printf("%sTarget Namespace: %s%s\n", colorGray, namespace, colorReset)

	// kubectl get pods -n $Namespace -o name
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "-o", "name")
	output, _ := cmd.CombinedOutput()
	pods := strings.Split(strings.TrimSpace(string(output)), "\n")

	if len(pods) == 0 || (len(pods) == 1 && pods[0] == "") {
		writeLog("ERROR", fmt.Sprintf("No pods found in '%s' namespace.", namespace), colorRed)
		os.Exit(1)
	}

	// Seed random
	rand.Seed(time.Now().UnixNano())
	podToKill := pods[rand.Intn(len(pods))]

	writeLog("ACTION", fmt.Sprintf("Simulating failure by killing pod: %s", podToKill), colorYellow)

	// kubectl delete $podToKill -n $Namespace --grace-period=0 --force
	delCmd := exec.Command("kubectl", "delete", podToKill, "-n", namespace, "--grace-period=0", "--force")
	delCmd.Run()

	writeLog("INFO", "Waiting for Kubernetes to self-heal...", colorWhite)
	time.Sleep(2 * time.Second)

	fmt.Printf("\n%s--- [ HEALING STATUS ] ---%s\n", colorCyan, colorReset)
	statusCmd := exec.Command("kubectl", "get", "pods", "-n", namespace)
	statusCmd.Stdout = os.Stdout
	statusCmd.Stderr = os.Stderr
	statusCmd.Run()

	writeLog("SUCCESS", "Kubernetes is recreating the pod automatically!", colorGreen)
}
