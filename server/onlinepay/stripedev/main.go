package main

// Spin off a server and listen for events send by Stripe CLI

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type stripePayload struct {
	Payload   string
	Signature string
}

func main() {
	// cmd := exec.Command("stripe", "listen", "-l", "--all-snapshot", "--forward-to", "localhost:1323/stripe_test")
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// err := cmd.Run()
	// fmt.Println(err)

	http.HandleFunc("/stripe_test", func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("Stripe-Signature")
		b, _ := io.ReadAll(r.Body)

		fmt.Println("received event")
		// if !bytes.Contains(b, []byte("checkout.session.completed")) {
		// 	return
		// }

		t := time.Now().Second()
		out, _ := json.Marshal(stripePayload{string(b), signature})
		os.WriteFile(fmt.Sprintf("test_%d.json", t), out, os.ModePerm)

		fmt.Println(string(b))
	})
	fmt.Println("Listening...")
	log.Fatal(http.ListenAndServe(":1324", nil))
}
