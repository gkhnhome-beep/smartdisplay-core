package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"smartdisplay-core/internal/logger"
)

// Tüm kullanıcıları döndürür
func LoadAllUsers() ([]User, error) {
	return loadUsers()
}

// Kullanıcı ekler
func AddUser(newUser User) error {
	users, err := loadUsers()
	if err != nil {
		users = []User{}
	}
	// Aynı kullanıcı adı varsa hata
	for _, u := range users {
		if u.Username == newUser.Username {
			return fmt.Errorf("Kullanıcı zaten mevcut")
		}
	}
	users = append(users, newUser)
	return saveUsers(users)
}

// Kullanıcı günceller
func UpdateUser(updated User) error {
	users, err := loadUsers()
	if err != nil {
		return err
	}
	found := false
	for i, u := range users {
		if u.Username == updated.Username {
			users[i] = updated
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Kullanıcı bulunamadı")
	}
	return saveUsers(users)
}

// Kullanıcı siler
func DeleteUser(username string) error {
	users, err := loadUsers()
	if err != nil {
		return err
	}
	idx := -1
	for i, u := range users {
		if u.Username == username {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("Kullanıcı bulunamadı")
	}
	users = append(users[:idx], users[idx+1:]...)
	return saveUsers(users)
}

// Kullanıcıları dosyaya kaydeder
func saveUsers(users []User) error {
	path := "data/users.json"
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(users)
}

// Role represents a user role
type Role string

// AuthContext holds authentication result
type AuthContext struct {
	Role          Role
	Authenticated bool
	PIN           string
}

// ValidatePIN checks the given PIN against users.json and returns AuthContext
func ValidatePIN(pin string) (*AuthContext, error) {
	logger.Info("[AUTH] ValidatePIN: called with pin=" + pin)
	users, err := loadUsers()
	if err != nil {
		logger.Error("[AUTH] ValidatePIN: loadUsers failed: " + err.Error())
		return &AuthContext{Role: Guest, Authenticated: false, PIN: ""}, err
	}
	for _, user := range users {
		logger.Info("[AUTH] ValidatePIN: checking user=" + user.Username + ", pin=" + user.PIN)
		if user.PIN == pin {
			logger.Info("[AUTH] ValidatePIN: PIN match for user=" + user.Username)
			return &AuthContext{
				Role:          user.Role,
				Authenticated: true,
				PIN:           pin,
			}, nil
		}
	}
	logger.Info("[AUTH] ValidatePIN: no match for pin=" + pin)
	return &AuthContext{Role: Guest, Authenticated: false, PIN: ""}, nil
}

const (
	Admin    Role = "admin"
	UserRole Role = "user"
	Guest    Role = "guest"
)

type Permission string

const (
	PermAlarm  Permission = "alarm"
	PermDevice Permission = "device"
	PermGuest  Permission = "guest"
)

var rolePermissions = map[Role][]Permission{
	Admin:    {PermAlarm, PermDevice, PermGuest},
	UserRole: {PermAlarm, PermDevice},
	Guest:    {PermGuest},
}

func HasPermission(role Role, perm Permission) bool {
	perms := rolePermissions[role]
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// AuthContext represents authenticated request context
// FAZ L1: PIN-based authentication

type User struct {
	Username string `json:"username"`
	PIN      string `json:"pin"`
	Role     Role   `json:"role"` // Keep this line as it is
}

func loadUsers() ([]User, error) {
	var paths = []string{"data/users.json", "internal/auth/users.json"}
	var f *os.File
	var err error
	for _, path := range paths {
		logger.Info("[AUTH] loadUsers: trying path=" + path)
		f, err = os.Open(path)
		if err == nil {
			defer f.Close()
			var users []User
			if err := json.NewDecoder(f).Decode(&users); err != nil {
				logger.Error("[AUTH] loadUsers: failed to decode users.json at " + path + ": " + err.Error())
				return nil, err
			}
			logger.Info("[AUTH] loadUsers: loaded " + fmt.Sprintf("%d", len(users)) + " users from " + path)
			return users, nil
		} else {
			logger.Info("[AUTH] loadUsers: not found at " + path)
		}
	}
	// Otomatik admin oluştur
	logger.Info("[AUTH] loadUsers: no users.json found, creating default admin user")
	defaultUser := User{
		Username: "admin",
		PIN:      "1234",
		Role:     Admin,
	}
	_ = saveUsers([]User{defaultUser})
	return []User{defaultUser}, nil
}

// hashPIN creates SHA-256 hash of PIN
