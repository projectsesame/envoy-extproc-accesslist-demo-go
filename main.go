package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
)

type processor interface {
	Init(opts *ep.ProcessingOptions, cmdArgs []string) error
	Finish()

	ep.RequestProcessor
}

var processors = map[string]processor{
	"allow-and-block": &allowAndBlockRequestProcessor{},
}

const (
	kAllowList = "allowlist"
	kBlockList = "blocklist"
	flagPrefix = "--"

	errMsg = "only one of --allowlist and --blocklist can be specified"
)

func parseArgs(args []string) (port *int, opts *ep.ProcessingOptions, remainArgs []string) {
	rootCmd := flag.NewFlagSet("root", flag.ExitOnError)
	port = rootCmd.Int("port", 50051, "the gRPC port.")

	opts = ep.NewDefaultOptions()

	rootCmd.BoolVar(&opts.LogStream, "log-stream", false, "log the stream or not.")
	rootCmd.BoolVar(&opts.LogPhases, "log-phases", false, "log the phases or not.")
	rootCmd.BoolVar(&opts.UpdateExtProcHeader, "update-extproc-header", false, "update the extProc header or not.")
	rootCmd.BoolVar(&opts.UpdateDurationHeader, "update-duration-header", false, "update the duration header or not.")

	cnt := 0

	paramChecker := func() error {
		cnt++
		if cnt >= 2 {
			return fmt.Errorf(errMsg)
		}
		return nil

	}
	// put back for later use
	rootCmd.Func(kAllowList, fmt.Sprintf("the white ip list (%s)", errMsg), func(s string) error {
		remainArgs = append(remainArgs, flagPrefix+kAllowList, s)
		return paramChecker()
	})
	rootCmd.Func(kBlockList, fmt.Sprintf("the black ip list (%s)", errMsg), func(s string) error {
		remainArgs = append(remainArgs, flagPrefix+kBlockList, s)
		return paramChecker()
	})

	rootCmd.Parse(args)

	remainArgs = append(remainArgs, rootCmd.Args()...)
	return
}

func main() {

	// cmd subCmd arg, arg2,...
	args := os.Args
	if len(args) < 2 {
		log.Fatal("Passing a processor is required.")
	}

	cmd := args[1]
	proc, exists := processors[cmd]
	if !exists {
		log.Fatalf("Processor \"%s\" not defined.", cmd)
	}

	port, opts, cmdArgs := parseArgs(os.Args[2:])
	if err := proc.Init(opts, cmdArgs); err != nil {
		log.Fatalf("Initialize the processor is failed: %v.", err.Error())
	}
	defer proc.Finish()

	ep.Serve(*port, proc)
}
