package main

import (
	"flag"
	"fmt"
	"runtime"
)

func main() {
	port := flag.String("port", "80", "a string")
	root := flag.String("root", ".", "a string")
	cpuMax := flag.Int("cpuMax", 1, "an int")
	flag.Parse()
	fmt.Println("cpuMax", *cpuMax)
	runtime.GOMAXPROCS(*cpuMax)

	StartServer(*port, *root)
}
