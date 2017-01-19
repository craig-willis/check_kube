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
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"

	client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/spf13/cobra"
)

// podsCmd represents the podsCmd command
var podsCmd = &cobra.Command{
	Use:   "pods [warning] [critical]",
	Short: "Check pod statuses",

	Run: func(cmd *cobra.Command, args []string) {

		var err error
		var statusCode = nagiosStatusOK
		var statusLine []string
		var warningCount int
		var criticalCount int

		if len(args) < 2 {
			cmd.Usage()
			os.Exit(-1)
		}

		warningLevel, _ := strconv.Atoi(args[0])
		criticalLevel, _ := strconv.Atoi(args[1])

		config := &restclient.Config{}
		config.BearerToken = token
		config.Host = server
		kubeClient, err := client.New(config)
		if err != nil {
			fmt.Printf("CRITICAL: %s\n", err)
			os.Exit(nagiosStatusUnknown)
		}

		pods, err := kubeClient.Pods("").List(
			api.ListOptions{
				LabelSelector: labels.Everything(),
				FieldSelector: fields.Everything(),
			},
		)
		if err != nil {
			fmt.Printf("CRITICAL: %s\n", err)
			os.Exit(nagiosStatusUnknown)
		}

		// Loop over all the pods
		for _, pod := range pods.Items {
			/*
				for _, cond := range pod.Status.Conditions {
					if cond.Type == "Ready" && cond.Status != "True" {
						notReadyCount++
					}
				}
			*/
			for _, status := range pod.Status.ContainerStatuses {
				if status.RestartCount > int32(warningLevel) {
					warningCount++
				}

				if status.RestartCount > int32(criticalLevel) {
					criticalCount++
				}
			}
		}

		if criticalCount != 0 {
			msg := fmt.Sprintf("%d pods exceeding CRITICAL restart threshold.", criticalCount)
			statusLine = append(statusLine, msg)
			statusCode = nagiosStatusCritical
		} else if warningCount != 0 {
			msg := fmt.Sprintf("%d pods exceeding WARNING restart threshold.", warningCount)
			statusLine = append(statusLine, msg)
			statusCode = nagiosStatusWarning
		}

		if statusCode != nagiosStatusOK {
			fmt.Println(strings.Join(statusLine, "\n"))
			os.Exit(statusCode)
		}

		fmt.Println("OK")
		os.Exit(nagiosStatusOK)
	},
}

func init() {
	RootCmd.AddCommand(podsCmd)
}
