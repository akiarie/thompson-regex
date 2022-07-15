package cmd

import (
	"fmt"
	"log"
	"os"

	"thompson-regex/compiler"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "thompson-regex [expression]",
	Short: "A regular-expression compiler that generates Go code",
	Long: `A simple compiler that takes a simple regular expression and outputs Go code
which, when compiled, produces a program which takes a string as its input and
outputs all the matches of the regular expression.  

It is based on Thompson's construction algorithm and is meant for educational
purposes only.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requires a regex argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		sievedexp, err := compiler.Sieve(args[0])
		if err != nil {
			log.Fatalln("cannot sive:", err)
		}
		rpnexp, err := compiler.RPNConvert(sievedexp)
		if err != nil {
			log.Fatalln("cannot convert to RPN:", err)
		}
		matcher, err := compiler.Compile(rpnexp)
		if err != nil {
			log.Fatalln("cannot produce Go matcher code:", err)
		}
		fmt.Println(matcher)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
