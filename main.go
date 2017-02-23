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

package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/cobra/cmd"
)

func main() {
	cmd.Execute()

	var cmdPrint = &cobra.Command{
		Use:   "print [string to print]",
		Short: "Print anything to the screen",
		Long: `print is for printing anything back to the screen.
            For many years people have printed back to the screen.
            `,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Print: " + strings.Join(args, " "))
		},
	}

	var cmdEcho = &cobra.Command{
		Use:   "echo [string to echo]",
		Short: "Echo anything to the screen",
		Long: `echo is for echoing anything back.
            Echo works a lot like print, except it has a child command.
            `,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Print: " + strings.Join(args, " "))
		},
	}

	var cmdTimes = &cobra.Command{
		Use:   "times [# times] [string to echo]",
		Short: "Echo anything to the screen more times",
		Long: `echo things multiple times back to the user by providing
            a count and a string.`,
		Run: func(cmd *cobra.Command, args []string) {
			for i := 0; i < echoTimes; i++ {
				fmt.Println("Echo: " + strings.Join(args, " "))
			}
		},
	}

	cmdTimes.Flags().IntVarP(&echoTimes, "times", "t", 1, "times to echo the input")

	var rootCmd = &cobra.Command{Use: "app"}
	rootCmd.AddCommand(cmdPrint, cmdEcho)
	cmdEcho.AddCommand(cmdTimes)
	rootCmd.Execute()
}
