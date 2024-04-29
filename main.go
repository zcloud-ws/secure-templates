package main

import (
	"github.com/zcloud-ws/secure-templates/pkg/app"
	"os"
)

func main() {
	app.InitApp(os.Args, nil, nil)
}
