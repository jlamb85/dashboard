package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

const (
	bcryptCost = 12 // Higher = more secure but slower
)

func main() {
	configFile := flag.String("config", "config/config.yaml", "Path to config.yaml file")
	updateConfig := flag.Bool("update", false, "Update config.yaml with the hash")
	envFormat := flag.Bool("env", false, "Output as environment variable format")
	flag.Parse()

	// Read password from stdin (hidden)
	fmt.Print("Enter password to hash: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
		os.Exit(1)
	}

	password := strings.TrimSpace(string(passwordBytes))
	if password == "" {
		fmt.Fprintf(os.Stderr, "Password cannot be empty\n")
		os.Exit(1)
	}

	// Confirm password
	fmt.Print("Confirm password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading confirmation: %v\n", err)
		os.Exit(1)
	}

	if password != strings.TrimSpace(string(confirmBytes)) {
		fmt.Fprintf(os.Stderr, "Passwords do not match\n")
		os.Exit(1)
	}

	// Generate bcrypt hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating hash: %v\n", err)
		os.Exit(1)
	}

	hashString := string(hash)

	// Output formats
	if *envFormat {
		fmt.Println("\n✓ Password hashed successfully!")
		fmt.Println("\nAdd this to your environment:")
		fmt.Printf("export AUTH_PASSWORD='%s'\n", hashString)
		fmt.Println("\nOr add to your shell profile (~/.bashrc, ~/.zshrc, etc.):")
		fmt.Printf("export AUTH_PASSWORD='%s'\n", hashString)
	} else if *updateConfig {
		if err := updateConfigFile(*configFile, hashString); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n✓ Config file updated: %s\n", *configFile)
		fmt.Println("\nPassword hash has been saved to config.yaml")
		fmt.Println("The application will now use bcrypt authentication.")
	} else {
		fmt.Println("\n✓ Password hashed successfully!")
		fmt.Println("\nBcrypt hash:")
		fmt.Println(hashString)
		fmt.Println("\nYou can:")
		fmt.Println("  1. Use --update to save to config.yaml")
		fmt.Println("  2. Use --env for environment variable format")
		fmt.Printf("  3. Manually set: export AUTH_PASSWORD='%s'\n", hashString)
	}
}

func updateConfigFile(configPath string, hash string) error {
	// Read the config file as raw text to preserve comments and formatting
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Read line by line and update the password line
	lines := strings.Split(string(content), "\n")
	updated := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Look for the password line
		if strings.HasPrefix(trimmed, "password:") {
			// Preserve indentation
			indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " ")))
			lines[i] = fmt.Sprintf("%spassword: \"%s\"  # Bcrypt hash - DO NOT share this value", indent, hash)
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("could not find password field in config file")
	}

	// Write back to file
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
