package f

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

// Jwt Sign Method.
type JwtSign struct {
	Key string
}

// Gen generate jwt token.
func (j *JwtSign) Gen(payload map[string]interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))
	tokenString, err := token.SignedString([]byte(j.Key))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Parse a JWT token.
func (j *JwtSign) Parse(tokenString string) (map[string]interface{}, error) {
	t, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.Key), nil
	})

	if err != nil {
		return nil, err
	}

	payload, ok := t.Claims.(jwt.MapClaims)
	if ok && t.Valid {
		return payload, nil
	}
	return nil, err
}
