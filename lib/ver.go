package lib

import "os"

import _ "embed"

//go:embed ver.txt
var ver string

func Ver(end bool) {
	println("DC MK Tool - ver: " + ver)
	if end {
		os.Exit(0)
	}
}
