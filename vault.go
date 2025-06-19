package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/argon2"
)

type Account struct {
	Service   string    `json:"service"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Category  string    `json:"category"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"createdAt"`
}

type Vault struct {
	Accounts []Account
	key      []byte
	userID   string
	um       *UserManager
}

const (
	saltSize  = 16
	keySize   = 32
	nonceSize = 12
)

func NewVault(userID string, um *UserManager) *Vault {
	return &Vault{
		Accounts: make([]Account, 0),
		key:      nil,
		userID:   userID,
		um:       um,
	}
}

func (v *Vault) Add(service, username, password, category, notes string) {
	v.Accounts = append(v.Accounts, Account{
		Service:   service,
		Username:  username,
		Password:  password,
		Category:  category,
		Notes:     notes,
		CreatedAt: time.Now(),
	})
}

// PERBAIKAN: Fungsi baru untuk mengubah akun yang ada
func (v *Vault) Update(index int, service, username, password, category, notes string) bool {
	if index < 0 || index >= len(v.Accounts) {
		return false // Index di luar jangkauan
	}
	v.Accounts[index].Service = service
	v.Accounts[index].Username = username
	v.Accounts[index].Password = password
	v.Accounts[index].Category = category
	v.Accounts[index].Notes = notes
	return true
}

func (v *Vault) Delete(index int) bool {
	if index < 0 || index >= len(v.Accounts) {
		return false
	}
	v.Accounts = append(v.Accounts[:index], v.Accounts[index+1:]...)
	return true
}

func deriveKey(password string, salt []byte) []byte {
	if salt == nil {
		salt = make([]byte, saltSize)
		if _, err := rand.Read(salt); err != nil {
			panic("Failed to generate salt: " + err.Error())
		}
	}
	return argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, keySize)
}

func (v *Vault) getVaultPath() string {
	return v.um.GetUserVaultPath(v.userID)
}

func (v *Vault) SaveWithPassword(masterPassword string) error {
	data, err := json.Marshal(v.Accounts)
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	key := deriveKey(masterPassword, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, data, nil)
	finalData := append(salt, append(nonce, ciphertext...)...)
	return os.WriteFile(v.getVaultPath(), finalData, 0600)
}

func (v *Vault) LoadWithPassword(masterPassword string) error {
	content, err := os.ReadFile(v.getVaultPath())
	if errors.Is(err, os.ErrNotExist) {
		v.Accounts = make([]Account, 0)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read vault file: %w", err)
	}
	if len(content) < saltSize+nonceSize {
		return errors.New("corrupted vault file: file too short")
	}
	salt := content[:saltSize]
	nonce := content[saltSize : saltSize+nonceSize]
	ciphertext := content[saltSize+nonceSize:]
	key := deriveKey(masterPassword, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return errors.New("invalid master password or corrupted vault")
	}
	err = json.Unmarshal(plaintext, &v.Accounts)
	if err != nil {
		return fmt.Errorf("failed to unmarshal vault data: %w", err)
	}
	return nil
}

// Fungsi lain tidak berubah
