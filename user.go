package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`
}

type UserManager struct {
	usersFile string
	usersDir  string
}

const (
	usersFile = "users.json"
	usersDir  = "user_vaults"
)

func NewUserManager() *UserManager {
	um := &UserManager{
		usersFile: usersFile,
		usersDir:  usersDir,
	}

	// Buat direktori untuk vault user jika belum ada
	if err := os.MkdirAll(usersDir, 0700); err != nil {
		panic("Gagal membuat direktori user vaults: " + err.Error())
	}

	return um
}

// Fungsi untuk generate ID unik untuk user
func generateUserID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("Gagal generate user ID: " + err.Error())
	}
	return fmt.Sprintf("%x", bytes)
}

// Hash password menggunakan bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Verifikasi password
func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Load semua users dari file
func (um *UserManager) loadUsers() ([]User, error) {
	var users []User

	content, err := os.ReadFile(um.usersFile)
	if os.IsNotExist(err) {
		return users, nil // File belum ada, return empty slice
	}
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file users: %w", err)
	}

	if len(content) == 0 {
		return users, nil // File kosong
	}

	err = json.Unmarshal(content, &users)
	if err != nil {
		return nil, fmt.Errorf("gagal parse users JSON: %w", err)
	}

	return users, nil
}

// Simpan users ke file
func (um *UserManager) saveUsers(users []User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal marshal users: %w", err)
	}

	return os.WriteFile(um.usersFile, data, 0600)
}

// Register user baru
func (um *UserManager) RegisterUser(username, email, password string) (*User, error) {
	// Validasi input
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if len(username) < 3 {
		return nil, errors.New("username minimal 3 karakter")
	}
	if len(password) < 6 {
		return nil, errors.New("password minimal 6 karakter")
	}
	if !strings.Contains(email, "@") {
		return nil, errors.New("format email tidak valid")
	}

	users, err := um.loadUsers()
	if err != nil {
		return nil, err
	}

	// Cek apakah username atau email sudah ada
	for _, user := range users {
		if strings.ToLower(user.Username) == strings.ToLower(username) {
			return nil, errors.New("username sudah digunakan")
		}
		if strings.ToLower(user.Email) == strings.ToLower(email) {
			return nil, errors.New("email sudah digunakan")
		}
	}

	// Hash password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("gagal hash password: %w", err)
	}

	// Buat user baru
	newUser := User{
		ID:           generateUserID(),
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		LastLogin:    time.Time{}, // Zero time, belum pernah login
	}

	// Tambahkan ke list users
	users = append(users, newUser)

	// Simpan ke file
	err = um.saveUsers(users)
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

// Login user
func (um *UserManager) LoginUser(usernameOrEmail, password string) (*User, error) {
	users, err := um.loadUsers()
	if err != nil {
		return nil, err
	}

	usernameOrEmail = strings.TrimSpace(usernameOrEmail)

	// Cari user berdasarkan username atau email
	var foundUser *User
	for i, user := range users {
		if strings.ToLower(user.Username) == strings.ToLower(usernameOrEmail) ||
			strings.ToLower(user.Email) == strings.ToLower(usernameOrEmail) {
			foundUser = &users[i]
			break
		}
	}

	if foundUser == nil {
		return nil, errors.New("user tidak ditemukan")
	}

	// Verifikasi password
	if !verifyPassword(password, foundUser.PasswordHash) {
		return nil, errors.New("password salah")
	}

	// Update last login
	foundUser.LastLogin = time.Now()

	// Simpan perubahan
	err = um.saveUsers(users)
	if err != nil {
		// Login tetap berhasil meskipun gagal update last login
		fmt.Printf("Warning: gagal update last login untuk user %s: %v\n", foundUser.Username, err)
	}

	return foundUser, nil
}

// Get user by ID
func (um *UserManager) GetUserByID(userID string) (*User, error) {
	users, err := um.loadUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.ID == userID {
			return &user, nil
		}
	}

	return nil, errors.New("user tidak ditemukan")
}

// Get vault file path untuk user tertentu
func (um *UserManager) GetUserVaultPath(userID string) string {
	return filepath.Join(um.usersDir, fmt.Sprintf("vault_%s.dat", userID))
}

// Get backup vault file path untuk user tertentu
func (um *UserManager) GetUserBackupVaultPath(userID string) string {
	return filepath.Join(um.usersDir, fmt.Sprintf("vault_%s_backup.dat", userID))
}

// Update password user
func (um *UserManager) UpdateUserPassword(userID, oldPassword, newPassword string) error {
	users, err := um.loadUsers()
	if err != nil {
		return err
	}

	// Cari user
	var userIndex = -1
	for i, user := range users {
		if user.ID == userID {
			userIndex = i
			break
		}
	}

	if userIndex == -1 {
		return errors.New("user tidak ditemukan")
	}

	// Verifikasi password lama
	if !verifyPassword(oldPassword, users[userIndex].PasswordHash) {
		return errors.New("password lama salah")
	}

	// Validasi password baru
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}

	// Hash password baru
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("gagal hash password baru: %w", err)
	}

	// Update password
	users[userIndex].PasswordHash = hashedPassword

	// Simpan perubahan
	return um.saveUsers(users)
}

// List semua users (untuk admin, jangan expose password hash)
func (um *UserManager) ListUsers() ([]User, error) {
	users, err := um.loadUsers()
	if err != nil {
		return nil, err
	}

	// Hapus password hash untuk keamanan
	for i := range users {
		users[i].PasswordHash = ""
	}

	return users, nil
}
