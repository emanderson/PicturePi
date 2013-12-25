package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"log"
	"fmt"
	"flag"
	"os"
	"sort"
	"strings"
	"html/template"
)

type Picture struct {
	RawFileName string
	PreviewFileName string
}

type PictureDirectory struct {
	Name string
	Pictures []Picture
}

var templateDir = flag.String("templatePath", "tmpl/eandersons.net/picturepi/", "directory where template files are stored")

func picpage(path string, w io.Writer) {
	page, err := ioutil.ReadFile(*templateDir + "html/picture_grid.html")
	if err != nil {
		log.Fatal("Couldn't open template file", err)
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

	t := template.Must(template.New("page").Parse(string(page)))
	t.Execute(w, picDir)
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	picpage(flag.Arg(0), w)
}

func main() {
	flag.Parse()
	fmt.Println(flag.Arg(0))
	fmt.Println(*templateDir)
	http.HandleFunc("/", HelloServer)
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(flag.Arg(0)))))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
