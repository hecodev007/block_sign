package emails

import (
	"custody-merchant-admin/module/log"
	"gopkg.in/gomail.v2" //go get gopkg.in/gomail.v2
)

const (
	// HtmlBody
	// The HTML body for the emails.
	HtmlBody = "<html><head><title>HOO 商户申请提交</title></head><body>" +
		"<h1>HOO Email</h1>" +
		"<p>This emails was sent with " +
		"<a href='https://aws.amazon.com/ses/'>Amazon SES</a> using " +
		"the <a href='https://github.com/go-gomail/gomail/'>Gomail " +
		"package</a> for <a href='https://golang.org/'>Go</a>.</p>" +
		"</body></html>"

	// Tags
	// The tags to apply to this message. Separate multiple key-value pairs
	// with commas.
	// If you comment out or remove this variable, you will also need to
	// comment out or remove the header on line 80.
	Tags = "genre=test,genre2=test2"
)

type EmailConfig struct {
	IamUserName  string `json:"iam_user_name"`
	Recipient    string `json:"recipient"`
	SmtpUsername string `json:"smtp_username"`
	SmtpPassword string `json:"smtp_password"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Title        string `json:"title"`
}

func (emailConfig *EmailConfig) SendEmail(TextBody string) (bool, error) {

	// Create a new message.
	m := gomail.NewMessage()

	// Set the main emails part to use HTML.
	//m.SetBody("text/html", HtmlBody)

	// Set the alternative part to plain text.
	m.AddAlternative("text/plain", TextBody)

	// Construct the message headers, including a Configuration Set and a Tag.
	m.SetHeaders(map[string][]string{
		"From":    {m.FormatAddress(emailConfig.IamUserName, emailConfig.IamUserName)},
		"To":      {emailConfig.Recipient},
		"Subject": {emailConfig.Title},
		// Comment or remove the next line if you are not using a configuration set
		//"X-SES-CONFIGURATION-SET": {ConfigSet},
		// Comment or remove the next line if you are not using custom tags
		"X-SES-MESSAGE-TAGS": {Tags},
	})

	// Send the emails.
	d := gomail.NewPlainDialer(emailConfig.Host, emailConfig.Port, emailConfig.SmtpUsername, emailConfig.SmtpPassword)

	// Display an error message if something goes wrong; otherwise,
	// display a message confirming that the message was sent.
	if err := d.DialAndSend(m); err != nil {
		log.Error(err)
		return false, err
	} else {
		return true, nil
	}
}
