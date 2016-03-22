package main

import (
	"flag"

	"github.com/mahata/gwiki/wiki"
)

var (
	useNginx = flag.Bool("nginx", false, "If the service uses Nginx")
)

func main() {
	flag.Parse()
	wiki.Run(*useNginx)
}
