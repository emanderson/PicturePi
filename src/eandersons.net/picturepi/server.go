package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
)

var closurePath = flag.String(
	"closurePath", "closure-library",
	"directory where closure library is found")
var imagePath = flag.String(
	"photoPath", "~/pictures",
	"directory where image files are found")
var port = flag.String(
	"port", "8080", "port to serve on")
var staticPath = flag.String(
	"staticPath", "static/",
	"directory where static files are stored")
var templatePath = flag.String(
	"templatePath", "tmpl/eandersons.net/picturepi/",
	"directory where template files are stored")

const ClosurePath = "/closure-library/"
const ImagePath = "/images/"
const StaticPath = "/static/"

type Picture struct {
	RawFileName      string
	PreviewFileName  string
	RelativeFileName string
	ParentDir        string
}

func (p *Picture) PreviewFileURL() string {
	return ImagePath + p.PreviewFileName
}

func (p *Picture) RawFileURL() string {
	return ImagePath + p.RawFileName
}

type PictureDirectory struct {
	Name     string
	Pictures []Picture
}

type DirectoryList struct {
	DirNames []string
}

func IsPictureFile(fileName string) bool {
	return (strings.HasSuffix(fileName, ".CR2") ||
		strings.HasSuffix(fileName, ".JPG") ||
		strings.HasSuffix(fileName, ".MOV"))
}

func picturePage(basePath string, picPath string, w io.Writer) {
	templates, err := template.ParseGlob(*templatePath + "html/" + "*.html")
	if err != nil {
		log.Fatal("picturePage: Error loading templates: ", err)
	}

	dirPath := path.Join(basePath, picPath)
	dir, err := os.Open(dirPath)
	if err != nil {
		log.Fatalf("Unable to open %s: %v", dirPath, err)
	}
	picFileNames, err := dir.Readdirnames(-1)
	if err != nil {
		log.Fatalf("Unable to read names from %s(%v): %v", dirPath, *dir, err)
	}
	sort.Strings(picFileNames)
	pictures := []Picture{}
	for _, picFileName := range picFileNames {
		if IsPictureFile(picFileName) {
			pictures = append(pictures,
				Picture{path.Join(picPath, picFileName),
					path.Join(picPath, fmt.Sprintf("%v-preview1.jpg", picFileName[0:len(picFileName)-4])),
					picFileName, picPath})
		}
	}

	picDir := PictureDirectory{picPath, pictures}

	t := templates.Lookup("picture_grid.html")
	t.Execute(w, picDir)
}

func zipAll(basePath string, picPath string, w http.ResponseWriter) {
	zipBase(basePath, picPath, w, IsPictureFile)
}

func zipSelected(basePath string, picPath string, fileNames []string, w http.ResponseWriter) {
	fileMap := make(map[string]bool)
	for _, name := range fileNames {
		fileMap[name] = true
	}
	zipBase(basePath, picPath, w, func(fileName string) bool {
		return fileMap[fileName]
	})
}

func zipBase(basePath string, picPath string, w http.ResponseWriter, acceptor func(string) bool) {
	h := w.Header()
	h.Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"photos-%s.zip\"", picPath))
	z := zip.NewWriter(w)
	dirPath := path.Join(basePath, picPath)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Printf("Found %v(%v)", files, err)
	}
	for _, fileInfo := range files {
		if acceptor(fileInfo.Name()) {
			fh, err := zip.FileInfoHeader(fileInfo)
			fh.Method = zip.Store
			f, err := z.CreateHeader(fh)
			if err != nil {
				log.Fatal("zip: error creating file writer: ", err)
			}
			p, _ := os.Open(path.Join(path.Join(basePath, picPath), fileInfo.Name()))
			_, err = io.Copy(f, p)
			if err != nil {
				log.Fatal("zip: error copying file: ", err)
			}
			p.Close()
		}
	}

	err = z.Close()
	if err != nil {
		log.Fatal("zip: error closing zip file: ", err)
	}
}

func listDirs(dirName string, prefix string, basePath string) []string {
	dirStrings := []string{}

	dirPath := path.Join(path.Join(basePath, prefix), dirName)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Printf("Found %v(%v)", files, err)
	}
	for _, file := range files {
		if (file.Mode() & os.ModeSymlink) == os.ModeSymlink {
			filename := path.Join(dirPath, file.Name())
			file, err = os.Stat(filename)
			if err != nil {
				log.Printf("Failed to read %s: %v", filename, err)
				continue
			}
		}
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			subDirStrings := listDirs(file.Name(), path.Join(prefix, dirName), basePath)
			for _, subDir := range subDirStrings {
				dirStrings = append(dirStrings, path.Join(file.Name(), subDir))
			}
			if len(subDirStrings) == 0 {
				dirStrings = append(dirStrings, file.Name())
			}
		}
	}
	return dirStrings
}

func listDirectories(picPath string, w io.Writer) {
	templates, err := template.ParseGlob(*templatePath + "html/*.html")
	if err != nil {
		log.Fatal("picturePage: Error loading templates: ", err)
	}

	dirNames := listDirs(".", "", picPath)

	picDir := DirectoryList{dirNames}

	t := templates.Lookup("dir_list.html")
	t.Execute(w, picDir)
}

func PicturePiServer(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/zip" {
		zipAll(*imagePath, req.URL.Query().Get("path"), w)
	} else if req.URL.Path == "/zipSelected" {
		req.ParseForm()
		zipSelected(*imagePath, req.Form.Get("path"), req.Form["selectedFiles"], w)
	} else if req.URL.Path == "/list" || req.URL.Path == "/" {
		listDirectories(*imagePath, w)
	} else if strings.HasPrefix(req.URL.Path, "/photos/") {
		picturePage(*imagePath, req.URL.Path[len("/photos/"):], w)
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/", PicturePiServer)
	http.Handle(ImagePath, http.StripPrefix(ImagePath, http.FileServer(http.Dir(*imagePath))))
	http.Handle(ClosurePath, http.StripPrefix(ClosurePath, http.FileServer(http.Dir(*closurePath))))
	http.Handle(StaticPath, http.StripPrefix(StaticPath, http.FileServer(http.Dir(*staticPath))))

	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
