package converter

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ConvertFile convert a file using Calibre
func ConvertFile(filename string, content []byte) (string, []byte, error) {
	dir, err := ioutil.TempDir("/tmp", "book")
	if err != nil {
		log.Print(err)
		return "", nil, err
	}
	reg, err := regexp.Compile("[^a-zA-Z0-9.]+")
	if err != nil {
		log.Print(err)
	}

	defer os.RemoveAll(dir) // clean up

	path := filepath.Join(dir, reg.ReplaceAllString(filename, ""))
	log.Print("Writing to disk: ", path)
	if err := ioutil.WriteFile(path, content, 0644); err != nil {
		log.Print("Writing to disk failed: ", err)
		return "", nil, err
	}

	baseName := strings.TrimSuffix(path, filepath.Ext(path))
	mobiName := baseName + ".mobi"

	cmd := exec.Command("ebook-convert", path, mobiName, "--mobi-keep-original-images")
	log.Printf("Running converter and waiting for it to finish...")
	output, cmdErr := cmd.Output()
	log.Printf("Output %s\n", output)
	if cmdErr != nil {
		log.Print("Error while running converter: ", cmdErr)
		return "", nil, cmdErr
	}

	content, err = ioutil.ReadFile(mobiName)
	if err != nil {
		log.Print(err)
		return "", nil, err
	}

	return mobiName, content, nil
}
