package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const TokenIssuer = "auth-login"

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	{
		if err != nil {
			return "", err
		}
	}

	return string(hash), nil
}

func CheckPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func NewToken(userID int64, role string, secret string) (string, time.Time, error) {
	var tokenID [32]byte

	_, err := rand.Read(tokenID[:])
	{
		if err != nil {
			return "", time.Time{}, err
		}
	}

	now := time.Now().UTC().Truncate(time.Second)

	expiresAt := now.Add(24 * time.Hour)

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    TokenIssuer,
			Subject:   strconv.FormatInt(userID, 10),
			ID:        hex.EncodeToString(tokenID[:]),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	{
		if err != nil {
			return "", time.Time{}, err
		}
	}

	return token, expiresAt, nil
}

func ParseToken(raw string, secret string) (*Claims, error) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(TokenIssuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt())

	if err != nil {
		return nil, err
	}

	return claims, nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}
