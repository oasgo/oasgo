package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var spec, packageName, destination string
var isAbbreviate bool

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
		if packageName == "" {
			packageName = "client"
		}
		renderClient(s, packageName, destination, isAbbreviate)
	},
}

var dtoCmd = &cobra.Command{
	Use:   "dto",
	Short: "generates DTO structs",
	Run: func(cmd *cobra.Command, args []string) {
		s := parse(spec)
		if packageName == "" {
			packageName = "dto"
		}
		renderDTO(s, packageName, destination, isAbbreviate)
	},
}

func main() {
	var rootCmd = &cobra.Command{}
	rootCmd.AddCommand(parseCmd, genCmd)
	rootCmd.PersistentFlags().StringVarP(&spec, "file", "f", "", "path to swagger spec")
	genCmd.AddCommand(clientCmd, dtoCmd)
	genCmd.PersistentFlags().StringVarP(&packageName, "package_name", "n", "", "name for generated package")
	genCmd.PersistentFlags().StringVarP(&destination, "destination", "d", "", "destination for generated package")
	genCmd.PersistentFlags().BoolVarP(&isAbbreviate, "abbreviate", "a", false, "abbreviate the names of generated structures")
	rootCmd.Execute()
}

// getRefName returns last element from ref string e.g.: "#/components/schemas/Pet"
func getRefName(ref string) string {
	path := strings.Split(ref, "/")
	return path[len(path)-1]
}
