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
}

func streamLogs(podName, searchText string, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command("kubectl", "logs", "-f", podName, "--tail=10")
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
			fmt.Printf("[%s] %s\n", podName, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdout: %v\n", err)
		return
	}

}

func getPodsByLabel(labelSelector string) ([]string, error) {
	cmd := exec.Command("kubectl", "get", "pods", "-l", labelSelector, "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error fetching pods: %v", err)
	}
	podNames := strings.Fields(string(output))
	return podNames, nil
}

func getPodsByDeployment(deploymentName string) ([]string, error) {
	cmd := exec.Command("kubectl", "get", "deployment", deploymentName, "-o", "jsonpath={.spec.selector.matchLabels}")
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

	return getPodsByLabel(labelSelector)
}

func parseArgs(args []string) CommandOptions {
	var opts CommandOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
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
	return opts
}

func handlePods(pods []string, searchText string) {
	var wg sync.WaitGroup
	for _, pod := range pods {
		wg.Add(1)
		go streamLogs(pod, searchText, &wg)
	}
	wg.Wait()
}

func handleDeployments(deployments []string, searchText string) {
	var wg sync.WaitGroup
	for _, deploy := range deployments {
		pods, err := getPodsByDeployment(deploy)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get pods for deployment %s: %v\n", deploy, err)
			continue
		}
		for _, pod := range pods {
			wg.Add(1)
			go streamLogs(pod, searchText, &wg)
		}
	}
	wg.Wait()
}

func main() {
	opts := parseArgs(os.Args[1:])

	switch opts.Mode {
	case "pod", "pod_find":
		handlePods(opts.PodNames, opts.SearchText)
	case "pod_multi", "find_multi":
		handlePods(opts.PodNames, opts.SearchText)
	case "deploy", "deploy_find":
		handleDeployments(opts.DeploymentNames, opts.SearchText)
	case "deploy_multi", "find_multi_deploy":
		handleDeployments(opts.DeploymentNames, opts.SearchText)
	default:
		fmt.Println("Invalid command. Usage:")
		// ... print usage instructions ...
	}
}
