package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gliderlabs/ssh"
)

const uploadsBase string = "./uploads/"

var uploadsRoot *os.Root

func init() {
	os.MkdirAll(uploadsBase, 0750)
	root, err := os.OpenRoot(uploadsBase)
	if err != nil {
		panic(err)
	}
	uploadsRoot = root
}

func processUpload(s ssh.Session, filename string) error {
	username := filepath.Base(s.User())
	userRoot, err := uploadsRoot.OpenRoot(username)
    if err != nil {
        // If not exists, create then reopen
        if mkErr := uploadsRoot.Mkdir(username, 0750); mkErr != nil {
            return mkErr
        }
        userRoot, err = uploadsRoot.OpenRoot(username)
        if err != nil {
            return err
        }
    }
	defer userRoot.Close()

	fh, err := userRoot.Create(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer fh.Close()

	io.Copy(fh, s)

	return nil
}

func processDownload(s ssh.Session, filename string) error {
	username := filepath.Base(s.User())
	userRoot, err := uploadsRoot.OpenRoot(username)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer userRoot.Close()

	fh, err := userRoot.Open(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer fh.Close()

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
