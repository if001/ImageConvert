package main

import (
	"fmt"
	"net/http"
	"log"
	"mime/multipart"
	"os"
	"io"
	"time"
	"crypto/md5"
	"encoding/hex"
	"path/filepath"
	"image"
	"image/gif"
    "image/png"
	"image/jpeg"
    "bytes"
    "errors"
	"strconv"
	"os/exec"
    "ImageConvert/lib"
    "github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
    "github.com/nfnt/resize"
)

func printHeader(header http.Header) {
	for k, v := range header {
		log.Println("Header field", k)
		log.Println("Value field", v)
	}
}
func toHash(filename string) string {
	var t = time.Now().Unix()
	text := strconv.FormatInt(t,10) + "-" + filename
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func isImage(file multipart.File) bool {
	_, format, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return (format == "jpeg") || (format == "png") || (format == "gif")
}

func convertCmd(writer http.ResponseWriter, request *http.Request) {
	if (request.Method != "POST") {
		fmt.Fprintln(writer, "許可したメソッドとはことなります。")
		return
	}
	const imageSaveDir = "/go_app/img/"
	const postFileKey = "upload_file"
	var file multipart.File
	var fileHeader *multipart.FileHeader
	var err error
	var saveFileName string
	var convertedFileName string
	const convertedPrefix = "-convert"
	const convertCmd = "/go_app/sphericalpano2rect"
    
	// printHeader(request.Header)

	file, fileHeader, err = request.FormFile(postFileKey)
	if err != nil {
		log.Println(err)
		fmt.Fprintln(writer, "faild upload")
		return
	}
	if !isImage(file) {
		log.Println(err)
		fmt.Fprintln(writer, "must image file")
		return
	}
	// todo image.DecodeConfigを使うと、fileが壊れるので再度取り直し
	file, fileHeader, err = request.FormFile(postFileKey)

	ext := filepath.Ext(fileHeader.Filename)
	saveFileName = toHash(fileHeader.Filename) + ext

	var saveImage *os.File
	saveImage, err = os.Create(imageSaveDir + saveFileName)
	if err != nil {
		log.Println(err)
		fmt.Fprintln(writer, "サーバ側でファイル確保できませんでした。")
		return
	}

	defer saveImage.Close()
	defer file.Close()
	_, err = io.Copy(saveImage, file)
	if err != nil {
		log.Println(err)
		fmt.Println("アップロードしたファイルの書き込みに失敗しました。")
		return
		//os.Exit(1)
	}
    
	convertedFileName = toHash(fileHeader.Filename) + convertedPrefix + ext
	cmd := exec.Command(convertCmd,"-c", "180,90","-u","degrees", imageSaveDir + saveFileName, imageSaveDir + convertedFileName)

	log.Println(cmd)
	result, err := cmd.Output()
	if err != nil {
		log.Println(err)
		log.Println(result)
		log.Println(string(result))
		fmt.Fprintln(writer, "convert command exec error")
		return
	}
	log.Println("covert result: ", result)

	var returnImage *os.File
	defer returnImage.Close()
	returnImage, err = os.Open(imageSaveDir + convertedFileName)
	if err != nil {
		log.Println(err)
		fmt.Fprintln(writer, "file open error")
		return
	}
	writer.Header().Set("Content-Type", "image/*")
	io.Copy(writer, returnImage)
}

func encodeImage(target io.Writer, imageData image.Image, imageFormat string) error {
    switch imageFormat {
    case "jpeg", "jpg":
        jpeg.Encode(target, imageData, nil)
    case "png":
        png.Encode(target, imageData)
    case "gif":
        gif.Encode(target, imageData, nil)
    default:
        return errors.New("invalid format")
    }
    return nil
}

func convertCal(writer http.ResponseWriter, request *http.Request) {
	if (request.Method != "POST") {
		fmt.Fprintln(writer, "許可したメソッドとはことなります。")
		return
	}
	const imageSaveDir = "/go_app/img/"
    //const imageSaveDir = "img/"
	const postFileKey = "upload_file"
	var file multipart.File
    
	file, _, err := request.FormFile(postFileKey)
	if err != nil {
		log.Println(err)
		fmt.Fprintln(writer, "faild upload")
		return
	}
	if !isImage(file) {
		log.Println(err)
		fmt.Fprintln(writer, "must image file")
		return
	}
	// todo image.DecodeConfigを使うと、fileが壊れるので再度取り直し
	file, _, err = request.FormFile(postFileKey)
    in_img, format, err := image.Decode(file)
    if err != nil {
		log.Println(err)
		fmt.Fprintln(writer, "decode error")
		return
	}

    // 半分にする
	in_img = resize.Resize(uint(in_img.Bounds().Dx()/2), uint(in_img.Bounds().Dy()/2), in_img, resize.Lanczos3)    

    // ---- flatten ------
    flatten_img := lib.ToCube(in_img)
    log.Println("flatten")
    
    // ---- cut ------
    flatten_img = lib.CutTopBottom(flatten_img)
    log.Println("cut top bottom")
    
    // ---- crop ----
	analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
	topCrop, err := analyzer.FindBestCrop(flatten_img,5, 5)
    if err != nil {
		log.Println(err)
		fmt.Fprintln(writer, "crop error")
		return
	}

    type SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	cropped_img := flatten_img.(SubImager).SubImage(topCrop)


    buf := new(bytes.Buffer)
    if encodeImage(buf, cropped_img, format) != nil {
        log.Println(err)
		fmt.Fprintln(writer, "bytes encode error")
        return
    }
	writer.Header().Set("Content-Type", "image/*")
    log.Println("end")
    writer.Write(buf.Bytes())
}


func health(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "I'm fine!!!")
}
func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	addr := ":8080"
	log.Println("listen on", addr)
	http.HandleFunc("/health", health)
	http.HandleFunc("/convert_cmd", convertCmd)
    http.HandleFunc("/convert", convertCal)
	log.Fatal(http.ListenAndServe(addr, Log(http.DefaultServeMux)))
}
