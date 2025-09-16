package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/gliderlabs/ssh"
)

const uploadsBase string = "./uploads/"

func processList(s ssh.Session) error {
	username := filepath.Base(s.User())

	dirPath := filepath.Join(uploadsBase, username)

	dir, err := os.ReadDir(dirPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintln(s, "Create an account first loser!")
			return nil
		} else {
			fmt.Fprintln(s, "Oopsie daisy")
			return nil
		}
	}

	for idx, entry := range dir {
		fmt.Fprintf(s, "%d. %s\n", idx, entry.Name())
	}

	return nil
}

func processUpload(s ssh.Session, filename string) error {

	username := filepath.Base(s.User())
	filename = filepath.Base(filename)

	log.Println("Handling upload operation for", username, ". file: ", filename)

	userDir := filepath.Join(uploadsBase, username)

	err := os.MkdirAll(userDir, 0755)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fp := filepath.Join(userDir, filename)
	fh, err := os.Create(fp)
	if err != nil {
		fmt.Println(err)
		return err
	}

	io.Copy(fh, s)

	return nil
}

func processDownload(s ssh.Session, filename string) error {
	username := filepath.Base(s.User())
	filename = filepath.Base(filename)

	fp := filepath.Join(uploadsBase, username, filename)

	fh, err := os.Open(fp)
	if err != nil {
		fmt.Println(err)
		return err
	}

	io.Copy(s, fh)

	return nil
}

func main() {
	ssh.Handle(func(s ssh.Session) {
		op := s.Command()[0]

		var err error = nil
		switch op {
		case "u":
			name := s.Command()[1]
			err = processUpload(s, name)
		case "d":
			name := s.Command()[1]
			err = processDownload(s, name)
		case "l":
			err = processList(s)
		}
		if err != nil {
			fmt.Println(err)
			fmt.Fprintln(s, err)
		}
	})

	ssh_opts := []ssh.Option{
		ssh.HostKeyFile("./server.key"),
	}

	log.Fatal(ssh.ListenAndServe(":8080", nil, ssh_opts...))
}
