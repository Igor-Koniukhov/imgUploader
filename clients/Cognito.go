package clients

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"log"
)

type CognitoClient interface {
	SignUp(email string, password string) (error, string)
}

type awsCognitoClient struct {
	cognitoClient *cognito.CognitoIdentityProvider
	appClientId   string
}

func (ctx *awsCognitoClient) SignUp(email string, password string) (error, string) {
	user := &cognito.SignUpInput{
		Username: aws.String(email),
		Password: aws.String(password),
		ClientId: aws.String(ctx.appClientId),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
		},
	}
	result, err := ctx.cognitoClient.SignUp(user)
	if err != nil {
		return err, ""
	}
	return nil, result.String()
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
