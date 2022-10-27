package net

import "fmt"

const (
	KYT_URL = "https://api.chainalysis.com/api/kyt/v1"
	//KYT_URL = "https://test.chainalysis.com/api/kyt/v1"
)

func TransferSent(uid string) string {
	return fmt.Sprintf("%s/users/%s/transfers/sent", KYT_URL, uid)
}

func TransferReceived(uid string) string {
	return fmt.Sprintf("%s/users/%s/transfers/received", KYT_URL, uid)
}

func SupportedAssets() string {
	return fmt.Sprintf("%s/assets", KYT_URL)
}
