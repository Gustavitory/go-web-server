package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

const keyServerAddrs string = "serverAddrs"
const zipFileName string = "archive.zip"
const textFileName string = "text.txt"

func createRootFile(text string) (file *os.File, e error) {
	err := os.WriteFile(textFileName, []byte(text), 0644)
	if err != nil {
		return nil, err
	}
	fmt.Printf("file created successfully")

	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return nil, err
	}
	fmt.Printf("zip created successfully")
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)

	textFile, err := os.Open(textFileName)
	if err != nil {
		return nil, err
	}
	fmt.Printf("textFile is open")
	defer textFile.Close()

	textFileInsideZip, err := zipWriter.Create("content/text.txt")
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(textFileInsideZip, textFile); err != nil {
		return nil, err
	}
	zipWriter.Close()
	return zipFile, nil
}

func getRoot(txtContent string) (f string, e error) {
	file, err := createRootFile(txtContent)
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}

func getRootHandler(w http.ResponseWriter, r *http.Request) {
	text := ""
	hasText := r.URL.Query().Has("text")
	if hasText {
		text = r.URL.Query().Get("text")
	} else {
		text = "default text"
	}
	file, err := getRoot(text)
	if err != nil {
		fmt.Printf("Error creating file.")
		http.Error(w, "Error creating file.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+zipFileName)

	http.ServeFile(w, r, file)
}

func getHello(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("%s: got /hello request\n", ctx.Value(keyServerAddrs))
	io.WriteString(w, "Hello, HHTP!\n")
}

func server() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getRootHandler)
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
		return nil
	} else if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
		return err
	}
	return nil
}

func main() {
	err := server()
	if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
