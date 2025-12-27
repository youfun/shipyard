package cliutils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PromptForSelection displays a list of items to the user and prompts them to select one.
// It supports a default selection, which is chosen if the user just presses Enter.
// It returns the zero-based index of the selected item.
func PromptForSelection(promptTitle string, items []string, defaultIndex int) (int, error) {
	fmt.Println(promptTitle)
	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, item)
	}

	prompt := "Enter the number: "
	if defaultIndex >= 0 && defaultIndex < len(items) {
		prompt = fmt.Sprintf("Enter the number (default %d): ", defaultIndex+1)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Handle default selection on empty input
		if input == "" && defaultIndex >= 0 && defaultIndex < len(items) {
			return defaultIndex, nil
		}

		// Handle numeric input
		selection, err := strconv.Atoi(input)
		if err == nil && selection > 0 && selection <= len(items) {
			return selection - 1, nil // Return zero-based index
		}
		fmt.Println("Invalid input, please try again.")
	}
}

// PromptForInput prompts the user for text input with an optional default value.
func PromptForInput(prompt string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultValue != "" {
		fmt.Printf("%s (default: %s): ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" && defaultValue != "" {
		return defaultValue
	}
	return input
}

// PromptForConfirmation asks the user a yes/no question and returns true if they confirm.
func PromptForConfirmation(prompt string, defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)
	suffix := " [y/N]: "
	if defaultYes {
		suffix = " [Y/n]: "
	}

	for {
		fmt.Print(prompt + suffix)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" {
			return defaultYes
		}

		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}

		fmt.Println("Invalid input, please enter y or n.")
	}
}
