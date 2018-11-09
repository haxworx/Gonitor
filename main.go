package main;

import "fmt"
import "./monitor"

func OnAdd(filepath string) {
	fmt.Printf("ADD: %s\n", filepath);
}

func OnDel(filepath string) {
	fmt.Printf("DEL: %s\n", filepath);
}

func OnMod(filepath string) {
	fmt.Printf("MOD: %s\n", filepath);
}

func main() {
	m := monitor.New();

	m.SetDirectory(".");

	m.SetPollInterval(3);

	m.SetOnAddFunc(OnAdd)
	m.SetOnDelFunc(OnDel)
	m.SetOnModFunc(OnMod)

	m.Watch();
}
