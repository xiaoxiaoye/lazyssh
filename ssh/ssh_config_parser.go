package ssh

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

// SSHConfig represents an SSH configuration.
type SSHConfig struct {
	Host         string
	Hostname     string
	User         string
	Port         string
	IdentityFile string
	otherOptions map[string]string
}

func (c SSHConfig) String() string {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

// ParseSSHConfig parses the .ssh/config file and returns a map of SSH configurations.
func ParseSSHConfig(filePath string) (map[string]SSHConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	configs := make(map[string]SSHConfig)
	scanner := bufio.NewScanner(file)

	var currentHost string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			// Skip empty lines and comments
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value := strings.Join(parts[1:], " ")

		switch strings.ToLower(key) {
		case "host":
			if value == "*" {
				continue
			}
			currentHost = value
			configs[currentHost] = SSHConfig{
				Host:         currentHost,
				otherOptions: make(map[string]string),
			}
		case "hostname":
			config := configs[currentHost]
			config.Hostname = value
			configs[currentHost] = config
		case "user":
			config := configs[currentHost]
			config.User = value
			configs[currentHost] = config
		case "port":
			config := configs[currentHost]
			config.Port = value
			configs[currentHost] = config
		case "identityfile":
			config := configs[currentHost]
			config.IdentityFile = value
			configs[currentHost] = config
		default:
			// Handle other options
			// config := configs[currentHost]
			// config.otherOptions[key] = value
			// configs[currentHost] = config
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for host, config := range configs {
		if config.Hostname == "" && config.User == "" {
			delete(configs, host)
		}
	}

	return configs, nil
}

// func main() {
// 	sshConfigPath := os.ExpandEnv("$HOME/.ssh/config")
// 	configs, err := ParseSSHConfig(sshConfigPath)
// 	if err != nil {
// 		fmt.Printf("Error parsing SSH config: %v\n", err)
// 		return
// 	}

// 	for host, config := range configs {
// 		fmt.Printf("Host: %s\n", host)
// 		fmt.Printf("  Hostname: %s\n", config.Hostname)
// 		fmt.Printf("  User: %s\n", config.User)
// 		fmt.Printf("  Port: %s\n", config.Port)
// 		fmt.Printf("  IdentityFile: %s\n", config.IdentityFile)
// 		for key, value := range config.otherOptions {
// 			fmt.Printf("  %s: %s\n", key, value)
// 		}
// 		fmt.Println()
// 	}
// }
