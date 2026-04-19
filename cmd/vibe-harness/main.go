package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/jgervais/vibe_harness/internal/config"
	"github.com/jgervais/vibe_harness/internal/output"
	"github.com/jgervais/vibe_harness/internal/scanner"
)

var (
	version   = "dev"
	rulesHash = "unknown"
)

var (
	versionFlag = flag.Bool("version", false, "Print version and exit")
	formatFlag  = flag.String("format", "human", "Output format: human, json, sarif")
	configFlag  = flag.String("config", "", "Path to config file")
	helpFlag    = flag.Bool("help", false, "Print usage and exit")
)

func main() {
	flag.Parse()

	if *helpFlag {
		fmt.Fprintf(os.Stderr, "Usage: vibe-harness [flags] <path>\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Printf("vibe-harness v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		fmt.Printf("rules hash: %s\n", rulesHash)
		os.Exit(0)
	}

	target := flag.Arg(0)
	if target == "" {
		fmt.Fprintf(os.Stderr, "error: path argument is required\n")
		os.Exit(2)
	}

	switch *formatFlag {
	case "human", "json", "sarif":
	default:
		fmt.Fprintf(os.Stderr, "error: invalid format %q (must be human, json, or sarif)\n", *formatFlag)
		os.Exit(2)
	}

	var cfg *config.Config
	if *configFlag != "" {
		if err := config.ValidateTOML(*configFlag); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		loaded, err := config.LoadConfig(*configFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		cfg = loaded
	} else {
		discovered, _ := config.AutoDiscoverConfig(target)
		if discovered != "" {
			if err := config.ValidateTOML(discovered); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(2)
			}
			loaded, err := config.LoadConfig(discovered)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(2)
			}
			cfg = loaded
		} else {
			d := config.DefaultConfig()
			cfg = &d
		}
	}

	result, err := scanner.Scan(target, cfg, version, rulesHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	switch *formatFlag {
	case "human":
		output.FormatHuman(os.Stderr, result)
	case "json":
		data, err := output.FormatJSON(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		os.Stdout.Write(data)
		fmt.Fprintln(os.Stdout)
	case "sarif":
		data, err := output.FormatSARIF(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		os.Stdout.Write(data)
		fmt.Fprintln(os.Stdout)
	}

	os.Exit(result.ExitCode)
}