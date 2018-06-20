package logger

import (
	"awselb/conf"
	"bytes"
	"fmt"
	"net/http"
)

func NewBot(elbName string, instanceId string, tryTime int) error {

	if tryTime == 10 || tryTime == 20 {
		envParams := conf.GetEnvParams()
		postData := `
        {
            "msgtype": "markdown",
            "markdown": {"title":"EC2服务更新阻塞","text": "#### EC2服务更新阻塞 \n ##### EC2: %s \n ##### ELB(ALB): %s \n ##### 阻塞次数: %d "}
        }`
		body := fmt.Sprintf(postData, instanceId, elbName, tryTime)
		jsonValue := []byte(body)
		//发送消息到钉钉群
		resp, err := http.Post(envParams["dingDingUrl"], "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			return err
		}
		Logging.Warn(resp)
	}

	return nil
}
