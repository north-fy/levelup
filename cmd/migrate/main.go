package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/north-fy/levelup/internal/config"
)

func main() {
	printDSN := flag.Bool("dsn", false, "print postgres DSN and exit")
	printCHDSN := flag.Bool("ch-dsn", false, "print clickhouse DSN and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch {
	case *printDSN:
		fmt.Print(cfg.DB.DSN())
	case *printCHDSN:
		// golang-migrate clickhouse driver uses the HTTP interface (port 8123).
		fmt.Printf(
			"clickhouse://%s:%d?username=%s&password=%s&database=%s&x-multi-statement=true",
			cfg.CH.Host, 8123, cfg.CH.User, cfg.CH.Password, cfg.CH.Database,
		)
	default:
		flag.Usage()
	}
}
