package models

type GasOutput struct {
	HttpResult GasHttpResult `json:"httpResult"`
	GasFee     GasFee        `json:"gasFee"`
	Suggested  Suggested     `json:"suggested"`
}

//请求结果
type GasHttpResult struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//计算结果
type GasFee struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}

//建议结果，向上取整
type Suggested struct {
	FastestFee  int64
	HalfHourFee int64
	HourFee     int64
}
