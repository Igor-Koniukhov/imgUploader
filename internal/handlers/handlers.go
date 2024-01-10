package handlers

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"imageAploaderS3/clients"
	"imageAploaderS3/internal/config"
	"imageAploaderS3/internal/render"
	"imageAploaderS3/models"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type Repository struct {
	app *config.AppConfig
}

func NewRepository(a *config.AppConfig) *Repository {
	return &Repository{
		app: a,
	}
}

func (m *Repository) HomePage(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "upload.page.tmpl", &models.TemplateData{
		Name: m.app.Name,
	})
}

func (m *Repository) AuthPageHandler(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "signup.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Signup(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	name := r.FormValue("name")
	m.app.Name = name
	m.app.Email = email
	password := r.FormValue("password")
	cognitoClient := clients.NewCognitoClient(os.Getenv("S3_REGION"), os.Getenv("CLIENT_ID"))
	err, _ := cognitoClient.SignUp(email, name, password)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/verify", http.StatusSeeOther)
	return
}

func (m *Repository) VerifyPageHandler(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "code.page.tmpl", &models.TemplateData{})

}

func (m *Repository) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	cognitoClient := clients.NewCognitoClient(os.Getenv("S3_REGION"), os.Getenv("CLIENT_ID"))
	err, result := cognitoClient.ConfirmSignUp(m.app.Email, code)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(result)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (m *Repository) LoginHandler(w http.ResponseWriter, r *http.Request) {

	email := r.FormValue("email")
	password := r.FormValue("password")
	cognitoClient := clients.NewCognitoClient(os.Getenv("S3_REGION"), os.Getenv("CLIENT_ID"))
	err, _, initiateAuthOutput := cognitoClient.SignIn(email, password)
	if err != nil {
		log.Println(err)
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	if initiateAuthOutput != nil && initiateAuthOutput.AuthenticationResult != nil && initiateAuthOutput.AuthenticationResult.IdToken != nil {

		setToken := &http.Cookie{
			Name:     "Authorization",
			Value:    *initiateAuthOutput.AuthenticationResult.IdToken,
			HttpOnly: true,
			SameSite: 0,
		}
		http.SetCookie(w, setToken)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (m *Repository) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{})
}

func (m *Repository) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
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