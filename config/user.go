package config

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp/totp"
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

func (u *User) ValidOTP(code string, prod bool) bool {
	if prod {
		return totp.Validate(code, u.OTPSecret)
	}
	return hotp.Validate(code, 1, u.OTPSecret) // for tests
}

// passApp methods

func (u *User) AddPassApp(pass string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err == nil {
		u.PassAppBcrypt = append(u.PassAppBcrypt, hex.EncodeToString(hashedPassword))
	}
}

func (u *User) DelPassApp(k int) {
	if k < len(u.PassAppBcrypt) && k >= 0 {
		u.PassAppBcrypt = append(u.PassAppBcrypt[:k], u.PassAppBcrypt[k+1:]...)
	}
}
