package main

import (
	"flag"
	"fmt"
	"net/http"

	v1 "github.com/ArichiAya/fshare/api/v1"
)

func main() {
	uploadDir := flag.String("upload", "G:/fileShareServer/upload", "上传文件存储目录")
	trashDir := flag.String("trash", "G:/fileShareServer/trash", "回收站目录")
	addr := flag.String("addr", ":8666", "监听地址")
	flag.Parse()

	mux := http.NewServeMux()
	fsServer := v1.NewFSserver(
		v1.WithUploadDir(*uploadDir),
		v1.WithTrashDir(*trashDir),
	)

	mux.HandleFunc("/upload", fsServer.Upload)
	mux.HandleFunc("/download", fsServer.Download)
	mux.HandleFunc("/downloaddir", fsServer.DownloadDir)
	mux.HandleFunc("/listfiles", fsServer.ListFiles)
	mux.HandleFunc("/delete", fsServer.Delete)
	mux.HandleFunc("/restore", fsServer.Restore)
	mux.HandleFunc("/clean", fsServer.Clean)

	srv := http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	fmt.Printf("Files Share Server Listening At %s\n", srv.Addr)
	fmt.Printf("  Upload: %s\n", *uploadDir)
	fmt.Printf("  Trash:  %s\n", *trashDir)
	srv.ListenAndServe()
}
