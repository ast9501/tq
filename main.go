package main

import (
	"fmt"
	"log"

	"github.com/ast9501/tq/pkg/cli"
)

func main() {
	argMap, buf, query, err := cli.CommandLineArgsBuffer(true)
	if err != nil {
		log.Fatal(err)
	}
	tqb, err := cli.Query(argMap, query, buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(tqb))
}
