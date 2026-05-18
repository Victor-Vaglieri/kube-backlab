package main

import (
	"flag"
	"fmt"
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

func writeBox(title string, lines []string, color string) {
	width := 60
	fmt.Printf("%s+%s+%s\n", color, strings.Repeat("-", width-2), colorReset)
	fmt.Printf("%s| %-56s |%s\n", color, title, colorReset)
	fmt.Printf("%s+%s+%s\n", color, strings.Repeat("-", width-2), colorReset)
	for _, line := range lines {
		fmt.Printf("%s| %-56s |%s\n", color, line, colorReset)
	}
	fmt.Printf("%s+%s+%s\n", color, strings.Repeat("-", width-2), colorReset)
}

func main() {
	pathFlag := flag.String("Path", "src", "Path to project source")
	projectFlag := flag.String("Project", "dev", "Project name")
	flag.Parse()

	absPath, _ := filepath.Abs(*pathFlag)
	projectName := strings.ToLower(*projectFlag)

	fmt.Printf("\n%s=== KUBE-BACKLAB MASTER SETUP ===%s\n", colorCyan, colorReset)
	fmt.Printf("%sTarget Project: %s%s\n", colorGray, absPath, colorReset)
	fmt.Printf("%sNamespace:      %s%s\n", colorGray, projectName, colorReset)

	// 1. Check Cluster
	writeLog("INFO", "Checking k3d cluster status...", colorWhite)
	clusterCmd := exec.Command("k3d", "cluster", "list", "dev-cluster", "--no-headers")
	clusterOutput, _ := clusterCmd.Output()
	clusterStatus := string(clusterOutput)

	if clusterStatus == "" {
		writeLog("ACTION", "Creating dev-cluster...", colorYellow)
		exec.Command("k3d", "cluster", "create", "dev-cluster", "-p", "8080:80@loadbalancer").Run()
	} else if strings.Contains(clusterStatus, "0/1") {
		writeLog("ACTION", "Cluster exists but is stopped. Starting dev-cluster...", colorYellow)
		exec.Command("k3d", "cluster", "start", "dev-cluster").Run()
	} else {
		writeLog("SUCCESS", "Cluster dev-cluster is running and ready.", colorGreen)
	}

	// 2. Setup Namespaces
	writeLog("INFO", "Ensuring namespaces exist...", colorWhite)
	namespaces := []string{projectName, "infra", "monitoring"}
	for _, ns := range namespaces {
		nsCmd := fmt.Sprintf("kubectl create namespace %s --dry-run=client -o yaml | kubectl apply -f -", ns)
		exec.Command("powershell", "-Command", nsCmd).Run()
	}

	// 3. Apply Project Specific Configs
	projectConfig := filepath.Join(absPath, "k8s-config.yaml")
	if _, err := os.Stat(projectConfig); err == nil {
		writeLog("SUCCESS", fmt.Sprintf("Detected k8s-config.yaml in project. Applying to namespace %s...", projectName), colorGreen)
		exec.Command("kubectl", "apply", "-f", projectConfig, "-n", projectName).Run()
	} else {
		writeLog("INFO", "Using root k8s manifests via Skaffold.", colorWhite)
	}

	// 4. Infrastructure Status
	fmt.Printf("\n%s--- [ INFRASTRUCTURE STATUS ] ---%s\n", colorCyan, colorReset)
	podCmd := exec.Command("kubectl", "get", "pods", "-A", "--no-headers")
	podOutput, _ := podCmd.Output()
	podLines := strings.Split(strings.TrimSpace(string(podOutput)), "\n")
	running := 0
	total := 0
	for _, line := range podLines {
		if line == "" { continue }
		total++
		if strings.Contains(line, "Running") {
			running++
		}
	}
	healthColor := colorYellow
	if running == total && total > 0 {
		healthColor = colorGreen
	}
	writeLog("INFO", fmt.Sprintf("Cluster Health: %d/%d pods running.", running, total), healthColor)

	// 5. Start Application
	writeLog("INFO", fmt.Sprintf("Ensuring application is deployed to namespace '%s'...", projectName), colorWhite)
	skaffoldCmd := exec.Command("skaffold", "run", "--status-check=false", "--namespace", projectName)
	skaffoldCmd.Env = append(os.Environ(), "PROJECT_PATH="+absPath)
	skaffoldCmd.Stdout = os.Stdout
	skaffoldCmd.Stderr = os.Stderr
	skaffoldCmd.Run()

	// 6. Apply Dynamic Routing
	appHost := projectName + ".dev.local"
	if projectName == "dev" {
		appHost = "hello.dev.local"
	}
	writeLog("INFO", fmt.Sprintf("Configuring dynamic routing for %s...", appHost), colorWhite)
	patch1 := fmt.Sprintf("[{'op': 'replace', 'path': '/spec/rules/0/host', 'value':'%s'}]", appHost)
	exec.Command("kubectl", "patch", "ingress", "hello-ingress", "-n", projectName, "--type=json", "-p", patch1).Run()
	patch2 := fmt.Sprintf("[{'op': 'replace', 'path': '/spec/tls/0/hosts/0', 'value':'%s'}]", appHost)
	exec.Command("kubectl", "patch", "ingress", "hello-ingress", "-n", projectName, "--type=json", "-p", patch2).Run()

	// 7. Access Information
	fmt.Println("")
	appUrl := fmt.Sprintf("http://%s:8080", appHost)
	writeBox("ACCESS ENDPOINTS (Namespace: "+projectName+")", []string{
		"Application: " + appUrl,
		"Grafana:     http://grafana.dev.local:8080",
		"Prometheus:  http://prometheus.dev.local:8080",
	}, colorCyan)

	fmt.Printf("%sCRITICAL STEP REQUIRED:%s\n", colorRed, colorReset)
	fmt.Printf("%sTo access the application, you MUST add this entry to your Windows 'hosts' file:%s\n", colorYellow, colorReset)
	fmt.Printf("   127.0.0.1 %s\n", appHost)
	fmt.Printf("%sLocation: C:\\Windows\\System32\\drivers\\etc\\hosts%s\n", colorGray, colorReset)

	writeBox("DATABASE INFO (PostgreSQL)", []string{
		"Host:     postgres-postgresql.infra.svc.cluster.local",
		"Port:     5432",
		"User:     postgres",
		"DB Name:  postgres",
		"Password: [Check secret.yaml or env]",
	}, colorCyan)

	writeBox("DEBUGGING & LOGS", []string{
		"App Logs:   go run debug-logs.go -Project " + projectName + " -f",
		"Failure Sim: go run simulate-failure.go -Project " + projectName,
		"Stop Lab:    go run stop-lab.go -Project " + projectName,
		"Live View:   k9s -n " + projectName + " (Recommended)",
		"Test Lab:    go run check-lab.go -Project " + projectName,
	}, colorCyan)

	fmt.Printf("%sTIP: To work on this project with hot-reload, run:%s\n", colorYellow, colorReset)
	fmt.Printf("   skaffold dev --namespace %s\n", projectName)
	fmt.Printf("\n%s=== SETUP COMPLETED SUCCESSFULLY ===%s\n\n", colorGreen, colorReset)
}
