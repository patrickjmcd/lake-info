package main

import (
	"github.com/patrickjmcd/lake-info/cmd"
	"github.com/patrickjmcd/lake-info/logger"
)

func main() {
	cmd.Execute()
}

func init() {
	logger.Setup()
}
