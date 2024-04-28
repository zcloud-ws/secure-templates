package main

import (
	"github.com/edimarlnx/secure-templates/pkg/app"
	"os"
)

func main() {
	app.InitApp(os.Args, nil, nil)
}
