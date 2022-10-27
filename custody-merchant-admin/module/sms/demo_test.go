package sms

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/xluohome/phonedata"
	"testing"
)

func TestPhoneFrom(t *testing.T) {
	info, err := phonedata.Find("18077700154")
	if err != nil {
		panic(err)
	}
	fmt.Println("归属地:", info)
	fmt.Println("省：", info.Province)
	fmt.Println("市：", info.City)
}

func TestHttpSms(t *testing.T) {
	sms, err := NewSms("I10271", "lA594d", "1000", "http://39.97.4.102:9090/sms/batch/impl").SendSms("+855 718079625", "[HOO] code: 111111. Valid for 5 minutes.")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(sms)
}

func TestAutoSns(t *testing.T) {
	newSnsService := NewSnsService(&AwsConfig{
		Region:      "ap-northeast-1",
		AccessKeyId: "AKIAZXPISQLLNKJRB5MF",
		SecretKey:   "qYh76D2Zc/+BpviyUkrvUxbxwYKdkFSFKNDpuCH1",
	})
	if err := newSnsService.SendSns(SnsData{
		Recipient: "+1 8253050689",
		Body:      "Use 999 854 to verify your Instagram account.",
	}); err != nil {
		return
	}
	return
}

func TestSNS(t *testing.T) {
	// 区域
	region := "ap-northeast-1"
	// 消息
	message := "9999"
	// 主题
	//subject := "fidding"
	// sns主题Arn
	//topicArn := "arn:aws-cn:sns:cn-northwest-1:638953167227:test"

	// Initial credentials loaded from SDK's default credential chain. Such as
	// the environment, shared credentials (~/.aws/credentials), or EC2 Instance
	// Role. These credentials will be used to to make the STS Assume Role API.
	sess := session.Must(session.NewSession())

	// Create a SNS client with additional configuration
	// 方式一: 使用文件证书位置默认为~/.aws/credentials
	//svc := sns.New(sess, aws.NewConfig().WithRegion(region))
	// 方式二: 使用传参方式
	creds := credentials.NewStaticCredentials(
		"AKIAZXPISQLLNKJRB5MF",
		"qYh76D2Zc/+BpviyUkrvUxbxwYKdkFSFKNDpuCH1",
		"",
	)
	svc := sns.New(sess, &aws.Config{Credentials: creds, Region: aws.String(region)})
	phone := "+63 995-595-4943"
	// 发送sns请求
	params := sns.PublishInput{
		Message: &message,
		//Subject: &subject,
		//TargetArn: &topicArn,
		PhoneNumber: &phone,
	}
	req, resp := svc.PublishRequest(&params)
	err := req.Send()
	if err == nil {
		fmt.Println(resp)
	} else {
		fmt.Print(req)
	}
}
