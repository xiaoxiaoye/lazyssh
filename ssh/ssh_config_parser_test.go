package ssh

import (
	"fmt"
	"os"
	"testing"
)

func Test_ParseSSH(t *testing.T) {
	sshConfigPath := os.ExpandEnv("$HOME/.ssh/config")
	configs, err := ParseSSHConfig(sshConfigPath)
	if err != nil {
		fmt.Printf("Error parsing SSH config: %v\n", err)
		return
	}

	for host, config := range configs {
		fmt.Printf("Host: %s\n", host)
		fmt.Printf("  Hostname: %s\n", config.Hostname)
		fmt.Printf("  User: %s\n", config.User)
		fmt.Printf("  Port: %s\n", config.Port)
		fmt.Printf("  IdentityFile: %s\n", config.IdentityFile)
		for key, value := range config.otherOptions {
			fmt.Printf("  %s: %s\n", key, value)
		}
		fmt.Println()
	}
}
