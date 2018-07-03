package update

import (
	"awselb/awstools"
	"awselb/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type AlbAction struct {
	//Region       string
	AlbArnNames  []string
	InstanceId   string
	AlbSrvClient *elbv2.ELBV2
}

func GetAlbSrvClient(region string) *elbv2.ELBV2 {
	// init alb service client
	sessn := session.Must(session.NewSession())
	return elbv2.New(sessn, awstools.GetAwsConfig(region))
}

func (a *AlbAction) DeRegisterInstanceFromAlb() {
	for _, targetGroupName := range a.AlbArnNames {
		// set deRegister flag
		var delState string
		delState = "succeed"

		// get targetListeningPort and targetGroupArn
		targetMetaData := getTargetGroupMetaData(targetGroupName, a.AlbSrvClient)
		targetListeningPort := targetMetaData.Port
		targetGroupArn := targetMetaData.TargetGroupArn

		// check instance State before deRegister
		nowInstanceState := describeAlbInstanceHealthyStatus(*targetGroupArn, a.InstanceId, targetListeningPort, a.AlbSrvClient)
		if nowInstanceState != "healthy" {
			logger.Logging.WithFields(logrus.Fields{"alb": targetGroupName, "instance-id": a.InstanceId, "targetPort": *targetListeningPort, "instance-state": nowInstanceState}).Warn("instance not in service")
			continue
		}

		// deRegister instance from targetGroups
		deRegisterParams := &elbv2.DeregisterTargetsInput{
			TargetGroupArn: targetGroupArn,
			Targets: []*elbv2.TargetDescription{
				{
					Id: aws.String(a.InstanceId),
				},
			},
		}
		deRegisterResp, err := a.AlbSrvClient.DeregisterTargets(deRegisterParams)
		if err != nil {
			logger.Logging.WithFields(logrus.Fields{"deRegister error": err, "DeRegisterTargetResp": deRegisterResp.GoString()}).Warn("deRegister ec2 from alb failed")
			delState = "failed"
		}

		// 循环检测直到deRegister成功
		var testCounts int // deRegister times counts
		testCounts = 1
		for {
			time.Sleep(5 * time.Second)

			// 如果阻塞,发送信息到钉钉告警
			logger.NewBot(targetGroupName, a.InstanceId, testCounts)

			// get instance State
			instanceState := describeAlbInstanceHealthyStatus(*targetGroupArn, a.InstanceId, targetListeningPort, a.AlbSrvClient)

			// deRegister and describeHealthy must both Pass
			if strings.TrimSpace(delState) == "succeed" && strings.TrimSpace(instanceState) == "unused" {
				logger.Logging.WithFields(logrus.Fields{"targetGroup:": targetGroupName, "instance-id": a.InstanceId, "instance-state": instanceState}).Info("deRegister ec2 from alb succeed")
				break
			}

			logger.Logging.WithFields(logrus.Fields{"targetGroup": targetGroupName, "delState": delState, "instanceState": instanceState, "deRegisterTimeCounts": testCounts}).Info("Try Drop ", a.InstanceId, " From ", targetGroupName)
			testCounts = testCounts + 1
		}
	}
}

func (a *AlbAction) RegisterInstanceToAlb() {
	for _, targetGroupName := range a.AlbArnNames {
		// set deRegister flag
		var regAlbState string
		regAlbState = "succeed"

		// get targetListeningPort and targetGroupArn
		targetMetaData := getTargetGroupMetaData(targetGroupName, a.AlbSrvClient)
		targetListeningPort := targetMetaData.Port
		targetGroupArn := targetMetaData.TargetGroupArn

		// check instance State before deRegister
		nowInstanceState := describeAlbInstanceHealthyStatus(*targetGroupArn, a.InstanceId, targetListeningPort, a.AlbSrvClient)
		if nowInstanceState != "unused" {
			logger.Logging.WithFields(logrus.Fields{"alb": targetGroupName, "instance-id": a.InstanceId, "targetPort": *targetListeningPort, "instance-state": nowInstanceState}).Warn("instance in service")
			continue
		}

		// register instance to targetGroups
		registerParams := &elbv2.RegisterTargetsInput{
			TargetGroupArn: targetGroupArn,
			Targets: []*elbv2.TargetDescription{
				{
					Id:   aws.String(a.InstanceId),
					Port: targetListeningPort,
				},
			},
		}
		registerResp, err := a.AlbSrvClient.RegisterTargets(registerParams)
		if err != nil {
			logger.Logging.WithFields(logrus.Fields{"Register error": err, "RegisterTargetResp": registerResp.GoString()}).Warn("Register ec2 to alb failed")
			regAlbState = "failed"
		}

		// 循环检测直到register成功
		var testCounts int // deRegister times counts
		testCounts = 1
		for {
			time.Sleep(5 * time.Second)

			// 如果阻塞,发送信息到钉钉告警
			logger.NewBot(targetGroupName, a.InstanceId, testCounts)

			// get instance State
			instanceState := describeAlbInstanceHealthyStatus(*targetGroupArn, a.InstanceId, targetListeningPort, a.AlbSrvClient)

			// Register and describeHealthy must both Pass
			if strings.TrimSpace(regAlbState) == "succeed" && strings.TrimSpace(instanceState) == "healthy" {
				logger.Logging.WithFields(logrus.Fields{"targetGroup": targetGroupName, "instance-id": a.InstanceId, "instance-state": instanceState}).Info("Register ec2 to alb succeed")
				break
			}

			logger.Logging.WithFields(logrus.Fields{"targetGroup": targetGroupName, "regAlbState": regAlbState, "instanceState": instanceState, "RegisterTimeCounts": testCounts}).Info("Try Reg ", a.InstanceId, " To ", targetGroupName)
			testCounts = testCounts + 1
		}

	}
}

func describeAlbInstanceHealthyStatus(albArnName, instanceId string, targetPort *int64, albSrvClient *elbv2.ELBV2) string {
	describeParams := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(albArnName),
		Targets: []*elbv2.TargetDescription{
			{
				Id:   aws.String(instanceId),
				Port: aws.Int64(*targetPort),
			},
		},
	}

	describeResp, err := albSrvClient.DescribeTargetHealth(describeParams)
	if err != nil {
		logger.Logging.Warn("describe alb instance health failed")
		return ""
	}

	return *describeResp.TargetHealthDescriptions[0].TargetHealth.State
}

func getTargetGroupMetaData(targetGroupName string, albSrvClient *elbv2.ELBV2) *targetGroupMetaData {
	var targetGroupNames []*string
	targetGroupNames = append(targetGroupNames, aws.String(targetGroupName))

	describeParams := &elbv2.DescribeTargetGroupsInput{
		Names: targetGroupNames,
	}
	describeResp, err := albSrvClient.DescribeTargetGroups(describeParams)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeLoadBalancerNotFoundException:
				logger.Logging.Warn(elbv2.ErrCodeLoadBalancerNotFoundException, aerr.Error())
			case elbv2.ErrCodeTargetGroupNotFoundException:
				logger.Logging.Warn(elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			default:
				logger.Logging.Fatal(aerr.Error())
			}
			//logger.Logging.Warn("describe targetGroups metaData failed", err)
		}
	}

	metaData := &targetGroupMetaData{
		Port:           describeResp.TargetGroups[0].Port,
		TargetGroupArn: describeResp.TargetGroups[0].TargetGroupArn,
	}

	return metaData
}

type targetGroupMetaData struct {
	// The port on which the targets are listening.
	Port           *int64
	TargetGroupArn *string
}
