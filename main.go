package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

const keyServerAddrs string = "serverAddrs"

func ErrorChecker(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	textFileName := "text.txt"

	hasText := r.URL.Query().Has("text")
	text := ""
	if hasText {
		text = r.URL.Query().Get("text")
	} else {
		text = "default text"
	}

	//creamos el archivo
	err := os.WriteFile(textFileName, []byte(text), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("file created successfully")

	archive, err := os.Create("archive.zip")
	ErrorChecker(err)
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	//abrimos el archivo
	file, err := os.Open(textFileName)
	ErrorChecker(err)
	defer file.Close()

	w1, err := zipWriter.Create("content/text.txt")
	ErrorChecker(err)

	if _, err := io.Copy(w1, file); err != nil {
		panic(err)
	}
	defer zipWriter.Close()
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=archive.zip")

	http.ServeFile(w, r, "archive.zip")

}

func getHello(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("%s: got /hello request\n", ctx.Value(keyServerAddrs))
	io.WriteString(w, "Hello, HHTP!\n")
}

func server() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)

	ctx := context.Background()

	server := &http.Server{
		Addr:    ":3333",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddrs, l.Addr().String())
			return ctx
		},
	}
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
		return err
	}
	return nil
}

func main() {
	err := server()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
