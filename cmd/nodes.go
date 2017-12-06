// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"

	client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/spf13/cobra"
)

const (
	// Nagios status codes
	nagiosStatusOK       = 0
	nagiosStatusWarning  = 1
	nagiosStatusCritical = 2
	nagiosStatusUnknown  = 3
)

// nodesCmd represents the nodes command
var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Check node statues",

	Run: func(cmd *cobra.Command, args []string) {

		var err error
		var statusCode = nagiosStatusOK
		var statusLine []string

		config := &restclient.Config{}
		config.BearerToken = token
		config.Host = server
		kubeClient, err := client.New(config)
		if err != nil {
			fmt.Printf("2 check_kube-nodes CRITICAL - %s\n", err)
			os.Exit(nagiosStatusUnknown)
		}

		nodes, err := kubeClient.Nodes().List(
			api.ListOptions{
				LabelSelector: labels.Everything(),
				FieldSelector: fields.Everything(),
			},
		)
		if err != nil {
			fmt.Printf("2 chek_kube-nodes CRITICAL - %s\n", err)
			os.Exit(nagiosStatusUnknown)
		}

		// Loop over all the nodes
		for _, node := range nodes.Items {

			// Loop over all the node conditions
			for _, condition := range node.Status.Conditions {

				// Check the NodeReady condition
				if condition.Type == api.NodeReady && condition.Status != api.ConditionTrue {
					msg := fmt.Sprintf("%s, %s, %s", node.Name, condition.Reason, condition.Message)
					statusLine = append(statusLine, msg)
					statusCode = nagiosStatusCritical
				}
			}
		}

		if statusCode != nagiosStatusOK {
			fmt.Printf("%d check_kube-nodes nodes=%d %s\n", statusCode, len(nodes.Items), strings.Join(statusLine, "\n"))
			os.Exit(statusCode)
		}

		fmt.Printf("0 check_kube-nodes nodes=%d OK\n", len(nodes.Items))
		os.Exit(nagiosStatusOK)
	},
}

func init() {
	RootCmd.AddCommand(nodesCmd)
}
