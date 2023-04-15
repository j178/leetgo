package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	vault "github.com/sosedoff/ansible-vault-go"
	"github.com/zalando/go-keyring"

	"github.com/j178/leetgo/constants"
)

const (
	originVaultHeader = "$ANSIBLE_VAULT;1.1;AES256"
	vaultHeader       = "$LEETGO_VAULT;1.1;AES256"
	serviceName       = constants.CmdName
	keyName           = "encryption_key"
)

// randomSecret generates a secure random string.
func randomSecret() (string, error) {
	buf := make([]byte, 64)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// https://github.com/99designs/keyring relies on CGO to use keychain on macOS.
// But it cannot compile successfully on macOS even with CGO enabled.
// Use a simper github.com/zalando/go-keyring instead.

func getEncryptKey() (string, error) {
	pw, err := keyring.Get(serviceName, keyName)
	if err == nil {
		return pw, nil
	}
	if err != keyring.ErrNotFound {
		return "", fmt.Errorf("keyring get error: %w", err)
	}

	// key not found, generate a new one
	passwd, err := randomSecret()
	if err != nil {
		return "", err
	}
	err = keyring.Set(serviceName, keyName, passwd)
	if err != nil {
		return "", err
	}
	return passwd, nil
}

func Encrypt(in string) (string, error) {
	passwd, err := getEncryptKey()
	if err != nil {
		return "", err
	}
	out, err := vault.Encrypt(in, passwd)
	if err != nil {
		return "", err
	}
	return vaultHeader + strings.TrimPrefix(out, originVaultHeader), nil
}

func Decrypt(in string) (string, error) {
	passwd, err := getEncryptKey()
	if err != nil {
		return "", err
	}
	in = originVaultHeader + strings.TrimPrefix(in, vaultHeader)
	return vault.Decrypt(in, passwd)
}
