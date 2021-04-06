package main

import (
	"log"

	"ldc.io/ldc/cmd/ldc-apiserver/app"
)

func main() {

	cmd := app.NewAPIServerCommand()

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
