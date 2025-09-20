package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Installs    []Install `yaml:"installs"`
	Tools       []Tool    `yaml:"tools"`
	Wordlists   []Tool    `yaml:"wordlists"`
}

type Install struct {
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Commands    map[string][]string `yaml:"commands"`
}

type Tool struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Commands    []Command `yaml:"commands"`
	Condition   string    `yaml:"condition,omitempty"`
	Binary      string    `yaml:"binary,omitempty"`
	Path        string    `yaml:"path,omitempty"`
}

type Command struct {
	Cmd string   `yaml:"cmd,omitempty"`
	Or  []string `yaml:"or,omitempty"`
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			os.Exit(0)
		}
	}()

	configFile, dryRun, listTools, binaryCheck, skipAvailable, skipTools, showHelp := parseFlags()

	if showHelp {
		displayHelp()
		return
	}

	if configFile == "" {
		displayHelp()
		os.Exit(0)
	}

	config, err := loadConfig(configFile)
	if err != nil {
		red := color.New(color.FgRed)
		red.Print("error loading config: ")
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if listTools {
		displayTools(config)
		return
	}

	if binaryCheck {
		displayBinaryCheck(config)
		return
	}

	if dryRun {
		yellow := color.New(color.FgYellow)
		yellow.Println("dry run mode - commands will not be executed")
		fmt.Printf("\n")
	}

	executeInstalls(config, dryRun)
	executeTools(config, dryRun, skipAvailable, skipTools)
	executeWordlists(config, dryRun, skipAvailable, skipTools)
}

func parseFlags() (string, bool, bool, bool, bool, []string, bool) {
	configFile := flag.String("c", "", "YAML config file (required)")
	dryRun := flag.Bool("dry", false, "dry run mode - show commands without executing")
	listTools := flag.Bool("list", false, "list all tools in config")
	binaryCheck := flag.Bool("bc", false, "check binary availability for all tools")
	skipAvailable := flag.Bool("skip", false, "skip installation for available binaries/paths")
	skipToolsFlag := flag.String("st", "", "skip specific tools by name (comma-separated)")
	showHelp := flag.Bool("h", false, "show help")
	flag.Parse()

	var skipTools []string
	if *skipToolsFlag != "" {
		skipTools = strings.Split(*skipToolsFlag, ",")
		for i := range skipTools {
			skipTools[i] = strings.TrimSpace(skipTools[i])
		}
	}

	if *showHelp {
		return "", false, false, false, false, nil, true
	}

	if *listTools && *configFile != "" {
		return *configFile, false, true, false, false, nil, false
	}

	if *binaryCheck && *configFile != "" {
		return *configFile, false, false, true, false, nil, false
	}

	if *configFile == "" {
		return "", false, false, false, false, nil, false
	}

	return *configFile, *dryRun, false, false, *skipAvailable, skipTools, false
}

func displayHelp() {
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	white := color.New(color.FgWhite)
	red := color.New(color.FgRed)
	faint := color.New(color.Faint)

	fmt.Printf("\n")
	green.Println(" example:")
	fmt.Printf("    ")
	cyan.Print("shotit")
	fmt.Printf(" -c ")
	faint.Print("tools.yaml")
	fmt.Printf("\n")
	fmt.Printf("    ")
	cyan.Print("shotit")
	fmt.Printf(" -c ")
	faint.Print("tools.yaml")
	fmt.Printf(" -st ffuf,gau,\"gf patterns\"\n")
	fmt.Printf("    ")
	cyan.Print("shotit")
	fmt.Printf(" -c ")
	faint.Print("tools.yaml")
	fmt.Printf(" -list\n\n")

	green.Println(" options:")
	fmt.Printf("    ")
	white.Print("-c")
	fmt.Printf("      YAML config file ")
	red.Print("(required)")
	fmt.Printf("\n")
	fmt.Printf("    ")
	white.Print("-bc")
	fmt.Printf("     check binary availability for all tools\n")
	fmt.Printf("    ")
	white.Print("-st")
	fmt.Printf("     skip specific tools by name (comma-separated)\n")
	fmt.Printf("    ")
	white.Print("-dry")
	fmt.Printf("    dry run mode - show commands without executing\n")
	fmt.Printf("    ")
	white.Print("-list")
	fmt.Printf("   list all tools in config with binary status\n")
	fmt.Printf("    ")
	white.Print("-skip")
	fmt.Printf("   skip installation for available binaries/paths\n")
	fmt.Printf("    ")
	white.Print("-h")
	fmt.Printf("      show this help message\n\n")

	faint.Println("    made with <3 by @haq")
	fmt.Printf("\n")
}

