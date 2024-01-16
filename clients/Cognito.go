package clients

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"log"
)

type CognitoClient interface {
	SignUp(email string, name string, password string, birthdate string) (error, string)
	ConfirmSignUp(email string, code string) (error, string)
	SignIn(email string, password string) (error, string, *cognito.InitiateAuthOutput)
}

type awsCognitoClient struct {
	cognitoClient *cognito.CognitoIdentityProvider
	appClientId   string
}

func (ctx *awsCognitoClient) SignUp(
	email string,
	name string,
	password string,
	birthdate string) (error, string) {
	user := &cognito.SignUpInput{
		Username: aws.String(email),
		Password: aws.String(password),
		ClientId: aws.String(ctx.appClientId),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			{
				Name:  aws.String("name"),
				Value: aws.String(name),
			},
			{
				Name:  aws.String("birthdate"),
				Value: aws.String(birthdate),
			},
		},
	}
	result, err := ctx.cognitoClient.SignUp(user)
	if err != nil {
		return err, ""
	}
	return nil, result.String()
}

func (ctx *awsCognitoClient) ConfirmSignUp(email string, code string) (error, string) {
	confirmSignUpInput := &cognito.ConfirmSignUpInput{
		Username:         aws.String(email),
		ConfirmationCode: aws.String(code),
		ClientId:         aws.String(ctx.appClientId),
	}
	result, err := ctx.cognitoClient.ConfirmSignUp(confirmSignUpInput)

	if err != nil {
		return err, ""
	}
	return nil, result.String()
}

func (ctx *awsCognitoClient) SignIn(email string, password string) (error, string, *cognito.InitiateAuthOutput) {
	initiateAuthInput := &cognito.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: aws.StringMap(map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		}),
		ClientId: aws.String(ctx.appClientId),
	}
	result, err := ctx.cognitoClient.InitiateAuth(initiateAuthInput)

	if err != nil {
		return err, "", nil
	}
	return nil, result.String(), result
}

func NewCognitoClient(cognitoRegion string, cognitoAppClientId string) CognitoClient {
	conf := &aws.Config{Region: aws.String(cognitoRegion)}
	sess, err := session.NewSession(conf)
	if err != nil {
		log.Println(err)
	}
	client := cognito.New(sess)

	return &awsCognitoClient{
		cognitoClient: client,
		appClientId:   cognitoAppClientId,
	}

}
