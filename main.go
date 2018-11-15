package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type ZipServer struct {
	FileName   string
	fileSystem map[string]string
	fileList   []string
}

func NewZipServer(file string) (*ZipServer, error) {
	log.Println("Opening:", file)

	var (
		fileSystem = make(map[string]string)
		files      = []string{}
	)

	reader, err := zip.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	for _, f := range reader.File {
		var buf = &bytes.Buffer{}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(buf, rc)
		if err != nil {
			return nil, err
		}
		fileSystem[f.Name] = buf.String()
		files = append(files, f.Name)
	}

	return &ZipServer{
		FileName:   file,
		fileSystem: fileSystem,
		fileList:   files,
	}, nil
}

func (zs ZipServer) List(w http.ResponseWriter, r *http.Request) error {
	for _, f := range zs.fileList {
		fmt.Fprintln(w, f)
	}
	return nil
}

func (zs ZipServer) Serve(w http.ResponseWriter, r *http.Request) error {
	trimed := strings.TrimPrefix(r.RequestURI, "/")
	fmt.Fprintf(w, "%s", zs.fileSystem[trimed])
	return nil
}

type AppHandler func(w http.ResponseWriter, r *http.Request) error

func (ah AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := ah(w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func main() {
	flag.Parse()

	zipServer, err := NewZipServer(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/_site", AppHandler(zipServer.List))
	http.Handle("/", AppHandler(zipServer.Serve))
	err = http.ListenAndServe("localhost:9999", nil)
	if err != nil {
		log.Fatal(err)
	}
}
