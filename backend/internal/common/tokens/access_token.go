package tokens

import (
	"os"
	"strconv"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessTokenClaims struct {
	Role  string `json:"role"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func GetAccessToken(id, email, role string) (string, error) {
	issuer := os.Getenv("JWT_ISSUER")
	audience := os.Getenv("JWT_AUDIENCE")
	jwtSecret := os.Getenv("JWT_SECRET")

	jwtExpiry, err := strconv.ParseInt(os.Getenv("JWT_EXPIRY"), 10, 16)
	if err != nil {
		logger.Logger.Fatal("JWT_EXPIRY must be a number")
	}

	jwtId, err := uuid.NewUUID()
	if err != nil {
		logger.Logger.Error("jwtID not set")
	}

	claims := AccessTokenClaims{
		Role:  role,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtExpiry) * time.Second).UTC()),
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
			ID:        jwtId.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	byteSecret := []byte(jwtSecret)

	return token.SignedString(byteSecret)
}
