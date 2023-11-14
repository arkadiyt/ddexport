package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/arkadiyt/ddexport"
)

func usage(flagset *flag.FlagSet) {
	fmt.Printf(`Usage:
	ddexport logs <options>
	ddexport spans <options>
Options:
`)
	flagset.PrintDefaults()
	os.Exit(1)
}

func main() {
	var output string
	var query string
	var from string
	var to string
	var limit int
	flagset := flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&output, "output", "", "The output file to save results to")
	flagset.StringVar(&query, "query", "", "The query to search for")
	flagset.StringVar(&from, "from", "now-30d", "The time range to search from")
	flagset.StringVar(&to, "to", "now", "The time range to search to")
	flagset.IntVar(&limit, "limit", 250, "The number of results per page")

	if len(os.Args) < 2 {
		usage(flagset)
	}

	flagset.Parse(os.Args[2:])

	if output == "" || query == "" {
		usage(flagset)
	}

	outputFile, err := os.Create(output)
	if err != nil {
		log.Fatalf("Error opening file '%s': %v", output, err)
	}
	ddexport, err := ddexport.New(query, to, from, limit, outputFile)
	if err != nil {
		log.Fatalf("Initialization error %v\n", err)
	}

	switch os.Args[1] {
	case "logs":
		ddexport.SearchLogs()
	case "spans":
		ddexport.SearchSpans()
	default:
		log.Fatalf("Unknown subcommand, use 'logs' or 'spans'")
	}

	outputFile.Close()
}
