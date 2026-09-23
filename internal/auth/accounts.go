// EXPERIMENTAL: Multi-account profile management - handles saving, switching, and migrating Spotify user credentials.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"spotumn/internal/config"
)

const MaxAccounts = 4

type Account struct {
	ID          string        `json:"id"`
	DisplayName string        `json:"display_name"`
	Token       *oauth2.Token `json:"token"`
	CreatedAt   time.Time     `json:"created_at"`
}

type accountStore struct {
	Accounts    []Account `json:"accounts"`
	ActiveIndex int       `json:"active_index"`
}

type AccountManager struct {
	file string
	mu   sync.RWMutex
	data accountStore
}

func NewAccountManager() *AccountManager {
	mgr := &AccountManager{
		file: filepath.Join(config.GetDir(), "accounts.json"),
	}
	_ = mgr.Load()
	return mgr
}

func (m *AccountManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.file)
	if err == nil {
		var store accountStore
		if err := json.Unmarshal(data, &store); err == nil {
			m.data = store
			if m.data.ActiveIndex < 0 || m.data.ActiveIndex >= len(m.data.Accounts) {
				m.data.ActiveIndex = 0
			}
			return nil
		}
	}

	// seamlessly migrate legacy single-user credentials into multi-account format
	legacyFile := filepath.Join(config.GetDir(), "credentials.json")
	if legData, err := os.ReadFile(legacyFile); err == nil {
		var stored StoredCredentials
		if err := json.Unmarshal(legData, &stored); err == nil && stored.Valid() && stored.ClientID == config.SpotifyClientID {
			m.data.Accounts = []Account{
				{
					ID:          "default",
					DisplayName: "Spotify User",
					Token:       &stored.Token,
					CreatedAt:   time.Now(),
				},
			}
			m.data.ActiveIndex = 0
			_ = m.saveLocked()
		}
	}

	return nil
}

func (m *AccountManager) saveLocked() error {
	bytes, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.file, bytes, 0600)
}

func (m *AccountManager) GetAccounts() []Account {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]Account, len(m.data.Accounts))
	copy(res, m.data.Accounts)
	return res
}

func (m *AccountManager) ListAccounts() []Account {
	return m.GetAccounts()
}

func (m *AccountManager) GetActiveIndex() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data.ActiveIndex
}

func (m *AccountManager) ActiveIndex() int {
	return m.GetActiveIndex()
}

func (m *AccountManager) GetActiveAccount() *Account {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.data.Accounts) == 0 || m.data.ActiveIndex < 0 || m.data.ActiveIndex >= len(m.data.Accounts) {
		return nil
	}
	acc := m.data.Accounts[m.data.ActiveIndex]
	return &acc
}

func (m *AccountManager) AddAccount(userID, displayName string, token *oauth2.Token) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, acc := range m.data.Accounts {
		if acc.ID == userID && userID != "" && userID != "default" {
			m.data.Accounts[i].Token = token
			if displayName != "" {
				m.data.Accounts[i].DisplayName = displayName
			}
			m.data.ActiveIndex = i
			_ = m.syncActiveTokenLocked()
			return i, m.saveLocked()
		}
	}

	if len(m.data.Accounts) >= MaxAccounts {
		return -1, fmt.Errorf("maximum of %d accounts reached", MaxAccounts)
	}

	if displayName == "" {
		displayName = fmt.Sprintf("Account %d", len(m.data.Accounts)+1)
	}

	newAcc := Account{
		ID:          userID,
		DisplayName: displayName,
		Token:       token,
		CreatedAt:   time.Now(),
	}

	m.data.Accounts = append(m.data.Accounts, newAcc)
	m.data.ActiveIndex = len(m.data.Accounts) - 1
	_ = m.syncActiveTokenLocked()
	return m.data.ActiveIndex, m.saveLocked()
}

func (m *AccountManager) SwitchAccount(idx int) (*Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if idx < 0 || idx >= len(m.data.Accounts) {
		return nil, errors.New("account index out of range")
	}

	m.data.ActiveIndex = idx
	_ = m.syncActiveTokenLocked()
	_ = m.saveLocked()

	acc := m.data.Accounts[idx]
	return &acc, nil
}

func (m *AccountManager) syncActiveTokenLocked() error {
	if len(m.data.Accounts) == 0 || m.data.ActiveIndex < 0 || m.data.ActiveIndex >= len(m.data.Accounts) {
		return nil
	}
	active := m.data.Accounts[m.data.ActiveIndex]
	if active.Token == nil {
		return nil
	}

	stored := StoredCredentials{
		Token:    *active.Token,
		ClientID: config.SpotifyClientID,
	}
	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}
	credFile := filepath.Join(config.GetDir(), "credentials.json")
	return os.WriteFile(credFile, data, 0600)
}
