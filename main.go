package main

import "github.com/nlink-jp/gem-cli/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}
