package main

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

// GetYubiKey requests a one time password on command line from
// the user and returns it - the one time password is created
// by pressing the yubikey button
func GetYubiKey() (string, error) {
	fmt.Printf("Press yubikey button: ")
	if pw, err := terminal.ReadPassword(0); err != nil {
		return "", fmt.Errorf("GetYubiKey(): %s", err.Error())
	} else {
		return string(pw), nil
	}
	return "", nil
}

func GetYubiKeyOrExit() string {
	key, err := GetYubiKey()
	if err != nil {
		fmt.Printf("Error reading in yubikey password from stdin: \n", err)
		os.Exit(1)
	}
	return key
}
