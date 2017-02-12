package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	_ "github.com/flimzy/kivik/driver/memory"
	"github.com/flimzy/kivik/serve"
)

func main() {
	pflag.BoolP("verbose", "v", false, "Verbose output")
	flagVerbose := pflag.Lookup("verbose")
	pflag.BoolP("quiet", "q", false, "Supress non-fatal warnings")
	flagQuiet := pflag.Lookup("quiet")

	cmdServe := &cobra.Command{
		Use:   "serve [options]",
		Short: "serve",
	}
	cmdServe.Flags().AddFlag(flagVerbose)
	cmdServe.Flags().AddFlag(flagQuiet)
	var addr string
	cmdServe.Flags().StringVarP(&addr, "http", "", ":5984", "HTTP bind address to serve")
	var driver string
	cmdServe.Flags().StringVarP(&driver, "driver", "d", "memory", "Backend driver to use")
	var dsn string
	cmdServe.Flags().StringVarP(&dsn, "dsn", "", "", "Data source name")
	cmdServe.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("Serving on %s\n", addr)
		service := serve.New(driver, dsn)
		err := service.Start(addr)
		fmt.Println(err)
		os.Exit(1)
	}

	cmdTest := &cobra.Command{
		Use:   "test [options] [server]",
		Short: "Run the test suite against the remote server",
	}
	cmdTest.Flags().AddFlag(flagVerbose)
	cmdTest.Flags().AddFlag(flagQuiet)
	cmdTest.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("Test [%s]\n", args)
	}

	rootCmd := &cobra.Command{
		Use:  "kivik",
		Long: "Kivik is a tool for hosting and testing CouchDB services",
	}
	rootCmd.AddCommand(cmdServe, cmdTest)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(2)
	}
}
