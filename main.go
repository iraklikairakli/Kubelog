package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type CommandOptions struct {
	Mode            string
	SearchText      string
	PodNames        []string
	DeploymentNames []string
	Namespace       string
}

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

func parseArgs(args []string) CommandOptions {
	var opts CommandOptions

	opts.Namespace = "default" // Default namespace

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-n":
			if i+1 < len(args) {
				opts.Namespace = args[i+1]
				i++
			}
		case "-p":
			opts.Mode = "pod"
			opts.PodNames = append(opts.PodNames, args[i+1])
			i++
		case "-pf":
			opts.Mode = "pod_find"
			opts.PodNames = append(opts.PodNames, args[i+1])
			opts.SearchText = args[i+2]
			i += 2
		case "-pm":
			opts.Mode = "pod_multi"
			for j := i + 1; j < len(args) && !strings.HasPrefix(args[j], "-"); j++ {
				opts.PodNames = append(opts.PodNames, args[j])
				i = j
			}
		case "-fmp":
			opts.Mode = "find_multi_pods"
			i++
			// Loop to add pod names. Stop one argument before the end to ensure the last argument is treated as search text.
			for ; i < len(args)-1 && !strings.HasPrefix(args[i], "-"); i++ {
				opts.PodNames = append(opts.PodNames, args[i])
			}
			// The last argument is the search text
			if i < len(args) {
				opts.SearchText = args[i]
			}

		case "-fm":
			opts.Mode = "find_multi"
			for j := i + 1; j < len(args) && args[j] != "-f"; j++ {
				opts.PodNames = append(opts.PodNames, args[j])
				i = j
			}
			if i+1 < len(args) {
				opts.SearchText = args[i+1]
			}
		case "-d":
			opts.Mode = "deploy"
			opts.DeploymentNames = append(opts.DeploymentNames, args[i+1])
			i++
		case "-df":
			opts.Mode = "deploy_find"
			opts.DeploymentNames = append(opts.DeploymentNames, args[i+1])
			opts.SearchText = args[i+2]
			i += 2
		case "-md":
			opts.Mode = "deploy_multi"
			for j := i + 1; j < len(args) && !strings.HasPrefix(args[j], "-"); j++ {
				opts.DeploymentNames = append(opts.DeploymentNames, args[j])
				i = j
			}
		case "-fmd":
			opts.Mode = "find_multi_deploy"
			for j := i + 1; j < len(args) && args[j] != "-f"; j++ {
				opts.DeploymentNames = append(opts.DeploymentNames, args[j])
				i = j
			}
			if i+1 < len(args) {
				opts.SearchText = args[i+1]
			}
		}
	}
	fmt.Println("Final parsed options:", opts) // Debugging print
	return opts
}

func handlePods(pods []string, searchText, namespace string) {
	var wg sync.WaitGroup
	podColorMap := make(map[string]string)
	for _, pod := range pods {
		wg.Add(1)
		go streamLogs(pod, searchText, namespace, &wg, podColorMap)
	}
	wg.Wait()
}

func handleDeployments(deployments []string, searchText, namespace string) {
	var wg sync.WaitGroup
	podColorMap := make(map[string]string) // Added to support color mapping
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

func main() {
	opts := parseArgs(os.Args[1:])

	switch opts.Mode {
	case "pod", "pod_find":
		handlePods(opts.PodNames, opts.SearchText, opts.Namespace)
	case "pod_multi":
		handlePods(opts.PodNames, "", opts.Namespace)
	case "find_multi_pods":
		handlePods(opts.PodNames, opts.SearchText, opts.Namespace)
	case "deploy", "deploy_find":
		handleDeployments(opts.DeploymentNames, opts.SearchText, opts.Namespace)
	case "deploy_multi":
		handleDeployments(opts.DeploymentNames, "", opts.Namespace)
	case "find_multi_deploy":
		handleDeployments(opts.DeploymentNames, opts.SearchText, opts.Namespace)
	default:
		fmt.Println("Invalid command. Usage:")
		// ... print usage instructions ...
	}
}
