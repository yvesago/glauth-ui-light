package config

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// Add User password methods

func (u *User) ValidPass(pass string, allowsha256 bool) bool {
	if allowsha256 {
		if u.ValidSHA256Pass(pass) {
			return true
		}
		return u.ValidBcryptPass(pass)
	}
	return u.ValidBcryptPass(pass)
}

func (u *User) ValidSHA256Pass(pass string) bool {
	if u.PassSHA256 == "" {
		return false
	}

	hashFull := sha256.New()
	hashFull.Write([]byte(pass))
	return u.PassSHA256 == hex.EncodeToString(hashFull.Sum(nil))
}

func (u *User) SetSHA256Pass(pass string) {
	hashFull := sha256.New()
	hashFull.Write([]byte(pass))
	u.PassSHA256 = hex.EncodeToString(hashFull.Sum(nil))
}

func (u *User) ValidBcryptPass(pass string) bool {
	if u.PassBcrypt == "" {
		return false
	}

	decoded, _ := hex.DecodeString(u.PassBcrypt)
	return bcrypt.CompareHashAndPassword(decoded, []byte(pass)) == nil
}

func (u *User) SetBcryptPass(pass string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err == nil {
		u.PassBcrypt = hex.EncodeToString(hashedPassword)
	}
}
