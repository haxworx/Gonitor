package main

import "flag"
import "fmt"
import "os"
import "./monitor"

func OnAdd(filepath string) {
	fmt.Printf("ADD: %s\n", filepath)
}

func OnDel(filepath string) {
	fmt.Printf("DEL: %s\n", filepath)
}

func OnMod(filepath string) {
	fmt.Printf("MOD: %s\n", filepath)
}

func main() {
	pollInterval := flag.Int("i", 3, "Set poll interval")
	clearStateFiles := flag.Bool("c", false, "Clear existing state files.")
	directory := flag.String("d", ".", "Set directory to monitor.")

	flag.Parse()

	m := monitor.New()

	m.SetDirectory(*directory)

	if *clearStateFiles {
		m.ClearStateFiles()
	}

	m.SetPollInterval(*pollInterval)

	m.SetOnAddFunc(OnAdd)
	m.SetOnDelFunc(OnDel)
	m.SetOnModFunc(OnMod)

	m.Watch()
}
