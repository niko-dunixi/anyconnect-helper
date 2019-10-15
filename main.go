package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/keyring"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	service   = "anyconnect-helper"
	entryName = "user-data"
)

type CredentialData struct {
	VpnHost  string `json:"vpn-host"`
	Group    string `json:"group"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func printHelp() {
	fmt.Println("Only allowed parameters:")
	fmt.Println(" --help		Shows this")
	fmt.Println(" --disconnect		Disconnects VPN")
	fmt.Println(" --reconnect		Disconnects and reconnects VPN")
}

func main() {
	args := getArgs(os.Args)
	if len(args) > 1 {
		printHelp()
		os.Exit(1)
	} else if len(args) == 1 {
		shouldExit := true
		argument := args[0]
		switch argument {
		case "--reconnect":
			disconnect()
			shouldExit = false
		case "--disconnect":
			disconnect()
		case "--help":
		default:
			printHelp()
		}
		if shouldExit {
			os.Exit(0)
		}
	}
	ring, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.FileBackend,
		},
		FileDir:                  "~/." + service,
		FilePasswordFunc:         getPassword,
		ServiceName:              service,
		KeychainName:             "login",
		KeychainTrustApplication: true,
		KeychainSynchronizable:   true,
	})
	if err != nil {
		panic(err)
	}
	keyringUserData, err := ring.Get(entryName)
	if err != nil {
		reader := bufio.NewReader(os.Stdin)
		credentials, err := CredentialsFromReader(reader)
		if err != nil {
			panic(err)
		}
		data, err := json.Marshal(credentials)
		if err != nil {
			panic(err)
		}
		err = ring.Set(keyring.Item{
			Key:         entryName,
			Label:       "AnyConnect Helper Credentials",
			Description: "Securely stored credentials for your AnyConnect user",
			Data:        data,
		})
		if err != nil {
			panic(err)
		}
		keyringUserData, _ = ring.Get(entryName)
	}
	credentialData := CredentialData{}
	_ = json.Unmarshal(keyringUserData.Data, &credentialData)

	anyConnectPath, err := getAnyConnectPath()
	if err != nil {
		panic(err)
	}
	anyConnectCommand := exec.Command(anyConnectPath, "-s", "connect", credentialData.VpnHost)
	anyConnectCommand.Stdin = strings.NewReader(credentialData.Group + "\n" + credentialData.Username + "\n" + credentialData.Password + "\n2\ny")
	anyConnectCommand.Stdout = os.Stdout
	err = anyConnectCommand.Run()
	if err != nil {
		panic(err)
	}
}

func getArgs(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	return items[1:]
}

func disconnect() {
	anyConnectPath, err := getAnyConnectPath()
	if err != nil {
		panic(err)
	}
	
	cmdErrChan := make(chan error, 1)
	disconnectCommand := exec.Command(anyConnectPath, "disconnect")
	disconnectCommand.Stdout = os.Stdout
	disconnectCommand.Stdin = os.Stdin
	go func() {
		cmdErrChan <- disconnectCommand.Run()
	}()
	select {
	case _ = <-cmdErrChan:
	case <-time.After(time.Second * 5):
		killCmd := exec.Command("bash", "-c", `sudo pkill -9 -f cisco`)
		killCmd.Stdout = os.Stdout
		killCmd.Stderr = os.Stderr
		_ = killCmd.Run()
	}
}

func CredentialsFromReader(reader *bufio.Reader) (CredentialData, error) {
	fmt.Print("Enter AnyConnect Vpn Host: ")
	vpnHost, _ := reader.ReadString('\n')
	vpnHost = strings.TrimSpace(vpnHost)

	fmt.Print("Enter Group: ")
	group, _ := reader.ReadString('\n')
	group = strings.TrimSpace(group)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	password, err := getPassword("Enter Password: ")
	if err != nil {
		return CredentialData{}, fmt.Errorf("something occurred while entering password: %v", err)
	}

	credentialData := CredentialData{
		VpnHost:  vpnHost,
		Group:    group,
		Username: username,
		Password: password,
	}
	return credentialData, nil
}

func getPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytes, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}
