package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/memory"
	"github.com/flimzy/kivik/serve"
	"github.com/flimzy/kivik/test"
)

func main() {
	pflag.BoolP("verbose", "v", false, "Verbose output")
	flagVerbose := pflag.Lookup("verbose")
	pflag.BoolP("quiet", "q", false, "Supress non-fatal warnings")
	flagQuiet := pflag.Lookup("quiet")

	cmdServe := &cobra.Command{
		Use:   "serve",
		Short: "Start a Kivik test server",
	}
	cmdServe.Flags().AddFlag(flagVerbose)
	cmdServe.Flags().AddFlag(flagQuiet)
	var listenAddr string
	cmdServe.Flags().StringVarP(&listenAddr, "http", "", ":5984", "HTTP bind address to serve")
	var driver string
	cmdServe.Flags().StringVarP(&driver, "driver", "d", "memory", "Backend driver to use")
	var dsn string
	cmdServe.Flags().StringVarP(&dsn, "dsn", "", "", "Data source name")
	cmdServe.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("Listening on %s\n", listenAddr)
		fmt.Println(serve.New(driver, dsn).Start(listenAddr))
		os.Exit(1)
	}

	cmdTest := &cobra.Command{
		Use:   "test",
		Short: "Run the test suite against the remote server",
	}
	cmdTest.Flags().AddFlag(flagVerbose)
	cmdTest.Flags().AddFlag(flagQuiet)
	cmdTest.Flags().StringVarP(&dsn, "dsn", "", "", "Data source name")
	var tests []string
	cmdTest.Flags().StringSliceVarP(&tests, "test", "", []string{"auto"}, "List of tests to run")
	var run string
	cmdTest.Flags().StringVarP(&run, "run", "", "", "Run only those tests matching the regular expression")
	cmdTest.Run = func(cmd *cobra.Command, args []string) {
		if err := test.RunTests("couch", dsn, tests, run); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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
