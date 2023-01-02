package config

import (
	"crypto/rand"
	"strings"

	"github.com/99designs/keyring"
	vault "github.com/sosedoff/ansible-vault-go"
)

const (
	originVaultHeader = "$ANSIBLE_VAULT;1.1;AES256"
	vaultHeader       = "$LEETGO_VAULT;1.1;AES256"
	serviceName       = CmdName
)

// randomSecret generates a secure random string.
func randomSecret() (string, error) {
	buf := make([]byte, 64)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func getEncryptKey(service string) (string, error) {
	ring, err := keyring.Open(
		keyring.Config{
			ServiceName: service,
		},
	)
	if err != nil {
		return "", err
	}
	pw, err := ring.Get("password")
	if err == nil {
		return string(pw.Data), nil
	}
	if err != keyring.ErrKeyNotFound {
		return "", err
	}
	passwd, err := randomSecret()
	if err != nil {
		return "", err
	}
	err = ring.Set(
		keyring.Item{
			Key:         "password",
			Data:        []byte(passwd),
			Description: "Password to encrypt/decrypt sensitive data",
		},
	)
	if err != nil {
		return "", err
	}
	return passwd, nil
}

func Encrypt(in string) (string, error) {
	passwd, err := getEncryptKey(serviceName)
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
	passwd, err := getEncryptKey(serviceName)
	if err != nil {
		return "", err
	}
	in = originVaultHeader + strings.TrimPrefix(in, vaultHeader)
	return vault.Decrypt(in, passwd)
}
