package gateway

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	jwtSecret   []byte
	audit       *AuditManager
}

func NewMiddleware(secret string, audit *AuditManager) (*Middleware, error) {
	return &Middleware{
		jwtSecret:   []byte(secret),
		audit:       audit,
	}, nil
}

func (m *Middleware) AuthenticateWS(r *http.Request) (bool, error) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		return false, nil
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return m.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return false, err
	}

	return true, nil
}

func (m *Middleware) LogIO(sessionID, direction, data string) {
	m.audit.Log(sessionID, direction, data)
}

func (m *Middleware) GenerateDevToken() (string, error) {
	return m.GenerateToken("dev-user", RoleOperator)
}
