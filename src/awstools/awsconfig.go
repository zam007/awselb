package awstools

import (
	"awselb/conf"
	"awselb/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

func GetAwsRegion() (region string) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/placement/availability-zone")
	if err != nil {
		logger.Logging.Fatal("cant't get aws region from ec2, exit")
	}else {
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	logger.Logging.WithFields(logrus.Fields{"region": string(data)}).Info("get ec2 region success !")
	return string(data)
}

func GetInstanceId() (insid string) {

	resp, err := http.Get("http://169.254.169.254/latest/meta-data/instance-id")
	if err != nil {
		logger.Logging.Fatal("can't get ec2 id")
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	logger.Logging.WithFields(logrus.Fields{"instance-id": string(data)}).Info("get ec2 id success !")
	return string(data)
}

func GetAwsConfig(region string) *aws.Config {
	var cr *credentials.Credentials

	envParams := conf.GetEnvParams()

	switch envParams["app_env"] {

	case "dev":
		logger.Logging.Warn("Dev env, Use key to access aws")

		accessKeyID := envParams["aws_id"]
		secretAccessKey := envParams["aws_key"]

		var cerValue = credentials.Value{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		}

		cr = credentials.NewStaticCredentialsFromCreds(cerValue)

	case "product":
		logger.Logging.Info("Product env, Use Ec2Role to access aws")

		session := session.Must(session.NewSession())
		ec2m := ec2metadata.New(session,
			&aws.Config{
				HTTPClient: &http.Client{
					Timeout: 20 * time.Second},
			})

		cr = credentials.NewCredentials(&ec2rolecreds.EC2RoleProvider{
			Client: ec2m,
		})

	default:
		logger.Logging.Fatal("no app_env , exit!")
	}

	return &aws.Config{
		Region:      aws.String(region),
		Credentials: cr,
	}
}
