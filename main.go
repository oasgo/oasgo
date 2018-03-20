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
	Run: func(cmd *cobra.Command, args []string) {
		s := parse(spec)
		render(s)
	},
}

func main() {
	var rootCmd = &cobra.Command{Use: "fen"}
	rootCmd.AddCommand(parseCmd, genCmd)
	rootCmd.PersistentFlags().StringVarP(&spec, "file", "f", "", "path to swagger spec")
	rootCmd.Execute()
}

// getRefName returns last element from ref string e.g.: "#/components/schemas/Pet"
func getRefName(ref string) string {
	path := strings.Split(ref, "/")
	return path[len(path)-1]
}
