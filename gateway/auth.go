package gateway

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserRole string

const (
	RoleViewer   UserRole = "viewer"
	RoleOperator UserRole = "operator"
	RoleAdmin    UserRole = "admin"
)

type UserIdentity struct {
	ID   string
	Role UserRole
}

type CustomClaims struct {
	UserID string   `json:"user_id"`
	Role   UserRole `json:"role"`
	jwt.RegisteredClaims
}

func (m *Middleware) ValidateToken(tokenStr string) (*UserIdentity, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return m.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return &UserIdentity{
		ID:   claims.UserID,
		Role: claims.Role,
	}, nil
}

func (m *Middleware) GenerateToken(userID string, role UserRole) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.jwtSecret)
}
