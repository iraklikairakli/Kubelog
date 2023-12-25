// parser.go (located in pkg/myapp/)
package myapp

import (
	"strings"
)

type CommandOptions struct {
	Mode            string
	SearchText      string
	PodNames        []string
	DeploymentNames []string
	Namespace       string
}

func ParseArgs(args []string) CommandOptions {
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
	return opts
}
