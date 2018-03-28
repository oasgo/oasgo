package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var spec string
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "parse openapi spec and prints swagger. Need for debug stuff",
	Run: func(cmd *cobra.Command, args []string) {
		s := parse(spec)
		fmt.Println(s)
	},
}
var genCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate golang file and print it to the output",
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "generate client golang file and print it to the output",
	Run: func(cmd *cobra.Command, args []string) {
		s := parse(spec)
		render(s, "client.tmpl")
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "generate server golang file and print it to the output",
	Run: func(cmd *cobra.Command, args []string) {
		s := parse(spec)
		render(s, "server.tmpl")
	},
}

func main() {
	var rootCmd = &cobra.Command{}
	rootCmd.AddCommand(parseCmd, genCmd)
	rootCmd.PersistentFlags().StringVarP(&spec, "file", "f", "", "path to swagger spec")
	genCmd.AddCommand(clientCmd, serverCmd)
	rootCmd.Execute()
}

// getRefName returns last element from ref string e.g.: "#/components/schemas/Pet"
func getRefName(ref string) string {
	path := strings.Split(ref, "/")
	return path[len(path)-1]
}
