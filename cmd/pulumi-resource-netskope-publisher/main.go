package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "version":
			fmt.Printf("pulumi-resource-netskope-publisher %s\n", version)
			return
		case "--schema":
			fmt.Fprintln(os.Stderr, "schema is published from schema.json in the repository")
			os.Exit(1)
		}
	}

	fmt.Fprintln(os.Stderr, "pulumi-resource-netskope-publisher is a release shim for the TypeScript component package")
	fmt.Fprintln(os.Stderr, "Use @johnneerdael/pulumi-netskope-publisher from npm for current component execution")
	os.Exit(1)
}
