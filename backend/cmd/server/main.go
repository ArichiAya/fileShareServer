package main

import (
	"fmt"
	"net/http"

	v1 "github.com/ArichiAya/fshare/api/v1"
)

func main() {
	mux := http.NewServeMux()
	fsServer := v1.NewFSserver()

	mux.HandleFunc("/upload", fsServer.Upload)
	mux.HandleFunc("/download", fsServer.Download)
	mux.HandleFunc("/listfiles", fsServer.ListFiles)
	mux.HandleFunc("/delete", fsServer.Delete)
	mux.HandleFunc("/restore", fsServer.Restore)
	mux.HandleFunc("/clean", fsServer.Clean)

	srv := http.Server{
		Addr:    ":8666",
		Handler: mux,
	}

	fmt.Println("Files Share Server Listening At ", srv.Addr)
	srv.ListenAndServe()
}