func displayTools(config *Config) {
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	red := color.New(color.FgRed)

	green.Print("Config: ")
	fmt.Printf("%s\n", config.Name)
	if config.Description != "" {
		green.Print("Description: ")
		fmt.Printf("%s\n", config.Description)
	}

	if len(config.Installs) > 0 {
		fmt.Printf("\n")
		green.Println("Installs:")
		for _, install := range config.Installs {
			fmt.Printf("  ")
			cyan.Print(install.Name)
			if install.Description != "" {
				fmt.Printf(" - %s", install.Description)
			}
			fmt.Printf("\n")
		}
	}

	if len(config.Tools) > 0 {
		fmt.Printf("\n")
		green.Println("Tools:")
		for _, tool := range config.Tools {
			fmt.Printf("  ")
			cyan.Print(tool.Name)
			if tool.Description != "" {
				fmt.Printf(" - %s", tool.Description)
			}
			if tool.Binary != "" {
				fmt.Printf(" [")
				if checkBinaryExists(tool.Binary) {
					green.Print(tool.Binary)
				} else {
					red.Print(tool.Binary)
				}
				fmt.Printf("]")
			}
			fmt.Printf("\n")
		}
	}

	if len(config.Wordlists) > 0 {
		fmt.Printf("\n")
		green.Println("Wordlists:")
		for _, wordlist := range config.Wordlists {
			fmt.Printf("  ")
			cyan.Print(wordlist.Name)
			if wordlist.Description != "" {
				fmt.Printf(" - %s", wordlist.Description)
			}
			if wordlist.Binary != "" {
				fmt.Printf(" [")
				if checkBinaryExists(wordlist.Binary) {
					green.Print(wordlist.Binary)
				} else {
					red.Print(wordlist.Binary)
				}
				fmt.Printf("]")
			} else if wordlist.Path != "" {
				fmt.Printf(" [")
				if checkPathExists(wordlist.Path) {
					green.Print(wordlist.Path)
				} else {
					red.Print(wordlist.Path)
				}
				fmt.Printf("]")
			}
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}

func loadConfig(filename string) (*Config, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %v", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	return &config, nil
}

func executeInstalls(config *Config, dryRun bool) {
	if len(config.Installs) == 0 {
		return
	}

	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	faint := color.New(color.Faint)

	green.Print("executing installs: ")
	fmt.Printf("%s\n\n", config.Name)

	for _, install := range config.Installs {
		cyan.Print("[" + install.Name + "]")
		if install.Description != "" {
			fmt.Printf(" %s", install.Description)
		}
		fmt.Printf("\n")

		executed := false
		for pkgManager, commands := range install.Commands {
			if checkPackageManager(pkgManager) {
				fmt.Printf("  ")
				yellow.Printf("using %s\n", pkgManager)
				for _, command := range commands {
					fmt.Printf("  ")
					faint.Printf("$ %s\n", command)

					if !dryRun {
						if err := executeCommand(command); err != nil {
							fmt.Printf("  ")
							red.Printf("Error: %v\n", err)
							break
						}
					}
				}
				executed = true
				break
			}
		}

		if !executed {
			fmt.Printf("  ")
			yellow.Println("no compatible package manager found")
		}
		fmt.Printf("\n")
	}
}

func executeTools(config *Config, dryRun bool, skipAvailable bool, skipTools []string) {
	if len(config.Tools) == 0 {
		return
	}

	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	faint := color.New(color.Faint)

	green.Print("executing tools: ")
	fmt.Printf("%s\n\n", config.Name)

	for _, tool := range config.Tools {
		fmt.Printf("[")
		cyan.Print(tool.Name)
		fmt.Printf("] %s\n", tool.Description)

		if contains(skipTools, tool.Name) {
			fmt.Printf("  ")
			yellow.Printf("skipping - %s specified in -st\n\n", tool.Name)
			continue
		}

		if tool.Binary != "" && checkBinaryExists(tool.Binary) {
			fmt.Printf("  ")
			if skipAvailable {
				yellow.Printf("skipping - %s already available\n\n", tool.Binary)
			} else {
				green.Printf("already installed - %s found\n\n", tool.Binary)
			}
			continue
		}

		if tool.Condition != "" && !checkCondition(tool.Condition) {
			fmt.Printf("  ")
			yellow.Printf("skipping - condition not met: %s\n\n", tool.Condition)
			continue
		}

		for _, command := range tool.Commands {
			if command.Cmd != "" {
				fmt.Printf("  ")
				faint.Printf("$ %s\n", command.Cmd)

				if !dryRun {
					if err := executeShellCommand(command.Cmd, dryRun); err != nil {
						fmt.Printf("  ")
						red.Printf("Error: %v\n", err)
						continue
					}
				}
			} else if len(command.Or) > 0 {
				executed := false
				for _, orCmd := range command.Or {
					fmt.Printf("  ")
					faint.Print("$ " + orCmd)

					if !dryRun {
						if err := executeShellCommand(orCmd, dryRun); err != nil {
							fmt.Printf(" ")
							red.Print("(failed)")
							fmt.Printf("\n")
							continue
						} else {
							fmt.Printf(" ")
							green.Print("(success)")
							fmt.Printf("\n")
							executed = true
							break
						}
					} else {
						fmt.Printf("\n")
					}
				}
				if !dryRun && !executed {
					fmt.Printf("  ")
					red.Println("all fallback commands failed")
				}
			}
		}
		fmt.Printf("\n")
	}

	if dryRun {
		yellow.Println("dry run completed - no commands were executed")
	} else {
		green.Println("installation completed")
	}
}

func executeWordlists(config *Config, dryRun bool, skipAvailable bool, skipTools []string) {
	if len(config.Wordlists) == 0 {
		return
	}

	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	faint := color.New(color.Faint)

	green.Print("executing wordlists: ")
	fmt.Printf("%s\n\n", config.Name)

	for _, wordlist := range config.Wordlists {
		fmt.Printf("[")
		cyan.Print(wordlist.Name)
		fmt.Printf("] %s\n", wordlist.Description)

		if contains(skipTools, wordlist.Name) {
			fmt.Printf("  ")
			yellow.Printf("skipping - %s specified in -st\n\n", wordlist.Name)
			continue
		}

		if (wordlist.Binary != "" && checkBinaryExists(wordlist.Binary)) || (wordlist.Path != "" && checkPathExists(wordlist.Path)) {
			if wordlist.Binary != "" {
				fmt.Printf("  ")
				if skipAvailable {
					yellow.Printf("skipping - %s already available\n\n", wordlist.Binary)
				} else {
					green.Printf("already installed - %s found\n\n", wordlist.Binary)
				}
			} else {
				fmt.Printf("  ")
				if skipAvailable {
					yellow.Printf("skipping - %s already available\n\n", wordlist.Path)
				} else {
					green.Printf("already exists - %s found\n\n", wordlist.Path)
				}
			}
			continue
		}

		if wordlist.Condition != "" && !checkCondition(wordlist.Condition) {
			fmt.Printf("  ")
			yellow.Printf("skipping - condition not met: %s\n\n", wordlist.Condition)
			continue
		}

		for _, command := range wordlist.Commands {
			if command.Cmd != "" {
				fmt.Printf("  ")
				faint.Printf("$ %s\n", command.Cmd)

				if !dryRun {
					if err := executeCommand(command.Cmd); err != nil {
						fmt.Printf("  ")
						red.Printf("Error: %v\n", err)
						continue
					}
				}
			} else if len(command.Or) > 0 {
				executed := false
				for _, orCmd := range command.Or {
					fmt.Printf("  ")
					faint.Print("$ " + orCmd)

					if !dryRun {
						if err := executeCommand(orCmd); err != nil {
							fmt.Printf(" ")
							red.Print("(failed)")
							fmt.Printf("\n")
							continue
						} else {
							fmt.Printf(" ")
							green.Print("(success)")
							fmt.Printf("\n")
							executed = true
							break
						}
					} else {
						fmt.Printf("\n")
					}
				}
				if !dryRun && !executed {
					fmt.Printf("  ")
					red.Println("all fallback commands failed")
				}
			}
		}
		fmt.Printf("\n")
	}

	if dryRun {
		yellow.Println("dry run completed - no commands were executed")
	} else {
		green.Println("wordlists installation completed")
	}
}

func displayBinaryCheck(config *Config) {
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	green.Print("binary check: ")
	fmt.Printf("%s\n\n", config.Name)

	availableCount := 0
	unavailableCount := 0

	if len(config.Tools) > 0 {
		green.Println("Tools:")
		for _, tool := range config.Tools {
			if tool.Binary != "" {
				if checkBinaryExists(tool.Binary) {
					fmt.Printf("  ")
					green.Print("✓ ")
					cyan.Print(tool.Binary)
					fmt.Printf(" (%s)\n", tool.Name)
					availableCount++
				} else {
					fmt.Printf("  ")
					red.Print("✗ ")
					cyan.Print(tool.Binary)
					fmt.Printf(" (%s)\n", tool.Name)
					unavailableCount++
				}
			} else {
				fmt.Printf("  ")
				yellow.Print("- ")
				cyan.Print(tool.Name)
				fmt.Printf(" (no binary specified)\n")
			}
		}
		fmt.Printf("\n")
	}

	if len(config.Wordlists) > 0 {
		green.Println("Wordlists:")
		for _, wordlist := range config.Wordlists {
			if wordlist.Binary != "" {
				if checkBinaryExists(wordlist.Binary) {
					fmt.Printf("  ")
					green.Print("✓ ")
					cyan.Print(wordlist.Binary)
					fmt.Printf(" (%s)\n", wordlist.Name)
					availableCount++
				} else {
					fmt.Printf("  ")
					red.Print("✗ ")
					cyan.Print(wordlist.Binary)
					fmt.Printf(" (%s)\n", wordlist.Name)
					unavailableCount++
				}
			} else if wordlist.Path != "" {
				if checkPathExists(wordlist.Path) {
					fmt.Printf("  ")
					green.Print("✓ ")
					cyan.Print(wordlist.Path)
					fmt.Printf(" (%s)\n", wordlist.Name)
					availableCount++
				} else {
					fmt.Printf("  ")
					red.Print("✗ ")
					cyan.Print(wordlist.Path)
					fmt.Printf(" (%s)\n", wordlist.Name)
					unavailableCount++
				}
			} else {
				fmt.Printf("  ")
				yellow.Print("- ")
				cyan.Print(wordlist.Name)
				fmt.Printf(" (no binary/path specified)\n")
			}
		}
		fmt.Printf("\n")
	}

	green.Printf("%d available", availableCount)
	fmt.Printf(", ")
	red.Printf("%d unavailable", unavailableCount)
	fmt.Printf("\n\n")
}

func checkBinaryExists(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}

func checkPathExists(path string) bool {
	expandedPath := os.ExpandEnv(path)
	_, err := os.Stat(expandedPath)
	return err == nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func executeShellCommand(command string, dryRun bool) error {
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("empty command")
	}

	if dryRun {
		return nil
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func checkCondition(condition string) bool {
	switch condition {
	case "linux":
		return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "linux") || fileExists("/etc/os-release")
	case "macos":
		return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "darwin") || fileExists("/System/Library/CoreServices/SystemVersion.plist")
	case "windows":
		return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "windows")
	default:
		_, err := exec.LookPath(condition)
		return err == nil
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func checkPackageManager(pkgManager string) bool {
	switch pkgManager {
	case "apt":
		_, err := exec.LookPath("apt")
		return err == nil
	case "pacman":
		_, err := exec.LookPath("pacman")
		return err == nil
	case "yum":
		_, err := exec.LookPath("yum")
		return err == nil
	case "dnf":
		_, err := exec.LookPath("dnf")
		return err == nil
	case "brew":
		_, err := exec.LookPath("brew")
		return err == nil
	case "apk":
		_, err := exec.LookPath("apk")
		return err == nil
	case "zypper":
		_, err := exec.LookPath("zypper")
		return err == nil
	default:
		_, err := exec.LookPath(pkgManager)
		return err == nil
	}
}

func executeCommand(command string) error {
	return executeShellCommand(command, false)
}
