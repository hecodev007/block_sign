package providers


type ProviderInterface interface {
	SendRequest(v interface{}, method string, params interface{}) error
	Close() error
}