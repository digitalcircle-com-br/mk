package lib

import "github.com/alecthomas/kong"

var CLI = struct {
	File          string   `help:"File to be used - Defaults are: .mk.yaml, .mk, mk, mk.yaml" short:"f"`
	Init          bool     `help:"Creates a new empty file (default is .mk.yaml in case no filename is provided)" short:"i"`
	Tasks         []string `arg:"" help:"Tasks to be run - Default is main." default:"."`
	Ver           bool     `help:"Prints version and exit" short:"v"`
	List          bool     `help:"Check file and print tasks" short:"l"`
	Dbg           bool     `help:"Debugs execution" short:"d"`
	DumpValidator bool     `help:"Dumps Validator JSON File" default:"false"`
	Env           bool     `help:"Dumps env and vars" default:"false" short:"e"`
}{}

func InitCli() {
	kong.Parse(&CLI)
	Ver(CLI.Ver)
}
