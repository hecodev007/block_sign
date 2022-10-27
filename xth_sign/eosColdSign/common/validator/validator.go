package validator

import (
	"errors"
	"time"
)

type ColdSign struct {
	ColdData
	ChainID   string `json:"chain_id" binding:"required"`
	EosCode   string `json:"eos_code" binding:"required"`
	Hash      string `json:"hash" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}
type ColdData struct {
	CoinName       string `json:"coinName" binding:"required"`
	MchID          string `json:"mchId" binding:"required"`
	OrderID        string `json:"orderId" binding:"required"`
	Expiration     Time   `json:"expiration" binding:"required"`
	RefBlockNum    int    `json:"ref_block_num" binding:"required"`
	RefBlockPrefix int    `json:"ref_block_prefix" binding:"required"`
	Account        string `json:"account" binding:"required"`
	Actor          string `json:"actor" binding:"required"`
	Data           string `json:"data" binding:"required"`
}
type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	formarter := "2006-01-02 15:04:05.999999999 -0700 MST"
	if y := time.Time(t).Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}
	time.Now().String()
	b := make([]byte, 0, len(formarter)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, formarter)
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *Time) UnmarshalJSON(data []byte) error {
	formarter := "2006-01-02 15:04:05.999999999 -0700 MST"
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	//var err error
	tmp, err := time.Parse(`"`+formarter+`"`, string(data))
	if err != nil {
		return err
	}
	*t = Time(tmp)
	return err
}

type ColdSignResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ColdData
		Signatures []string `json:"signatures"`
	} `json:"data"`
}
