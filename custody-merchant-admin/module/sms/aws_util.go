package sms

import (
	"custody-merchant-admin/module/log"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sns"
)

type AwsConfig struct {
	Region      string //aws区域
	AccessKeyId string //aws访问密钥id
	SecretKey   string //aws访问密钥
}

//EmailData
//邮件
type EmailData struct {
	Recipient string //接收者
	Body      string //正文
	Subject   string //主题
	Sender    string //发送者
	CharSet   string //编码
}

type EmailConfig struct {
	*AwsConfig
	*EmailData
}

var EmailConf *EmailConfig

//新建一个邮件服务
func NewEmailService(ec *AwsConfig) *EmailConfig {
	EmailConf = &EmailConfig{
		AwsConfig: ec,
		EmailData: nil,
	}
	return EmailConf
}

//发送邮件信息
func (ec *EmailConfig) SendEmail(ecData *EmailData) error {
	cfgs := &aws.Config{
		Region: aws.String(ec.Region),
	}
	if ec.AccessKeyId != "" && ec.SecretKey != "" {
		cfgs.Credentials = credentials.NewStaticCredentials(ec.AccessKeyId, ec.SecretKey, "")
	}
	sess, err := session.NewSession(cfgs)
	if err != nil {
		return err
	}

	svc := ses.New(sess)
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{aws.String(ecData.Recipient)},
			ToAddresses: []*string{aws.String(ecData.Recipient)},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(ecData.CharSet),
					Data:    aws.String(ecData.Body),
				},
				Text: &ses.Content{
					Charset: aws.String(ecData.CharSet),
					Data:    aws.String(ecData.Body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(ecData.CharSet),
				Data:    aws.String(ecData.Body),
			},
		},
		Source: aws.String(ecData.Sender),
	}

	if _, err := svc.SendEmail(input); err != nil {
		return err
	}
	return nil
}

//短信
type SnsData struct {
	Recipient string //接收者
	Body      string //正文
}

type SnsConfig struct {
	*AwsConfig
	*SnsData
}

var SnsConf *SnsConfig

//NewSnsService
//新建一个短信通知服务
func NewSnsService(ec *AwsConfig) *SnsConfig {
	SnsConf = &SnsConfig{
		AwsConfig: ec,
	}
	return SnsConf
}

//SendSns
//发送短信通知
func (sc *SnsConfig) SendSns(snsData SnsData) error {
	cfgs := &aws.Config{
		Region: aws.String(sc.Region),
	}
	if sc.AccessKeyId != "" && sc.SecretKey != "" {
		cfgs.Credentials = credentials.NewStaticCredentials(sc.AccessKeyId, sc.SecretKey, "")
	}
	sess, err := session.NewSession(cfgs)
	if err != nil {
		return err
	}
	svc := sns.New(sess)
	input := &sns.PublishInput{
		Message:     aws.String(snsData.Body),
		PhoneNumber: aws.String(snsData.Recipient),
	}
	out, err := svc.Publish(input)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	fmt.Println(out)
	return nil
}

func (aws *AwsConfig) sendSns(recipient, msg string) (err error) {
	newSnsService := NewSnsService(aws)
	if err = newSnsService.SendSns(SnsData{
		Recipient: recipient,
		Body:      msg,
	}); err != nil {
		return
	}
	return
}
