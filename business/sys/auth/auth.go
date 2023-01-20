// Package auth provides authentication and authorization support.
package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
)

//// ErrForbidden is returned when a auth issue is identified.
//var ErrForbidden = errors.New("attempted action is not allowed")

// KeyLookup declares a method set of behavior for looking up
// private and public keys for JWT use.
type KeyLookup interface {
	PrivateKeyPEM(kid string) (*rsa.PrivateKey, error)
	PublicKeyPEM(kid string) (*rsa.PublicKey, error)
}

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Auth struct {
	activeKID string
	keyLookup KeyLookup
	method    jwt.SigningMethod
	parser    *jwt.Parser
	keyFunc   func(t *jwt.Token) (interface{}, error)
	cache     map[string]string
}

func New(activeKID string, keyLookup KeyLookup) (*Auth, error) {

	// The active KID represents the private key used to signed new tokens
	_, err := keyLookup.PrivateKeyPEM(activeKID)
	if err != nil {
		return nil, errors.New("configuring algorith RS256")
	}

	method := jwt.GetSigningMethod("RS256")
	if method == nil {
		return nil, errors.New("configuring algorith RS256")
	}

	keyFunc := func(t *jwt.Token) (interface{}, error){
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}
		kidID, ok := kid.(string)
		if !ok {
			return keyLookup.PublicKeyPEM(kidID)
		}
		return kidID, nil
	}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))

	a:= Auth{
		activeKID: activeKID,
		keyLookup: keyLookup,
		method: method,
		keyFunc: keyFunc,
		parser: parser,
	}

	return &a, nil
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Auth) GenerateToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = a.activeKID

	privateKey, err := a.keyLookup.PrivateKeyPEM(a.activeKID)
	if err != nil {
		return "", fmt.Errorf("private key: %w", err)
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return str, nil
}
