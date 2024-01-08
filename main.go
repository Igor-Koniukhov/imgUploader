package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/joho/godotenv"
	"html/template"
	"imageAploaderS3/clients"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}
	log.Println("The API has started.")
	cognitoClient := clients.NewCognitoClient(os.Getenv("S3_REGION"), os.Getenv("CLIENT_ID"))
	err, result := cognitoClient.SignUp("testemail777@gmail.com", "123456")
	if err != nil {
		log.Println(err)
	}
	fmt.Println("-------------Result: ", result)
	http.HandleFunc("/", homePage)
	http.HandleFunc("/upload", uploadFileHandler)
	err = http.ListenAndServe(os.Getenv("PORT"), nil)
	if err != nil {
		log.Println(err)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/upload.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Println(err)
	}
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			log.Println(err)
		}
		file, handler, err := r.FormFile("myFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)

		sess, _ := session.NewSession(&aws.Config{
			Region: aws.String(os.Getenv("S3_REGION")),
		})
		uploader := s3manager.NewUploader(sess)
		fileKey := fmt.Sprintf("%s-%s", time.Now().Format("20060102-150405"), handler.Filename)

		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(os.Getenv("S3_BUCKET")),
			Key:    aws.String(fileKey),
			Body:   file,
		})

		if err != nil {
			fmt.Println("Failed to upload", err)
			return
		}

		time.Sleep(time.Millisecond * 300)

		_, err = fmt.Fprintf(w, "Successfully uploaded %s%s\n", os.Getenv("SOURCE_URL"), fileKey)
		if err != nil {
			log.Println(err)
		}
		_, err = fmt.Fprintf(w, "Successfully uploaded to resized %sresized/%s\n", os.Getenv("SOURCE_URL"), fileKey)
		if err != nil {
			log.Println(err)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		_, err := fmt.Fprint(w, "Only POST method is supported")
		if err != nil {
			log.Println(err)
		}
	}
}
