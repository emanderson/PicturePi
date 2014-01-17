package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
)

var closurePath = flag.String("closurePath", "closure-library", "directory where closure library is found")
var imagePath = flag.String("photoPath", "~/pictures", "directory where image files are found")
var port = flag.String("port", "8080", "port to serve on")
var staticPath = flag.String("staticPath", "static/", "directory where static files are stored")
var templatePath = flag.String("templatePath", "tmpl/eandersons.net/picturepi/", "directory where template files are stored")

const ClosurePath = "/closure-library/";
const ImagePath = "/images/";
const StaticPath = "/static/";

type Picture struct {
	RawFileName string
	PreviewFileName string
}

func (p *Picture) PreviewFileURL() string {
	return ImagePath + p.PreviewFileName;
}

func (p *Picture) RawFileURL() string {
	return ImagePath + p.RawFileName;
}

type PictureDirectory struct {
	Name string
	Pictures []Picture
}

func picturePage(picPath string, w io.Writer) {
	templates, err := template.ParseGlob(*templatePath + "html/*.html")
	if err != nil {
		log.Fatal("picturePage: Error loading templates: ", err)
	}

	dir, _ := os.Open(picPath)
	picFileNames, _ := dir.Readdirnames(0)
	sort.Strings(picFileNames)
	pictures := []Picture{}
	for _, picFileName := range picFileNames {
		if strings.HasSuffix(picFileName, ".CR2") {
			pictures = append(pictures, Picture{picFileName, fmt.Sprintf("%v-preview1.jpg", picFileName[0:len(picFileName)-4])})
		}
	}

	picDir := PictureDirectory{dir.Name(), pictures}

	t := templates.Lookup("picture_grid.html")
	t.Execute(w, picDir)
}

func zipAll(picPath string, w io.Writer) {
	z := zip.NewWriter(w)
	dir, _ := os.Open(picPath)
	picFiles, _ := dir.Readdir(0)
	for _, picFile := range picFiles {
		if strings.HasSuffix(picFile.Name(), ".CR2") {
			fh, err := zip.FileInfoHeader(picFile)
			fh.Method = zip.Store
			f, err := z.CreateHeader(fh)
			if err != nil {
				log.Fatal("zipAll: error creating file writer: ", err)
			} 
			p, _ := os.Open(path.Join(picPath, picFile.Name()))
			_, err = io.Copy(f, p)
			if err != nil {
				log.Fatal("zipAll: error copying file: ", err)
			}
			p.Close()
		}
	}

	err := z.Close()
	if err != nil {
		log.Fatal("zipAll: error closing zip file: ", err)
	}

}

func PicturePiServer(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/zip" {
		zipAll(*imagePath, w)
	} else {
		picturePage(*imagePath, w)
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/", PicturePiServer)
	http.Handle(ImagePath, http.StripPrefix(ImagePath, http.FileServer(http.Dir(*imagePath))))
	http.Handle(ClosurePath, http.StripPrefix(ClosurePath, http.FileServer(http.Dir(*closurePath))))
	http.Handle(StaticPath, http.StripPrefix(StaticPath, http.FileServer(http.Dir(*staticPath))))

	err := http.ListenAndServe(":" + *port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
