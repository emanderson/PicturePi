package main

import (
	"io"
	"net/http"
	"log"
	"fmt"
	"flag"
	"os"
	"sort"
	"strings"
	"html/template"
)

const ImagePath = "/images/";
const ClosurePath = "/closure-library/";
const StaticPath = "/static/";

type Picture struct {
	RawFileName string
	PreviewFileName string
}

func (p *Picture) RawFileURL() string {
	return ImagePath + p.RawFileName;
}

func (p *Picture) PreviewFileURL() string {
	return ImagePath + p.PreviewFileName;
}

type PictureDirectory struct {
	Name string
	Pictures []Picture
}

var templateDir = flag.String("templatePath", "tmpl/eandersons.net/picturepi/", "directory where template files are stored")
var staticDir = flag.String("staticPath", "static/", "directory where static files are stored")
var closureDir = flag.String("closurePath", "closure-library", "directory where closure library is found")
var imageDir = flag.String("photoPath", "~/pictures", "directory where image files are found")

func picpage(path string, w io.Writer) {
	templates, err := template.ParseGlob(*templateDir + "html/*.html")
	if err != nil {
		log.Fatal("Error loading templates: ", err)
	}
	for _, t := range templates.Templates() {
		fmt.Println(t.Name())
	}

	dir, _ := os.Open(path)
	picFileNames, _ := dir.Readdirnames(0)
	sort.Strings(picFileNames)
	pictures := []Picture{}
	for _, picFileName := range picFileNames {
		if strings.HasSuffix(picFileName, ".CR2") {
			pictures = append(pictures, Picture{picFileName, fmt.Sprintf("%v-preview1.jpg", picFileName[0:len(picFileName)-4])})
		}
	}

	picDir := PictureDirectory{dir.Name(), pictures}
	picDir = picDir

	t := templates.Lookup("picture_grid.html")
	t.Execute(w, picDir)
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	picpage(*imageDir, w)
}

func main() {
	flag.Parse()
	fmt.Println(*imageDir)
	fmt.Println(*templateDir)
	fmt.Println(*closureDir)
	http.HandleFunc("/", HelloServer)
	http.Handle(ImagePath, http.StripPrefix(ImagePath, http.FileServer(http.Dir(*imageDir))))
	http.Handle(ClosurePath, http.StripPrefix(ClosurePath, http.FileServer(http.Dir(*closureDir))))
	http.Handle(StaticPath, http.StripPrefix(StaticPath, http.FileServer(http.Dir(*staticDir))))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
