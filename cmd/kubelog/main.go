// main.go (located in cmd/myapp/)
package main

import (
	"fmt"
	"kubelog/pkg/myapp" // Adjust this import path based on your actual module name
	"os"
)

func main() {
	opts := kubelog.ParseArgs(os.Args[1:]) // Use myapp prefix for ParseArgs

	switch opts.Mode {
	case "pod", "pod_find":
		kubelog.HandlePods(opts.PodNames, opts.SearchText, opts.Namespace) // Use myapp prefix for HandlePods
	case "pod_multi":
		kubelog.HandlePods(opts.PodNames, "", opts.Namespace)
	case "find_multi_pods":
		kubelog.HandlePods(opts.PodNames, opts.SearchText, opts.Namespace)
	case "deploy", "deploy_find":
		kubelog.HandleDeployments(opts.DeploymentNames, opts.SearchText, opts.Namespace)
	case "deploy_multi":
		kubelog.HandleDeployments(opts.DeploymentNames, "", opts.Namespace)
	case "find_multi_deploy":
		kubelog.HandleDeployments(opts.DeploymentNames, opts.SearchText, opts.Namespace)
	default:
		fmt.Println("Invalid command. Usage:")
		// ... print usage instructions ...
	}
}
