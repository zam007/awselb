# awselb
aws elb tools 用于从elb/alb摘除EC2后执行docker更新脚本，最后再注册回elb/alb，配合autoscaling达到roling update service的目的。

相应权限需要配合IAM使用

#### 用法

```
 ./awselb update \
 -p passwd \
 -f /home/worker/update_docker.sh \
 --hash xxxxxxxxxxxxx \
 -t 20 \
 --alb alb-service \
 --elb elb-service
 -r cn-north-1 
``` 
##### 参数说明 
 ```
 -p APP的密码，防止用户直接在服务器上执行该应用
 -f 更新docker服务的脚本
 --hash 脚本的hash值，防止脚本意外变更后被执行
 -t docker更新脚本会接收一个参数,为docker img 的tag号
 --alb 需要更新的ALB名称，逗号分隔
 --elb 需要更新ELB的名称，逗号分隔
 -r aws region
 ```


