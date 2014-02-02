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
	RelativeFileName string
	ParentDir string
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

type DirectoryList struct {
	DirNames []string
}

func picturePage(basePath string, picPath string, w io.Writer) {
	templates, err := template.ParseGlob(*templatePath + "html/*.html")
	if err != nil {
		log.Fatal("picturePage: Error loading templates: ", err)
	}

	dir, _ := os.Open(path.Join(basePath, picPath))
	picFileNames, _ := dir.Readdirnames(0)
	sort.Strings(picFileNames)
	pictures := []Picture{}
	for _, picFileName := range picFileNames {
		if strings.HasSuffix(picFileName, ".CR2") {
			pictures = append(pictures, Picture{path.Join(picPath, picFileName), path.Join(picPath, fmt.Sprintf("%v-preview1.jpg", picFileName[0:len(picFileName)-4])), picFileName, picPath})
		}
	}

	picDir := PictureDirectory{picPath, pictures}

	t := templates.Lookup("picture_grid.html")
	t.Execute(w, picDir)
}

func zipAll(basePath string, picPath string, w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"photos-%s.zip\"", picPath));
	z := zip.NewWriter(w)
	dir, _ := os.Open(path.Join(basePath, picPath))
	picFiles, _ := dir.Readdir(0)
	for _, picFile := range picFiles {
		if strings.HasSuffix(picFile.Name(), ".CR2") {
			fh, err := zip.FileInfoHeader(picFile)
			fh.Method = zip.Store
			f, err := z.CreateHeader(fh)
			if err != nil {
				log.Fatal("zipAll: error creating file writer: ", err)
			} 
			p, _ := os.Open(path.Join(path.Join(basePath, picPath), picFile.Name()))
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

// TODO: share common functionality with zipAll in a helper
func zipSelected(basePath string, picPath string, fileNames []string, w http.ResponseWriter) {
	fileMap := make(map[string]bool)
	for _, name := range fileNames {
		fileMap[name] = true
	}
	h := w.Header()
	h.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"selectedPhotos-%s.zip\"", picPath));
	z := zip.NewWriter(w)
	dir, _ := os.Open(path.Join(basePath, picPath))
	picFiles, _ := dir.Readdir(0)
	for _, picFile := range picFiles {
		if fileMap[picFile.Name()] {
			fh, err := zip.FileInfoHeader(picFile)
			fh.Method = zip.Store
			f, err := z.CreateHeader(fh)
			if err != nil {
				log.Fatal("zipSelected: error creating file writer: ", err)
			} 
			p, _ := os.Open(path.Join(path.Join(basePath, picPath), picFile.Name()))
			_, err = io.Copy(f, p)
			if err != nil {
				log.Fatal("zipSelected: error copying file: ", err)
			}
			p.Close()
		}
	}

	err := z.Close()
	if err != nil {
		log.Fatal("zipSelected: error closing zip file: ", err)
	}

}

func listDirs(dirName string, prefix string, basePath string) []string {
	dirStrings := []string{}

	dir, _ := os.Open(path.Join(path.Join(basePath, prefix), dirName))
	files, _ := dir.Readdir(0)
	for _, file := range files {
		if file.IsDir() {
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
		zipSelected(*imagePath, req.Form.Get("path"), req.Form["selectedFiles"], w);
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

	err := http.ListenAndServe(":" + *port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
