package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/pbkdf2"
)

// Vault provides encryption/decryption for sensitive data
type Vault struct {
	key      []byte
	keyFile  string
	unlocked bool
}

const (
	saltSize   = 32
	keySize    = 32
	iterations = 100000
)

// NewVault creates a new vault instance
func NewVault() *Vault {
	homeDir, _ := os.UserHomeDir()
	return &Vault{
		keyFile: filepath.Join(homeDir, ".config", "systask", ".vault.key"),
	}
}

// IsUnlocked returns whether the vault is unlocked
func (v *Vault) IsUnlocked() bool {
	return v.unlocked && len(v.key) > 0
}

// Unlock unlocks the vault with a password
func (v *Vault) Unlock(password string) error {
	// Read or create salt
	salt, err := v.getSalt()
	if err != nil {
		return err
	}

	// Derive key from password
	v.key = pbkdf2.Key([]byte(password), salt, iterations, keySize, sha256.New)
	v.unlocked = true

	return nil
}

// Lock locks the vault
func (v *Vault) Lock() {
	v.key = nil
	v.unlocked = false
}

// Encrypt encrypts plaintext
func (v *Vault) Encrypt(plaintext string) (string, error) {
	if !v.unlocked {
		return "", fmt.Errorf("vault is locked")
	}

	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext
func (v *Vault) Decrypt(encrypted string) (string, error) {
	if !v.unlocked {
		return "", fmt.Errorf("vault is locked")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// getSalt gets or creates a salt for key derivation
func (v *Vault) getSalt() ([]byte, error) {
	saltFile := v.keyFile + ".salt"

	// Try to read existing salt
	salt, err := os.ReadFile(saltFile)
	if err == nil && len(salt) == saltSize {
		return salt, nil
	}

	// Create new salt
	salt = make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// Ensure directory exists
	dir := filepath.Dir(saltFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	// Save salt
	if err := os.WriteFile(saltFile, salt, 0600); err != nil {
		return nil, err
	}

	return salt, nil
}

// EncryptPassword encrypts a password for storage
func EncryptPassword(password, masterPassword string) (string, error) {
	v := NewVault()
	if err := v.Unlock(masterPassword); err != nil {
		return "", err
	}
	defer v.Lock()
	return v.Encrypt(password)
}

// DecryptPassword decrypts a stored password
func DecryptPassword(encrypted, masterPassword string) (string, error) {
	v := NewVault()
	if err := v.Unlock(masterPassword); err != nil {
		return "", err
	}
	defer v.Lock()
	return v.Decrypt(encrypted)
}

// IsEncrypted checks if a string looks encrypted
func IsEncrypted(s string) bool {
	if len(s) < 20 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
