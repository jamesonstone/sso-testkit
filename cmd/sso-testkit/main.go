package main

import (
	"os"

	"github.com/jamesonstone/sso-testkit/internal/app"
)

func main() {
	os.Exit(app.Main(os.Args[1:], os.Stdout, os.Stderr, os.LookupEnv))
}
