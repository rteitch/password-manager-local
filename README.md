# 🔐 Go Password Vault

A secure web-based password manager built with Go. This app allows users to register and log in securely, then store, edit, and retrieve account credentials encrypted using AES-GCM + Argon2.

## ✨ Features

- 🔒 **Secure Vault**: AES-GCM encryption with Argon2 key derivation from the master password.
- 👥 **User System**: Register, login, logout, and change password.
- 📚 **Vault CRUD**: Add, edit, delete account credentials.
- 🗂️ **Account Categorization**: Add notes and group services.
- 💾 **Encrypted Storage**: Each user has a personal, encrypted vault file.
- 🔌 **REST API**: Built with Go's `net/http` package.

## 🚀 Getting Started

### 1. Clone the Project

```bash
git clone https://github.com/rteitch/go-password-vault.git
cd go-password-vault
```

### 2. Set Environment Variables

```bash
export SESSION_KEY=your-secret-key
export PORT=8080
```

> `SESSION_KEY` is used for securing sessions (cookies). Don't use the default in production.

### 3. Run the App

```bash
go run .
```

Visit `http://localhost:8080`

## 🔧 Project Structure

```
├── vault.go          # Vault encryption, storage, account model
├── user.go           # User registration/login, password hashing
├── server.go         # HTTP routes, middleware, session handling
├── static/           # HTML/CSS/JS frontend files
├── users.json        # Registered users
├── user_vaults/      # Encrypted vault files
```

## 📡 API Endpoints

| Method | Endpoint        | Auth | Description                     |
|--------|------------------|------|---------------------------------|
| POST   | /register        | ❌   | Register new user               |
| POST   | /login           | ❌   | Login and start session         |
| POST   | /logout          | ✅   | Logout                          |
| GET    | /list            | ✅   | List all stored accounts        |
| POST   | /add             | ✅   | Add a new account               |
| POST   | /update          | ✅   | Update account by index         |
| POST   | /delete          | ✅   | Delete account by index         |
| POST   | /reset-password  | ✅   | Change master password          |
| GET    | /session-check   | ✅   | Check if session is active      |

## 🔄 Example API JSON Formats

### `/register`
```json
{
  "username": "john",
  "email": "john@example.com",
  "password": "secret123"
}
```

### `/login`
```json
{
  "username": "john",
  "password": "secret123"
}
```

### `/add`
```json
{
  "service": "Gmail",
  "username": "john@gmail.com",
  "password": "mygmailpassword",
  "category": "Email",
  "notes": "Main Gmail account"
}
```

### `/update`
```json
{
  "index": 0,
  "service": "Gmail",
  "username": "john@gmail.com",
  "password": "newpassword123",
  "category": "Email",
  "notes": "Updated password"
}
```

### `/delete`
```json
{
  "index": 0
}
```

### `/reset-password`
```json
{
  "oldPassword": "secret123",
  "newPassword": "newSecret456"
}
```

## 🖼️ Frontend (Static Directory)

Minimal `index.html` example:

```html
<!DOCTYPE html>
<html>
<head><title>Password Vault</title></head>
<body>
  <h1>Welcome to Go Vault</h1>
  <p>Use a REST client (like Postman) to interact with the backend.</p>
</body>
</html>
```

## 🐳 Run with Docker

### Dockerfile

```Dockerfile
FROM golang:1.21-alpine

WORKDIR /app
COPY . .

RUN go build -o vault .

EXPOSE 8080
CMD ["./vault"]
```

### Build & Run

```bash
docker build -t go-vault .
docker run -p 8080:8080 -e SESSION_KEY=mysecret go-vault
```

## 🛡️ Security Notes

- Don't expose `users.json` or `user_vaults/` in production.
- Always use HTTPS in production.
- Change the default session key.

## 📄 License

MIT © 2025 YourName
