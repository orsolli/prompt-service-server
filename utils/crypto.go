package utils

import (
    "encoding/base64"
    "crypto/rand"
    "errors"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key-here") // In production, use environment variable

type Claims struct {
    KeyHash string `json:"key_hash"`
    jwt.RegisteredClaims
}

func GenerateCSRFToken() (string, error) {
    // Generate a JWT with a secret key
    now := time.Now()
    claims := &Claims{
        KeyHash: "placeholder", // This would be the user's public key hash
        RegisteredClaims: jwt.RegisteredClaims{
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func VerifyJWT(tokenString string) error {
    // Verify the JWT signature
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    
    if err != nil {
        return err
    }
    
    if !token.Valid {
        return errors.New("invalid token")
    }
    
    return nil
}

func GenerateKeyPair() (string, string, error) {
    // Generate a new keypair
    // This is a simplified version - in reality, you'd use crypto/ed25519
    
    // Generate a random key
    key := make([]byte, 32)
    if _, err := rand.Read(key); err != nil {
        return "", "", err
    }
    
    // Return base64 encoded key
    return base64.StdEncoding.EncodeToString(key), "", nil
}
