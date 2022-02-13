package config

import (
	"log"
	"strings"
	"testing"

	"encoding/base32"

	"github.com/pquerna/otp/hotp"

	"github.com/stretchr/testify/assert"
)

var Data Ctmp

func initUsersValues() {
	v1 := User{
		Name:         "user1",
		UIDNumber:    5000,
		PrimaryGroup: 6501,
		PassSHA256:   "6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a",
		OTPSecret:    "3hnvnk4ycv44glzigd6s25j4dougs3rk",
	}
	Data.Users = append(Data.Users, v1)
	v2 := User{
		Name:         "user2",
		UIDNumber:    5001,
		PrimaryGroup: 6504,
		PassBcrypt:   "243261243130244B62463462656F7265504F762E794F324957746D656541326B4B46596275674A79336A476845764B616D65446169784E41384F4432",
	}
	Data.Users = append(Data.Users, v2)
	v3 := User{
		Name:         "user3",
		UIDNumber:    5001,
		PrimaryGroup: 6505,
	}
	Data.Users = append(Data.Users, v3)
}

func resetData() {
	Data.Users = []User{}
}

func TestUserModel(t *testing.T) {
	defer resetData()

	cfg := WebConfig{
		DBfile:  "sample-simple.cfg",
		Debug:   true,
		Verbose: false,
		CfgUsers: CfgUsers{
			Start:         5000,
			GIDAdmin:      6501,
			GIDcanChgPass: 6500,
		},
		PassPolicy: PassPolicy{
			AllowReadSSHA256: true,
		},
	}

	initUsersValues()

	sha256User := Data.Users[0]
	bcryptUser := Data.Users[1]
	nopassUser := Data.Users[2]

	// Test passwords
	log.Println("= Test passwords")

	assert.Equal(t, false, sha256User.ValidPass("badpass", cfg.PassPolicy.AllowReadSSHA256), "unvalid sha256 pass")
	assert.Equal(t, false, sha256User.ValidPass("badpass", false), "unvalid sha256 pass")
	assert.Equal(t, false, bcryptUser.ValidPass("badpass", false), "unvalid sha256 pass")

	assert.Equal(t, true, sha256User.ValidPass("dogood", cfg.PassPolicy.AllowReadSSHA256), "valid sha256 pass")
	assert.Equal(t, true, bcryptUser.ValidPass("dogood", false), "valid bcrypt pass")
	assert.Equal(t, true, bcryptUser.ValidPass("dogood", true), "valid bcrypt pass with sha256 not set")

	assert.Equal(t, false, sha256User.ValidPass("dogood", false), "sha256 pass forbidden")

	assert.Equal(t, false, nopassUser.ValidPass("dogood", cfg.PassPolicy.AllowReadSSHA256), "unvalid user without pass")
	assert.Equal(t, false, nopassUser.ValidPass("dogood", false), "unvalid user without pass")

	// Set passwords
	log.Println("= Set passwords")
	sha256User.SetSHA256Pass("otherpass")
	assert.Equal(t, true, sha256User.ValidPass("otherpass", cfg.PassPolicy.AllowReadSSHA256), "change sha256 pass")

	bcryptUser.SetBcryptPass("otherpass")
	assert.Equal(t, true, bcryptUser.ValidPass("otherpass", false), "change bcrypt pass")

	otpgood := "3hnvnk4ycv44glzigd6s25j4dougs3rk"
	otpgood2 := "gvxdgn3hpfvwu2lhmz3gmm3z"
	otpbad := "810bk3t6mdt5j579m29mjm"
	_, err := base32.StdEncoding.DecodeString(strings.ToUpper(otpgood))
	assert.Equal(t, nil, err, "good otp")
	_, err = base32.StdEncoding.DecodeString(strings.ToUpper(otpgood2))
	assert.Equal(t, nil, err, "good otp2")
	_, err = base32.StdEncoding.DecodeString(strings.ToUpper(otpbad))
	assert.Equal(t, "illegal base32 data at input byte 0", err.Error(), "bad otp")

	passcode, _ := hotp.GenerateCode(otpgood, 1)
	log.Println(passcode)

	valid := hotp.Validate(passcode, 1, otpgood)
	log.Println(valid)
	sha256User.OTPSecret = otpgood
	assert.Equal(t, true, sha256User.ValidOTP(passcode, false), "Valid hotp")
	assert.Equal(t, false, sha256User.ValidOTP(passcode, true), "Bad totp")
}
