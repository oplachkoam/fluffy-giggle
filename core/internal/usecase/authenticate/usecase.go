// Package authenticate checks the single operator account from env.
package authenticate

import (
	"crypto/sha256"
	"crypto/subtle"
)

type Usecase struct {
	login    [32]byte
	password [32]byte
}

func New(login, password string) *Usecase {
	return &Usecase{login: sha256.Sum256([]byte(login)), password: sha256.Sum256([]byte(password))}
}

// Execute compares hashes in constant time, so neither the length nor the prefix
// of the real values can be told from the timing.
func (u *Usecase) Execute(login, password string) bool {
	l := sha256.Sum256([]byte(login))
	p := sha256.Sum256([]byte(password))
	return subtle.ConstantTimeCompare(l[:], u.login[:])&subtle.ConstantTimeCompare(p[:], u.password[:]) == 1
}
