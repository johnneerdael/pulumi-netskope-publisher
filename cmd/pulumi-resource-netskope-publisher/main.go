package main

import (
	"context"
	"fmt"
	"os"

	netskopeprovider "github.com/johnneerdael/pulumi-netskope-publisher/internal/provider"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "version":
			fmt.Printf("pulumi-resource-netskope-publisher %s\n", version)
			return
		case "--schema":
			schema, err := netskopeprovider.Schema(context.Background(), 1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "schema error: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(schema)
			return
		}
	}

	provider, err := netskopeprovider.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "provider error: %s\n", err)
		os.Exit(1)
	}

	if err := provider.Run(context.Background(), netskopeprovider.Name, version); err != nil {
		fmt.Fprintf(os.Stderr, "provider error: %s\n", err)
		os.Exit(1)
	}
}
