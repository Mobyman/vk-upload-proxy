package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"io/ioutil"
	"crypto/md5"
	"encoding/hex"
	"strings"
	"errors"
	"log"
	"encoding/json"
	"flag"
)

const DIRECTORY = "/tmp/"

type UploadBody struct {
	From string
	To   string
}

func Upload(url, file string) (response string, err error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	f, err := os.Open(file)
	if err != nil {
		return "", errors.New("File not found: " + file)
	}
	defer f.Close()
	fw, err := w.CreateFormFile("photo", file)
	if err != nil {
		return "", errors.New("File not found: " + file)
	}
	if _, err = io.Copy(fw, f); err != nil {
		return "", errors.New("File not found: " + file)
	}
	// Add the other fields
	if fw, err = w.CreateFormField("photo"); err != nil {
		return "", errors.New("File not found: " + file)
	}
	if _, err = fw.Write([]byte("photo")); err != nil {
		return "", errors.New("File not found: " + file)
	}
	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return "", errors.New("File not found: " + file)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.New("Error status: " + res.Status)
	}

	err = os.Remove(file)
	if err != nil {
		return "", err
	}

	return string(body), nil

}

func GetMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Download(url string) (string, error) {

	urlSplit := strings.Split(url, ".")
	ext := urlSplit[len(urlSplit)-1]

	if !(ext == "png" || ext == "jpg") {
		return "", errors.New("Not allowed extension: " + ext)
	}

	dFilename := fmt.Sprintf("%s.%s", GetMD5(url), ext)
	dFilepath := DIRECTORY + dFilename

	if _, err := os.Stat(dFilepath); os.IsNotExist(err) {

		out, err := os.Create(dFilepath)
		if err != nil {
			return "", err
		}

		defer out.Close()
		resp, err := http.Get(url)

		defer resp.Body.Close()
		_, err = io.Copy(out, resp.Body)

		if err != nil {
			return "", err
		}
	}

	return dFilepath, nil

}

func proxyHandler(w http.ResponseWriter, r *http.Request) {

	bodyString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	log.Println(string(r.RequestURI) + " " + string(bodyString))

	var body UploadBody
	err = json.Unmarshal(bodyString, &body)
	if err != nil {
		log.Println(err)
	}

	file, err := Download(body.From)
	if err != nil {
		fmt.Print(err)
		return
	}

	res, err := Upload(body.To, file)
	w.Header().Set("Content-Type", "application/json")

	if len(res) != 0 {
		fmt.Fprint(w, res)
	} else {
		fmt.Fprint(w, err)
	}
}

func main() {
	var host string = "127.0.0.1"
	var port string = "9090"

	flag.StringVar(&host, "host", host, "bind host")
	flag.StringVar(&port, "port", port, "bind port")
	flag.Parse()

	fmt.Println("Server started " + host + ":" + port)

	http.HandleFunc("/upload", proxyHandler)
	err := http.ListenAndServe(host+":"+port, nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
