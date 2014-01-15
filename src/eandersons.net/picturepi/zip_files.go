package main

import (
	"fmt"
	"os"
	"strings"
	"archive/zip"
	"log"
	"io"
)

func main() {
	fmt.Println("Hello")

	z, err := os.Create("output.zip")
	if err != nil {
		log.Fatal(err)
	}

	w := zip.NewWriter(z)

	dir, _ := os.Open("/Users/emily/example_photos")

	picFileNames, _ := dir.Readdir(0)
	for _, picFileName := range picFileNames {
		if strings.HasSuffix(picFileName.Name(), ".CR2") {
			fh, err := zip.FileInfoHeader(picFileName)
			fh.Method = zip.Store
			f, err := w.CreateHeader(fh)
			if err != nil {
				log.Fatal(err)
			} 
			p, _ := os.Open("/Users/emily/example_photos/" + picFileName.Name())
			_, err = io.Copy(f, p)
			if err != nil {
				log.Fatal(err)
			}
			p.Close()
		}
	}
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}

	z.Close()
	
}
