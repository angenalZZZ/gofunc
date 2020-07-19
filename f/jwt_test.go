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
	http.HandleFunc("/token", func(res http.ResponseWriter, req *http.Request) {
		claims := map[string]interface{}{"exp": time.Now().Add(time.Hour * 24).Unix()}
		token, err := f.NewJwtToken(claims)
		if err != nil {
			http.Error(res, "Internal Server Error", 500)
			return
		}
		res.Header().Add("Authorization", "Bearer "+token)
		fmt.Fprintf(res, "%s", f.EncodedJson(struct {
			Token string `json:"token"`
			Exp   int    `json:"exp"`
		}{token, 3600 * 24}))
		//res.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/auth", func(res http.ResponseWriter, req *http.Request) {
		var userToken string
		if reqVal := req.Header.Get("Authorization"); reqVal != "" {
			userToken = strings.Split(reqVal, " ")[1]
		} else if reqVal := req.URL.Query().Get("token"); reqVal != "" {
			userToken = reqVal
		}
		if m, ok := f.IsJwtToken(userToken); ok {
			fmt.Fprintf(res, "%s", f.EncodedMap(m))
		} else {
			http.Error(res, "Unauthorized", 401)
		}
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		t.Fatal(err)
	}
}
