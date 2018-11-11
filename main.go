package main

import "flag"
import "fmt"
import "./monitor"

const (
	DefaultStateDirectory  = ".gonitor"
	DefaultStateFile = "statefile"
)

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
	directory := flag.String("dir", ".", "Set directory to monitor.")
	stateFileEnabled := flag.Bool("s", true, "Disable state file use.")
	stateFileDirectory := flag.String("d", DefaultStateDirectory, "Set custom state file directory.");
	stateFileName := flag.String("n", DefaultStateFile, "Set custom state file name.")

	flag.Parse()

	m := monitor.New()

	m.SetDirectory(*directory)

	if *stateFileEnabled {
		m.SetStateFile(*stateFileDirectory, *stateFileName)
		if *clearStateFiles {
			m.ClearStateFiles()
		}
	}

	m.SetPollInterval(*pollInterval)

	m.SetOnAddFunc(OnAdd)
	m.SetOnDelFunc(OnDel)
	m.SetOnModFunc(OnMod)

	m.Watch()
}
