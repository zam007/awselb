package cmd

import (
	"awselb/awstools"
	"awselb/cmd/update"
	"awselb/conf"
	"awselb/logger"
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var fileName, hash, region string
var elb, alb []string
var dockerTag *int64

// command update
var CmdUpdate = &cobra.Command{
	Use:   "update [command]",
	Short: "update umscloud service",
	Long: `step 1 : delete ec2 from elb,
step 2 : update service by update bash script,
step 3 : register ec2 to elb,
step 4 : check if elb all healthy,
    `,
	Args: checkUpdateArgs,
	Run:  updateRun,
}

func updateRun(cmd *cobra.Command, args []string) {
	// get ec2 id
	instanceId := awstools.GetInstanceId()
	//instanceId := "i-0a62f48ba79d4d471"

	// get elb session
	elbSrvClient := update.GetElbSrvClient(region)
	// get alb session
	albSrvClient := update.GetAlbSrvClient(region)

	// deRegister from elb
	if elb != nil {
		// get input elb names
		elbNames := getElbArray(elb)

		// init elbAction struct
		elbAction := &update.ElbAction{
			ElbNames:     elbNames,
			InstanceId:   instanceId,
			ElbSrvClient: elbSrvClient,
		}

		// deRegister instance from elb and check
		elbAction.DeRegisterInstanceFromElb()
	}
	// deRegister from alb
	if alb != nil {
		// get input alb arnNames
		albArnNames := getElbArray(alb)

		// init albAction struct
		albAction := &update.AlbAction{
			AlbArnNames:  albArnNames,
			InstanceId:   instanceId,
			AlbSrvClient: albSrvClient,
		}

		// deRegister instance from alb and check
		albAction.DeRegisterInstanceFromAlb()
	}

	// update docker image
	logger.Logging.Info("begin update")
	out, err := sh.Command(fileName, strconv.FormatInt(*dockerTag, 10)).Output()
	if err != nil {
		logger.Logging.Warn(err)
	}
	logger.Logging.Info(string(out))

	// register to elb
	if elb != nil {
		// get input elb names
		elbNames := getElbArray(elb)

		// init elbAction struct
		elbAction := &update.ElbAction{
			ElbNames:     elbNames,
			InstanceId:   instanceId,
			ElbSrvClient: elbSrvClient,
		}

		// Register instance from elb and check
		elbAction.RegisterInstanceToElb()
	}
	// register to alb
	if alb != nil {
		// get input alb arnNames
		albArnNames := getElbArray(alb)

		// init albAction struct
		albAction := &update.AlbAction{
			AlbArnNames:  albArnNames,
			InstanceId:   instanceId,
			AlbSrvClient: albSrvClient,
		}

		// register instance from alb and check
		albAction.RegisterInstanceToAlb()
	}

}

func checkUpdateArgs(cmd *cobra.Command, args []string) error {
	// get conf
	envParams := conf.GetEnvParams()

	// check authorisation to run this app
	if PassWord == "" {
		logger.Logging.Fatal("you must input password to run this app")
	} else if PassWord != envParams["app_password"] {
		logger.Logging.Fatal("Password error!")
	} else {
		logger.Logging.Info("password check success !")
	}

	// Check the script file exists
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			logger.Logging.WithFields(logrus.Fields{"script name": fileName}).Fatal("file not found")
		}
	} else {
		logger.Logging.WithFields(logrus.Fields{"file": fileName}).Info("find script file")
	}

	// check the deploy script's md5
	f, err := os.Open(fileName)
	if err != nil {
		logger.Logging.WithFields(logrus.Fields{"file name": fileName}).Fatal("open scripts failed ")
	}
	defer f.Close()
	r := bufio.NewReader(f)

	md5hash := md5.New()
	if _, err := io.Copy(md5hash, r); err != nil {
		return err
	}

	if hex.EncodeToString(md5hash.Sum(nil)) == hash {
		logger.Logging.WithFields(logrus.Fields{"md5sum": hash}).Info("script file md5 check success")
	} else {
		logger.Logging.WithFields(logrus.Fields{"input script's md5": hex.EncodeToString(md5hash.Sum(nil)), "need md5": hash}).Fatal("script md5 check failed")
	}

	// Check the tag is useful
	if *dockerTag <= 0 {
		logger.Logging.Fatal("you must set the right docker tag (tag > 0)")
	} else {
		logger.Logging.WithFields(logrus.Fields{"tag": *dockerTag}).Info("find docker img tag")
	}

	// elb or alb must choose one
	if elb == nil && alb == nil {
		logger.Logging.Fatal(" At least one service needs to be selected (elb or alb)")
	}
	if alb != nil {
		logger.Logging.WithFields(logrus.Fields{"alb": alb}).Info("alb will update")
	}
	if elb != nil {
		logger.Logging.WithFields(logrus.Fields{"elb": elb}).Info("elb will update")
	}

	// Check aws region
	if region == "" {
		logger.Logging.Warn("no region input , try to get from ec2")
		region = awstools.GetAwsRegion()
	}

	logger.Logging.Info("fine , prepare to update server now !")

	return nil
}

//update flag ["elb1,elb2"] to ["elb1","elb2"]
func getElbArray(flagElb []string) []string {
	var newElbArray []string
	for _, e := range strings.Split(flagElb[0], ",") {
		if len(e) < 1 {
			logger.Logging.Fatal("bad elb arrays, please check")
		}
		newElbArray = append(newElbArray, e)
	}
	return newElbArray
}

func getCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}

func init() {

	// Add String Flags
	CmdUpdate.Flags().StringVarP(&fileName, "file", "f", "", "Assign the update bash script (required)")
	CmdUpdate.Flags().StringVar(&hash, "hash", "", "Assign the script's Hash value (required)")
	CmdUpdate.Flags().StringArrayVar(&elb, "elb", nil, "Aws elb service,If Have Multiple Elb Split With ','")
	CmdUpdate.Flags().StringArrayVar(&alb, "alb", nil, "Aws alb service,If Have Multiple Elb Split With ','")
	CmdUpdate.Flags().StringVarP(&region, "region", "r", "", "AWS region")
	// Add Int Flags
	dockerTag = CmdUpdate.Flags().Int64P("tag", "t", 0, "Assign the docker image tag that need to update (required)")

	// required Flags
	CmdUpdate.MarkFlagRequired("file")
	CmdUpdate.MarkFlagRequired("hash")
	CmdUpdate.MarkFlagRequired("tag")
	CmdUpdate.MarkFlagRequired("region")

	// Reg Cmd
	// reg cmd to root
	rootCmd.AddCommand(CmdUpdate)
}
