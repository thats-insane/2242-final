package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/thats-insane/2242-FINAL/example_5/cookies"
)

var secretKey []byte

type User struct {
	Name string
	Age  int
}

func main() {
	gob.Register(&User{})

	var err error

	secretKey, err = hex.DecodeString("13d6b4dff8f84a10851021ec8608f814570d562c92fe6b5ec4c9f595bcb3234b")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/set", setCookieHandler)
	mux.HandleFunc("/get", getCookieHandler)

	log.Print("Listening...")
	err = http.ListenAndServe(":3000", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func setCookieHandler(w http.ResponseWriter, r *http.Request) {
	user := User{
		Name: "Alice",
		Age:  21,
	}

	var buf bytes.Buffer

	err := gob.NewEncoder(&buf).Encode(&user)
	if err != nil {
		log.Print(err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "example",
		Value:    buf.String(),
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	err = cookies.WriteEncrypted(w, cookie, secretKey)
	if err != nil {
		log.Print(err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("cookie set!"))
}

func getCookieHandler(w http.ResponseWriter, r *http.Request) {
	gobEncodedValue, err := cookies.ReadEncrypted(r, "example", secretKey)
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			http.Error(w, "cookie not found", http.StatusBadRequest)
		case errors.Is(err, cookies.ErrInvalidValue):
			http.Error(w, "invalid cookie", http.StatusBadRequest)
		default:
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	var user User

	reader := strings.NewReader(gobEncodedValue)

	if err := gob.NewDecoder(reader).Decode(&user); err != nil {
		log.Println(err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Name: %q\n", user.Name)
	fmt.Fprintf(w, "Age: %d\n", user.Age)
}
