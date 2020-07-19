package f

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	json "github.com/json-iterator/go"
	"strings"
)

var (
	JwtDefaultKey    = []byte("HGJ766GR")
	jwtDefaultHeader = jwt1header{Typ: "JWT", Alg: "HS256"}
	jwtDefaultClaims = []string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti"}
	jwtEncodeString  = base64.RawURLEncoding.EncodeToString
	jwtDecodeString  = base64.RawURLEncoding.DecodeString
	jwtEncodeJson    = json.Marshal
	jwtDecodeJson    = json.Unmarshal
)

// NewJwtToken returns a token (string) and error.
// The token is a fully qualified JWT to be sent to a client via HTTP Header or other method.
// Error returned will be from the jwtNewEncoded unexported function.
func NewJwtToken(claims map[string]interface{}) (string, error) {
	enc, err := jwt1NewEncoded(claims)
	if err != nil {
		return "", err
	}
	return enc.token, nil
}

// IsJwtToken returns a bool indicating whether a token (string) provided has been signed by our server.
// If true, the client is authenticated and may proceed.
func IsJwtToken(token string) (map[string]interface{}, bool) {
	var (
		err error
		dec jwt1decoded
	)

	// decode the token
	if dec, err = jwt1NewDecoded(token); err != nil {
		// may want to log some error here so we have visibility
		// intentionally simplifying return type to bool for ease
		// of use in API. Caller should only do `if auth.Passes(str) {}`
		return nil, false
	}

	// base64 decode payload
	var payload []byte
	if payload, err = jwtDecodeString(dec.payload); err != nil {
		return nil, false
	}
	dst := map[string]interface{}{}
	if err = jwtDecodeJson(payload, &dst); err != nil {
		return nil, false
	}
	if signed, err := dec.sign(); err != nil || signed.token() != token {
		return nil, false
	}
	return dst, true
}

func jwt1NewEncoded(claims map[string]interface{}) (jwt1encoded, error) {
	jwt1header, err := jwtEncodeJson(jwtDefaultHeader)
	if err != nil {
		return jwt1encoded{}, err
	}

	for _, claim := range jwtDefaultClaims {
		if _, ok := claims[claim]; !ok {
			claims[claim] = nil
		}
	}

	payload, err := jwtEncodeJson(claims)
	if err != nil {
		return jwt1encoded{}, err
	}

	d := jwt1decoded{jwt1header: string(jwt1header), payload: string(payload)}
	d.jwt1header = jwtEncodeString([]byte(d.jwt1header))
	d.payload = jwtEncodeString([]byte(d.payload))
	signed, err := d.sign()
	if err != nil {
		return jwt1encoded{}, err
	}
	return jwt1encoded{token: signed.token()}, nil
}

func jwt1NewDecoded(token string) (jwt1decoded, error) {
	e := jwt1encoded{token: token}
	d, err := e.parseToken()
	if err != nil {
		return d, nil
	}
	return d, nil
}

type jwt1header struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
}

type jwt1encoded struct {
	token string
}

type jwt1decoded struct {
	jwt1header string
	payload    string
}

type jwt1signedDecoded struct {
	jwt1decoded
	signature string
}

func (s jwt1signedDecoded) token() string {
	return fmt.Sprintf("%s.%s.%s", s.jwt1header, s.payload, s.signature)
}

func (d *jwt1decoded) sign() (jwt1signedDecoded, error) {
	if d.jwt1header == "" || d.payload == "" {
		return jwt1signedDecoded{}, errors.New("missing jwt1header or payload on Decoded")
	}

	hashed := hmac.New(sha256.New, JwtDefaultKey)
	unsigned := strings.Join([]string{d.jwt1header, d.payload}, ".")
	_, err := hashed.Write([]byte(unsigned))
	if err != nil {
		return jwt1signedDecoded{}, err
	}

	signed := jwt1signedDecoded{jwt1decoded: *d}
	signed.signature = jwtEncodeString(hashed.Sum(nil))

	return signed, nil
}

func (e jwt1encoded) parseToken() (jwt1decoded, error) {
	parts := strings.Split(e.token, ".")
	if len(parts) != 3 {
		return jwt1decoded{}, errors.New("error: incorrect # of results from string parsing")
	}

	d := jwt1decoded{
		jwt1header: parts[0],
		payload:    parts[1],
	}
	return d, nil
}
