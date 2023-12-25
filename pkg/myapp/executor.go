// executor.go (located in pkg/myapp/)
package myapp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var colors = []string{
	"\033[31m", // Red
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[34m", // Blue
	"\033[35m", // Magenta
	"\033[36m", // Cyan
	"\033[37m", // White
}

func getColorCode(podName string, podColorMap map[string]string) string {
	if color, exists := podColorMap[podName]; exists {
		return color
	}
	color := colors[len(podColorMap)%len(colors)]
	podColorMap[podName] = color
	return color
}

func streamLogs(podName, searchText, namespace string, wg *sync.WaitGroup, podColorMap map[string]string) {
	defer wg.Done()
	colorCode := getColorCode(podName, podColorMap)

	cmd := exec.Command("kubectl", "logs", "-f", podName, "-n", namespace, "--tail=10")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating StdoutPipe for Cmd: %v\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Cmd: %v\n", err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if searchText == "" || strings.Contains(line, searchText) {
			fmt.Printf("%s[%s]%s %s\n", colorCode, podName, "\033[0m", line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdout: %v\n", err)
		return
	}
}

func getPodsByLabel(labelSelector, namespace string) ([]string, error) {
	cmd := exec.Command("kubectl", "get", "pods", "-l", labelSelector, "-n", namespace, "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error fetching pods: %v", err)
	}
	podNames := strings.Fields(string(output))
	return podNames, nil
}

func getPodsByDeployment(deploymentName, namespace string) ([]string, error) {
	cmd := exec.Command("kubectl", "get", "deployment", deploymentName, "-n", namespace, "-o", "jsonpath={.spec.selector.matchLabels}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error fetching deployment selector: %v", err)
	}

	var labels map[string]string
	if err := json.Unmarshal(output, &labels); err != nil {
		return nil, fmt.Errorf("error parsing deployment selector: %v", err)
	}

	var selectorParts []string
	for key, value := range labels {
		selectorParts = append(selectorParts, fmt.Sprintf("%s=%s", key, value))
	}
	labelSelector := strings.Join(selectorParts, ",")

	return getPodsByLabel(labelSelector, namespace)
}

func HandlePods(pods []string, searchText, namespace string) {
	var wg sync.WaitGroup
	podColorMap := make(map[string]string)
	for _, pod := range pods {
		wg.Add(1)
		go streamLogs(pod, searchText, namespace, &wg, podColorMap)
	}
	wg.Wait()
}

func HandleDeployments(deployments []string, searchText, namespace string) {
	var wg sync.WaitGroup
	podColorMap := make(map[string]string)
	for _, deploy := range deployments {
		pods, err := getPodsByDeployment(deploy, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get pods for deployment %s: %v\n", deploy, err)
			continue
		}
		for _, pod := range pods {
			wg.Add(1)
			go streamLogs(pod, searchText, namespace, &wg, podColorMap)
		}
	}
	wg.Wait()
}
