// main.go (located in cmd/myapp/)
package main

import (
	"fmt"
	"myapp/pkg/myapp" // Adjust this import path based on your actual module name
	"os"
)

func main() {
	opts := myapp.ParseArgs(os.Args[1:]) // Use myapp prefix for ParseArgs

	switch opts.Mode {
	case "pod", "pod_find":
		myapp.HandlePods(opts.PodNames, opts.SearchText, opts.Namespace) // Use myapp prefix for HandlePods
	case "pod_multi":
		myapp.HandlePods(opts.PodNames, "", opts.Namespace)
	case "find_multi_pods":
		myapp.HandlePods(opts.PodNames, opts.SearchText, opts.Namespace)
	case "deploy", "deploy_find":
		myapp.HandleDeployments(opts.DeploymentNames, opts.SearchText, opts.Namespace)
	case "deploy_multi":
		myapp.HandleDeployments(opts.DeploymentNames, "", opts.Namespace)
	case "find_multi_deploy":
		myapp.HandleDeployments(opts.DeploymentNames, opts.SearchText, opts.Namespace)
	default:
		fmt.Println("Invalid command. Usage:")
		// ... print usage instructions ...
	}
}
