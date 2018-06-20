package update

import (
	"awselb/awstools"
	"awselb/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type ElbAction struct {
	//Region       string
	ElbNames     []string
	InstanceId   string
	ElbSrvClient *elb.ELB
}

func GetElbSrvClient(region string) *elb.ELB {
	// init elb service client
	sess := session.Must(session.NewSession())
	return elb.New(sess, awstools.GetAwsConfig(region))
}

func (e *ElbAction) DeRegisterInstanceFromElb() {
	for _, elbName := range e.ElbNames {

		var ec2InElb []string // ec2 in elbName
		// set deRegister flag default "succeed", if find this instance still in elb set flag to "failed"
		var delState string
		delState = "succeed"

		// check instance State before deRegister
		nowState := describeElbInstanceHealthStatus(elbName, e.InstanceId, e.ElbSrvClient)
		if nowState != "InService" {
			logger.Logging.WithFields(logrus.Fields{"elb": elbName, "instance-id": e.InstanceId, "instance-state": nowState}).Warn("instance not in service")
			continue
		}

		// deRegister instance from elb
		deRegisterParams := &elb.DeregisterInstancesFromLoadBalancerInput{
			Instances: []*elb.Instance{
				{
					InstanceId: aws.String(e.InstanceId),
				},
			},
			LoadBalancerName: aws.String(elbName),
		}
		deRegisterResp, err := e.ElbSrvClient.DeregisterInstancesFromLoadBalancer(deRegisterParams)
		if err != nil {
			logger.Logging.WithFields(logrus.Fields{"deRegister error": err}).Warn("deRegister ec2 from elb failed")
		}
		for k := range deRegisterResp.Instances {
			ec2InElb = append(ec2InElb, *deRegisterResp.Instances[k].InstanceId)
		}

		if len(ec2InElb) > 0 {
			for _, value := range ec2InElb {
				if e.InstanceId == value {
					delState = "failed"
					break
				}
			}
		}

		// 循环检测实例相对于该ELB的状态，直到deRegister成功
		var testCounts int // deRegister times counts
		testCounts = 1
		for {
			time.Sleep(5 * time.Second)

			// 如果阻塞,发送信息到钉钉告警
			logger.NewBot(elbName, e.InstanceId, testCounts)

			// get instance State
			instanceState := describeElbInstanceHealthStatus(elbName, e.InstanceId, e.ElbSrvClient)

			// deRegister and describeHealthy must both Pass
			if strings.TrimSpace(delState) == "succeed" && strings.TrimSpace(instanceState) == "OutOfService" {
				logger.Logging.WithFields(logrus.Fields{"elb": elbName, "instance-id": e.InstanceId, "instance-state": instanceState}).Info("deRegister ec2 from elb succeed")
				break
			}

			logger.Logging.WithFields(logrus.Fields{"elb": elbName, "delState": delState, "instanceState": instanceState, "deRegisterTimeCounts": testCounts}).Info("Deleting")
			testCounts = testCounts + 1
		}
	}
}

func (e *ElbAction) RegisterInstanceToElb() {
	for _, elbName := range e.ElbNames {

		var ec2InElb []string // ec2 in elbName

		// check instance State before deRegister
		nowState := describeElbInstanceHealthStatus(elbName, e.InstanceId, e.ElbSrvClient)
		if nowState != "OutOfService" {
			logger.Logging.WithFields(logrus.Fields{"elb": elbName, "instance-id": e.InstanceId, "instance-state": nowState}).Warn("instance is in service now")
			continue
		}

		//register instance from elb
		registerParams := &elb.RegisterInstancesWithLoadBalancerInput{
			Instances: []*elb.Instance{
				{
					InstanceId: aws.String(e.InstanceId),
				},
			},
			LoadBalancerName: aws.String(elbName),
		}
		registerResp, err := e.ElbSrvClient.RegisterInstancesWithLoadBalancer(registerParams)
		if err != nil {
			logger.Logging.WithFields(logrus.Fields{"Register error": err}).Warn("Register ec2 to elb failed")
		}
		for k := range registerResp.Instances {
			ec2InElb = append(ec2InElb, *registerResp.Instances[k].InstanceId)
		}

		// set register flag default "failed", if find in elb, set flag to "succeed"
		var regState string
		regState = "failed"
		for _, value := range ec2InElb {
			if e.InstanceId == value {
				regState = "succeed"
				break
			}
		}

		// 循环检测实例相对于ELB的状态直到register成功
		var testCounts int // deRegister times counts
		testCounts = 1
		for {
			time.Sleep(5 * time.Second)

			// 如果阻塞,发送信息到钉钉告警
			logger.NewBot(elbName, e.InstanceId, testCounts)

			// get  instance State
			instanceState := describeElbInstanceHealthStatus(elbName, e.InstanceId, e.ElbSrvClient)

			// register and describeHealthy must both Pass
			if strings.TrimSpace(regState) == "succeed" && strings.TrimSpace(instanceState) == "InService" {
				logger.Logging.WithFields(logrus.Fields{"elb": elbName, "instance-id": e.InstanceId, "instance-state": instanceState}).Info("register ec2 to elb succeed")
				break
			}

			logger.Logging.WithFields(logrus.Fields{"elb": elbName, "regState": regState, "instanceState": instanceState, "registerTimeCounts": testCounts}).Info("Try Reg Elb Now")
			testCounts = testCounts + 1
		}
	}
}

func describeElbInstanceHealthStatus(elbName, instanceId string, elbSrvClient *elb.ELB) string {
	describeParams := &elb.DescribeInstanceHealthInput{
		LoadBalancerName: aws.String(elbName),
		Instances: []*elb.Instance{
			{
				InstanceId: aws.String(instanceId),
			},
		},
	}

	describeResp, err := elbSrvClient.DescribeInstanceHealth(describeParams)
	if err != nil {
		logger.Logging.Warn("describe elb instance health failed")
		return ""
	}

	return *describeResp.InstanceStates[0].State
}
