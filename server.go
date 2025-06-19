package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/sessions"
)

var userManager *UserManager
var store *sessions.CookieStore

func init() {
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		sessionKey = "default-insecure-key-for-development"
		log.Println("PERINGATAN: Menggunakan kunci sesi default. Atur environment variable SESSION_KEY di produksi!")
	}
	store = sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	userManager = NewUserManager()
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "vault-session")
		if err != nil {
			http.Error(w, "Gagal mengambil sesi", http.StatusInternalServerError)
			return
		}
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Tidak terautentikasi", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	staticDir := "static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		os.Mkdir(staticDir, 0755)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/add", authMiddleware(addHandler))
	http.HandleFunc("/list", authMiddleware(listHandler))
	http.HandleFunc("/delete", authMiddleware(deleteHandler))
	http.HandleFunc("/reset-password", authMiddleware(resetPasswordHandler))
	http.HandleFunc("/session-check", sessionCheckHandler)

	// PERBAIKAN: Menambahkan handler untuk endpoint /update
	http.HandleFunc("/update", authMiddleware(updateHandler))

	log.Printf("Server berjalan di http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Gagal memulai server: %v", err)
	}
}

// PERBAIKAN: Fungsi handler baru untuk mengubah data akun
func updateHandler(w http.ResponseWriter, r *http.Request) {
	vault, masterPassword, err := getVaultFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req struct {
		Index    int    `json:"index"`
		Service  string `json:"service"`
		Username string `json:"username"`
		Password string `json:"password"`
		Category string `json:"category"`
		Notes    string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if !vault.Update(req.Index, req.Service, req.Username, req.Password, req.Category, req.Notes) {
		http.Error(w, "Index tidak valid", http.StatusBadRequest)
		return
	}

	if err := vault.SaveWithPassword(masterPassword); err != nil {
		http.Error(w, "Gagal menyimpan vault setelah memperbarui", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Handler lain tidak berubah...
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	_, err := userManager.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Registrasi berhasil, silakan login.")
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user, err := userManager.LoginUser(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Username atau password salah", http.StatusUnauthorized)
		return
	}
	vault := NewVault(user.ID, userManager)
	err = vault.LoadWithPassword(req.Password)
	if err != nil && err.Error() != "invalid master password or corrupted vault" {
		if _, ok := err.(*os.PathError); !ok {
			http.Error(w, "Gagal memuat vault: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	session, _ := store.Get(r, "vault-session")
	session.Values["authenticated"] = true
	session.Values["userID"] = user.ID
	session.Values["masterPassword"] = req.Password
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Gagal membuat sesi", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "vault-session")
	session.Values["authenticated"] = false
	session.Values["userID"] = nil
	session.Values["masterPassword"] = nil
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Gagal logout", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func getVaultFromSession(r *http.Request) (*Vault, string, error) {
	session, err := store.Get(r, "vault-session")
	if err != nil {
		return nil, "", fmt.Errorf("gagal mendapatkan sesi: %w", err)
	}
	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
		return nil, "", fmt.Errorf("userID tidak ditemukan di sesi")
	}
	masterPassword, ok := session.Values["masterPassword"].(string)
	if !ok || masterPassword == "" {
		return nil, "", fmt.Errorf("masterPassword tidak ditemukan di sesi")
	}
	vault := NewVault(userID, userManager)
	if err := vault.LoadWithPassword(masterPassword); err != nil {
		if !os.IsNotExist(err) && err.Error() != "invalid master password or corrupted vault" {
			return nil, "", fmt.Errorf("gagal memuat vault: %w", err)
		}
	}
	return vault, masterPassword, nil
}
func addHandler(w http.ResponseWriter, r *http.Request) {
	vault, masterPassword, err := getVaultFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req struct {
		Service  string `json:"service"`
		Username string `json:"username"`
		Password string `json:"password"`
		Category string `json:"category"`
		Notes    string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	vault.Add(req.Service, req.Username, req.Password, req.Category, req.Notes)
	if err := vault.SaveWithPassword(masterPassword); err != nil {
		http.Error(w, "Gagal menyimpan vault", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func listHandler(w http.ResponseWriter, r *http.Request) {
	vault, _, err := getVaultFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vault.Accounts)
}
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	vault, masterPassword, err := getVaultFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req struct {
		Index int `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if !vault.Delete(req.Index) {
		http.Error(w, "Index tidak valid", http.StatusBadRequest)
		return
	}
	if err := vault.SaveWithPassword(masterPassword); err != nil {
		http.Error(w, "Gagal menyimpan vault setelah menghapus", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "vault-session")
	userID := session.Values["userID"].(string)
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	vault := NewVault(userID, userManager)
	if err := vault.LoadWithPassword(req.OldPassword); err != nil {
		http.Error(w, "Password lama salah", http.StatusBadRequest)
		return
	}
	if err := userManager.UpdateUserPassword(userID, req.OldPassword, req.NewPassword); err != nil {
		http.Error(w, "Gagal memperbarui password user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := vault.SaveWithPassword(req.NewPassword); err != nil {
		http.Error(w, "Gagal menyimpan ulang vault dengan password baru: "+err.Error(), http.StatusInternalServerError)
		log.Printf("KRITIS: Password user %s diubah, tetapi vault gagal dienkripsi ulang. Diperlukan intervensi manual.", userID)
		return
	}
	session.Values["masterPassword"] = req.NewPassword
	session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}
func sessionCheckHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "vault-session")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
}
