package handlers

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-xray-sdk-go/xray"
	"imageUploader/clients"
	"imageUploader/internal/config"
	"imageUploader/internal/render"
	"imageUploader/internal/repository/dbrepo"
	"imageUploader/models"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type User interface {
	HomePage(w http.ResponseWriter, r *http.Request)
	AuthPageHandler(w http.ResponseWriter, r *http.Request)
	Signup(w http.ResponseWriter, r *http.Request)
	VerifyPageHandler(w http.ResponseWriter, r *http.Request)
	VerifyHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LoginPageHandler(w http.ResponseWriter, r *http.Request)
	UploadFileHandler(w http.ResponseWriter, r *http.Request)
	GetUserName(w http.ResponseWriter, r *http.Request)
	AuthSet(next http.Handler) http.Handler
	UserDataSet(next http.Handler) http.Handler
	XRayMiddleware(appName string) func(next http.Handler) http.Handler
}

type Repository struct {
	App  *config.AppConfig
	repo dbrepo.UserRepository
}

func NewUserHandlers(a *config.AppConfig, repo dbrepo.UserRepository) *Repository {
	return &Repository{
		App:  a,
		repo: repo,
	}
}

func (m *Repository) HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Context().Value("newKey"), "from home")
	err := render.RenderTemplate(w, r, "upload.page.tmpl", &models.TemplateData{
		Name: m.App.Name,
	})
	if err != nil {
		return
	}
}

func (m *Repository) AuthPageHandler(w http.ResponseWriter, r *http.Request) {
	err := render.RenderTemplate(w, r, "signup.page.tmpl", &models.TemplateData{})
	if err != nil {
		return
	}
}

func (m *Repository) Signup(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	userFromCache, err := m.repo.GetUserFromCacheByEmail(r.Context(), email)
	if err != nil {
		fmt.Println("Error getting user from cache: ", err)
	}
	m.App.UserInfoFromCache = userFromCache
	userByEmail, err := m.repo.GetUserByEmail(r.Context(), email)
	if userByEmail != nil {
		m.App.ErrorMessage = "User already exists!"
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	name := r.FormValue("name")
	birthdateStr := r.FormValue("birthdate")
	m.App.Name = name
	m.App.Email = email
	m.App.Birthdate = birthdateStr
	password := r.FormValue("password")
	cognitoClient := clients.NewCognitoClient(r.Context(), os.Getenv("S3_REGION"), os.Getenv("CLIENT_ID"))
	err, _ = cognitoClient.SignUp(email, name, password, birthdateStr)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if userByEmail == nil {
		user, _, err := m.repo.CreateUser(r.Context(), &models.User{
			Name:      name,
			Email:     email,
			BirthDate: birthdateStr,
		})
		if err != nil {
			return
		}
		fmt.Println(user)
		http.Redirect(w, r, "/verify", http.StatusSeeOther)
		return
	}

}

func (m *Repository) VerifyPageHandler(w http.ResponseWriter, r *http.Request) {
	err := render.RenderTemplate(w, r, "code.page.tmpl", &models.TemplateData{})
	if err != nil {
		return
	}

}

func (m *Repository) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	cognitoClient := clients.NewCognitoClient(r.Context(), os.Getenv("S3_REGION"), os.Getenv("CLIENT_ID"))
	err, result := cognitoClient.ConfirmSignUp(m.App.Email, code)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(result)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (m *Repository) LoginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	cognitoClient := clients.NewCognitoClient(r.Context(), os.Getenv("S3_REGION"), os.Getenv("CLIENT_ID"))
	err, _, initiateAuthOutput := cognitoClient.SignIn(email, password)
	if err != nil {
		log.Println(err)
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	if initiateAuthOutput != nil && initiateAuthOutput.AuthenticationResult != nil && initiateAuthOutput.AuthenticationResult.IdToken != nil {

		accessToken := &http.Cookie{
			Name:     "AccessToken",
			Value:    *initiateAuthOutput.AuthenticationResult.AccessToken,
			HttpOnly: true,
			SameSite: 0,
		}
		refreshToken := &http.Cookie{
			Name:     "RefreshToken",
			Value:    *initiateAuthOutput.AuthenticationResult.RefreshToken,
			HttpOnly: true,
			SameSite: 0,
		}
		tokenId := &http.Cookie{
			Name:     "TokenId",
			Value:    *initiateAuthOutput.AuthenticationResult.IdToken,
			HttpOnly: true,
			SameSite: 0,
		}
		http.SetCookie(w, accessToken)
		http.SetCookie(w, refreshToken)
		http.SetCookie(w, tokenId)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (m *Repository) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	err := render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
		Error:    m.App.ErrorMessage,
		UserInfo: m.App.UserInfoFromCache,
	})
	if err != nil {
		return
	}
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

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(os.Getenv("S3_REGION")),
		})
		if err != nil {
			fmt.Println(err)
		}
		s3Client := s3.New(sess)
		xray.AWS(s3Client.Client)
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
		user, err := m.repo.GetUserByEmail(r.Context(), m.App.Email)
		if err != nil {
			return
		}
		url := fmt.Sprintf("%s%s", os.Getenv("SOURCE_URL"), fileKey)
		_, err = m.repo.SaveUserImgUrl(r.Context(), user.ID, url)
		if err != nil {
			return
		}

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

func (m *Repository) GetUserName(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, m.App.Name)
	if err != nil {
		return
	}

}
