package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	//"github.com/google/uuid"
	"io"
	"log"
	"os"
	"github.com/golang-jwt/jwt/v4"
	"time"
)
type Claims struct {
	jwt.RegisteredClaims
	Roles []string `json:"roles"`
}

func main() {
	err := genToken()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func genToken() error {

	name:= "zarf\\keys\\54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem"
	file, err := os.Open(name)
	if err != nil {
		return err
	}

	privatePEM, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return fmt.Errorf("reading auth private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privatePEM))
	if err != nil {
		return fmt.Errorf("parsing private pem: %w", err)
	}

	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1234",
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{"ADMIN"},
	}
	method := jwt.GetSigningMethod("RS256")
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"


	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}

	fmt.Println("===== TOKEN BEGIN =====")
	fmt.Println(tokenStr)
	fmt.Println("===== TOKEN END =====")
	fmt.Println("\n")

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Create a file for the public key information in PEM form.
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the public key file.
	if err := pem.Encode(os.Stdout, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, fmt.Errorf("missing key id (kid) in token header")
		}
		kidID, ok := kid.(string)
		if !ok {
			return nil, errors.New("user token key id (kid) must be string")
		}
		fmt.Println("KID",kidID)
		return &privateKey.PublicKey, nil
	}

	var parsedclaims Claims
	parsedToken, err := parser.ParseWithClaims(tokenStr, &parsedclaims, keyFunc)
	if err != nil {
		return fmt.Errorf("parssing token: %w", err)
	}

	if !parsedToken.Valid {
		return errors.New("invalid token")
	}
	fmt.Println("===============")
	fmt.Println("Token Validated")

	return nil
}
