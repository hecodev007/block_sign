package rpc

import (
	"fmt"
	"github.com/spf13/viper"
)

func SuggestBrainKey() string {
	return `{"jsonrpc": "2.0", "id":"2", "method": "suggest_brain_key", "params": [] }`
}

func RegisterAccount(owner string, active string, accountname string) string {
	return fmt.Sprintf(`{"jsonrpc":"2.0", "method":"register_account", "params": ["%s","%s","%s","%s","%s",1,true], "id":"2"}`, accountname, owner, active, viper.GetString("account.viper"), viper.GetString("account.viper"))
}

func ImportKey(key string) string {
	return fmt.Sprintf(`{"jsonrpc": "2.0", "id":"2", "method": "import_key", "params": ["%s"] }`, key)
}

func GetAccount(account string) string {
	return fmt.Sprintf(`{"jsonrpc": "2.0", "id":"2", "method": "get_account", "params": ["%s"] }`, account)
}
