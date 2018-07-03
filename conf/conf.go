package conf

import (
	"os"
	"strings"
)

func GetEnvParams() map[string]string {
	envParams := map[string]string{
		"app_env":         "product", // use "dev" or "product"
		"app_password":    "",
		"region_url":      "",
		"instance_id_url": "",
		"dingDingUrl": "",
	}

	//如果在环境变量中检测到有envParams中的key,则将envParams中key相关的value替换为环境变量中的值
	//example: APP_ENV=
	for k := range envParams {
		if v := os.Getenv(strings.ToUpper(k)); v != "" {
			envParams[k] = v
		}
	}

	//开发环境，使用key访问
	if envParams["app_env"] == "dev" {
		envParams["aws_id"] = ""
		envParams["aws_key"] = ""
	}

	return envParams
}
