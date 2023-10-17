package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/disintegration/imaging"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func main() {
	const srcFolder = "example-media"
	files, err := os.ReadDir(srcFolder)
	if err != nil {
		log.Panic(err)
	}
	for _, file := range files {
		fileName := file.Name()
		if fileName == ".gitignore" || file.IsDir() {
			continue
		}

		fileNameSplit := strings.Split(fileName, ".")
		fileNameExtension := fileNameSplit[len(fileNameSplit)-1]

		fileBytes, err := os.ReadFile(fmt.Sprintf("%s/%s", srcFolder, fileName))
		if err != nil {
			log.Println(err)
			continue
		}

		var image image.Image
		switch fileNameExtension {
		case "pdf": //PDF files.
			image, err = PDFToThumbnail(1, fileBytes)
			if err != nil {
				log.Println(err)
				continue
			}
		case "mp4", "m2v", "mkv", "avi", "mpg", "mov", "wmv", "hevc": // Video files.
			image, err = VideoToThumbnail(fileBytes)
			if err != nil {
				log.Println(err)
				continue
			}
		case "png", "jpg", "jpeg", "bmp", "gif": // Image files.
			image, err = imaging.Decode(bytes.NewBuffer(fileBytes), imaging.AutoOrientation(true))
			if err != nil {
				log.Println(err)
				continue
			}
		}

		if image != nil {
			imageBuffer := new(bytes.Buffer)

			err = png.Encode(imageBuffer, image)
			if err != nil {
				log.Println(err)
				continue
			}

			thumbnail := imaging.Fit(image, 28, 28, imaging.Lanczos)

			thumbnailBuffer := new(bytes.Buffer)
			err = png.Encode(thumbnailBuffer, thumbnail)
			if err != nil {
				log.Println(err)
				continue
			}

			file, err := os.OpenFile(fmt.Sprintf("save/%s.png", fileNameSplit[0]), os.O_WRONLY|os.O_CREATE, 0664)
			if err != nil {
				log.Println(err)
				continue
			}

			defer file.Close()

			_, err = file.Write(thumbnailBuffer.Bytes())
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func VideoToThumbnail(file []byte) (image.Image, error) {
	imageBytes := bytes.NewBuffer(nil)

	err := ffmpeg.Input("pipe:").
		WithInput(bytes.NewBuffer(file)).
		Output("pipe:", ffmpeg.KwArgs{"vframes": "1", "ss": "00:00:00.100", "vcodec": "png", "f": "image2"}).
		WithOutput(imageBytes, os.Stdout).
		Run()
	if err != nil {
		return nil, err
	}

	return imaging.Decode(imageBytes, imaging.AutoOrientation(true))
}

func PDFToThumbnail(pageNumber int, file []byte) (image.Image, error) {
	pageSetting := fmt.Sprintf("-sPageList=%d", pageNumber)

	cmd := exec.Command(
		"gs",
		"-sDEVICE=pngalpha",
		"-o",
		"-",
		pageSetting,
		"-dTextAlphaBits=4",
		"-dGraphicsAlphaBits=4",
		"-r300",
		"-q",
		"-",
	)
	cmd.Stdin = bytes.NewReader(file)
	imageOut, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	imageInput := bytes.NewBuffer(imageOut)

	return imaging.Decode(imageInput, imaging.AutoOrientation(true))
}
