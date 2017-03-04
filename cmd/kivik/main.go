package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/memory"
	"github.com/flimzy/kivik/driver/proxy"
	"github.com/flimzy/kivik/logger/file"
	"github.com/flimzy/kivik/serve"
	"github.com/flimzy/kivik/test"
)

func main() {
	var verbose bool
	pflag.BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	flagVerbose := pflag.Lookup("verbose")

	cmdServe := &cobra.Command{
		Use:   "serve",
		Short: "Start a Kivik test server",
	}
	cmdServe.Flags().AddFlag(flagVerbose)
	var listenAddr string
	cmdServe.Flags().StringVarP(&listenAddr, "http", "", ":5984", "HTTP bind address to serve")
	var driver string
	cmdServe.Flags().StringVarP(&driver, "driver", "d", "memory", "Backend driver to use")
	var dsn string
	cmdServe.Flags().StringVarP(&dsn, "dsn", "", "", "Data source name")
	var logFile string
	cmdServe.Flags().StringVarP(&logFile, "log", "l", "", "Server log file")
	cmdServe.Run = func(cmd *cobra.Command, args []string) {
		var logger *file.Logger
		if logFile != "" {
			var err error
			logger, err = file.New(logFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open log file '%s': %s", logFile, err)
				os.Exit(1)
			}
		}

		client, err := kivik.New(driver, dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect: %s", err)
			os.Exit(1)
		}
		loggingClient := serve.LoggingClient{
			Client:    proxy.NewClient(client),
			Logger:    logger,
			LogWriter: logger,
		}
		fmt.Printf("Listening on %s\n", listenAddr)
		fmt.Println(serve.New(loggingClient).Start(listenAddr))
		os.Exit(1)
	}

	cmdTest := &cobra.Command{
		Use:   "test [Remote Server DSN]",
		Short: "Run the test suite against the remote server",
	}
	cmdTest.Flags().AddFlag(flagVerbose)
	// cmdTest.Flags().StringVarP(&dsn, "dsn", "", "", "Data source name")
	var tests []string
	cmdTest.Flags().StringSliceVarP(&tests, "test", "", []string{"auto"}, "List of tests to run")
	var listTests bool
	cmdTest.Flags().BoolVarP(&listTests, "list", "l", false, "List available tests")
	var run string
	cmdTest.Flags().StringVarP(&run, "run", "", "", "Run only those tests matching the regular expression")
	var rw bool
	cmdTest.Flags().BoolVarP(&rw, "write", "w", false, "Allow tests which write to the database")
	var cleanup bool
	cmdTest.Flags().BoolVarP(&cleanup, "cleanup", "c", false, "Clean up after previous test run, then exit")
	cmdTest.Run = func(cmd *cobra.Command, args []string) {
		if listTests {
			test.ListTests()
			os.Exit(0)
		}
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}
		test.RunTests(test.Options{
			Driver:  "couch",
			DSN:     args[0],
			Verbose: verbose,
			RW:      rw,
			Suites:  tests,
			Match:   run,
			Cleanup: cleanup,
		})
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
