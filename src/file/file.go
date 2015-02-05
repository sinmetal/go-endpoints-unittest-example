package file

import (
	"appengine"
	"appengine/file"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func init() {
	http.HandleFunc("/file", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	switch r.Method {
	default:
		http.Error(w, "unsupported method.", http.StatusMethodNotAllowed)
	case "POST":
		uploadFile(c, w, r)
	case "GET":
		downloadFile(c, w, r)
	}
}

func uploadFile(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 * 1024 * 1024) // 10MB
	if err != nil {
		if err.Error() == "permission denied" {
			fmt.Fprint(w, "アップロード可能な容量を超えています。\n")
		} else {
			c.Errorf("%s", err.Error())
		}
		return
	}
	file, fileHeader, err := r.FormFile("filename")
	if err != nil {
		c.Errorf("%s", err.Error())
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		c.Errorf("%s", err.Error())
		return
	}

	absFilename, err := directStore(c, data, fileHeader.Filename)
	if err != nil {
		c.Errorf("%s", err.Error())
		return
	}

	log.Printf("%s", absFilename)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(absFilename))
}

func downloadFile(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	bn, err := file.DefaultBucketName(c)
	if err != nil {
		c.Errorf("%s", err.Error())
	}

	filename := "/gs/" + bn + "/" + r.FormValue("name")
	log.Printf("filename : " + filename)
	fr, err := file.Open(c, filename)
	if err != nil {
		c.Errorf("%s", err.Error())
	}
	defer fr.Close()

	w.Header().Set("Cache-Control:public", "max-age=120")
	io.Copy(w, fr)
}

func directStore(c appengine.Context, data []byte, filename string) (absFilename string, err error) {
	bn, err := file.DefaultBucketName(c)
	if err != nil {
		return "", err
	}

	opts := &file.CreateOptions{
		MIMEType:   "image/png",
		BucketName: bn,
	}

	wc, absFilename, err := file.Create(c, filename, opts)
	if err != nil {
		return "", err
	}
	defer wc.Close()

	_, err = wc.Write(data)
	if err != nil {
		return "", err
	}

	return absFilename, nil
}
