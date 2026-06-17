package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"taper/storage"
	"taper/ui"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help") {
		printUsage()
		return
	}

	if err := storage.EnsureLayout(); err != nil {
		log.Fatalf("taper: config layout: %v", err)
	}

	model, err := ui.NewModel()
	if err != nil {
		log.Fatalf("taper: init: %v", err)
	}

	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Println("Taper — terminal race tracking and planning")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  taper              start the TUI")
	fmt.Println("  taper help         show this help")
	fmt.Println()
	fmt.Println("Data is stored in ~/.config/taper/")
	fmt.Println("  races.json         race list and settings")
	fmt.Println("  races/<id>/        per-race log.md, strategy.md, packing.md")
}
