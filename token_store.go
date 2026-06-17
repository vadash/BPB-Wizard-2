package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/99designs/keyring"
	"golang.org/x/oauth2"
)

const (
	tokenStoreService = "\u0042\u0050\u0042-Wizard"
	tokenStoreKey     = "cloudflare-oauth-token"
)

type tokenStore struct{}

func newTokenStore() tokenStore {
	return tokenStore{}
}

func (s tokenStore) Load() (*oauth2.Token, error) {
	if token, err := s.loadKeyring(); err == nil {
		return token, nil
	}

	return s.loadFile()
}

func (s tokenStore) Save(token *oauth2.Token) error {
	if err := s.saveKeyring(token); err == nil {
		return nil
	}

	return s.saveFile(token)
}

func (s tokenStore) Delete() error {
	_ = s.deleteKeyring()
	_ = os.Remove(tokenFilePath())
	return nil
}

func (s tokenStore) loadKeyring() (*oauth2.Token, error) {
	ring, err := openTokenKeyring()
	if err != nil {
		return nil, err
	}

	item, err := ring.Get(tokenStoreKey)
	if err != nil {
		return nil, err
	}

	return decodeToken(item.Data)
}

func (s tokenStore) saveKeyring(token *oauth2.Token) error {
	ring, err := openTokenKeyring()
	if err != nil {
		return err
	}

	data, err := encodeToken(token)
	if err != nil {
		return err
	}

	return ring.Set(keyring.Item{
		Key:         tokenStoreKey,
		Data:        data,
		Label:       "Cloudflare OAuth token",
		Description: "OAuth token used by wizard to access Cloudflare",
	})
}

func (s tokenStore) deleteKeyring() error {
	ring, err := openTokenKeyring()
	if err != nil {
		return err
	}

	if err := ring.Remove(tokenStoreKey); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return err
	}

	return nil
}

func (s tokenStore) loadFile() (*oauth2.Token, error) {
	data, err := os.ReadFile(tokenFilePath())
	if err != nil {
		return nil, err
	}

	return decodeToken(data)
}

func (s tokenStore) saveFile(token *oauth2.Token) error {
	data, err := encodeToken(token)
	if err != nil {
		return err
	}

	path := tokenFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}

	return os.Chmod(path, 0600)
}

func openTokenKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: tokenStoreService,
		AllowedBackends: []keyring.BackendType{
			keyring.WinCredBackend,
			keyring.KeychainBackend,
			keyring.SecretServiceBackend,
			keyring.KWalletBackend,
			keyring.KeyCtlBackend,
			keyring.PassBackend,
		},
		KWalletAppID:  tokenStoreService,
		KWalletFolder: tokenStoreService,
		WinCredPrefix: tokenStoreService,
		KeychainName:  "login",
		KeyCtlScope:   "user",
		KeyCtlPerm:    0x3f000000,
		PassPrefix:    tokenStoreService,
		PassCmd:       "pass",
		PassDir:       "",
	})
}

func tokenFilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		if home, homeErr := os.UserHomeDir(); homeErr == nil {
			dir = filepath.Join(home, ".config")
		} else {
			dir = "."
		}
	}

	return filepath.Join(dir, "\u0042\u0050\u0042-Wizard", "oauth-token.json")
}

func encodeToken(token *oauth2.Token) ([]byte, error) {
	return json.Marshal(token)
}

func decodeToken(data []byte) (*oauth2.Token, error) {
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}
