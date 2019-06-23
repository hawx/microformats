package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	mf2 "github.com/andyleap/microformats"
	"hawx.me/code/microformats/authorship"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: microformats COMMAND URL")
		os.Exit(1)
	}

	var cmd func(string) error
	switch os.Args[1] {
	case "parse":
		cmd = runParse
	case "author":
		cmd = runAuthor
	default:
		cmd = func(_ string) error {
			fmt.Fprintln(os.Stderr, "Usage: microformats COMMAND URL")
			os.Exit(1)
			return nil
		}
	}

	if err := cmd(os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func runParse(uri string) error {
	baseURL, err := url.Parse(uri)
	if err != nil {
		return err
	}

	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	parser := mf2.New()
	data := parser.Parse(resp.Body, baseURL)

	return json.NewEncoder(os.Stdout).Encode(data)
}

func runAuthor(uri string) error {
	authorURL, err := url.Parse(uri)
	if err != nil {
		return err
	}

	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	author := authorship.Parse(resp.Body, authorURL)

	return json.NewEncoder(os.Stdout).Encode(author)
}
