package f_test

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/f"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestJwtAuth(t *testing.T) {
	http.HandleFunc("/auth/new", func(res http.ResponseWriter, req *http.Request) {
		claims := map[string]interface{}{"exp": time.Now().Add(time.Hour * 24).Unix()}
		token, err := f.NewJwtToken(claims)
		if err != nil {
			http.Error(res, "Error", 500)
			return
		}
		res.Header().Add("Authorization", "Bearer "+token)
		res.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/auth", func(res http.ResponseWriter, req *http.Request) {
		userToken := strings.Split(req.Header.Get("Authorization"), " ")[1]
		if token, ok := f.IsJwtToken(userToken); ok {
			fmt.Fprintf(res, "%v", token)
		} else {
			http.Error(res, "Error", 401)
		}
	})

	t.Fatal(http.ListenAndServe(":8080", nil))
}
