package config

import (
	"strings"

	vault "github.com/sosedoff/ansible-vault-go"
)

const (
	originVaultHeader = "$ANSIBLE_VAULT;1.1;AES256"
	vaultHeader       = "$LEETGO_VAULT;1.1;AES256"
	password          = "password"
)

func Encrypt(in string) (string, error) {
	out, err := vault.Encrypt(in, password)
	if err != nil {
		return "", err
	}
	return vaultHeader + strings.TrimPrefix(out, originVaultHeader), nil
}

func Decrypt(in string) (string, error) {
	in = originVaultHeader + strings.TrimPrefix(in, vaultHeader)
	return vault.Decrypt(in, password)
}
