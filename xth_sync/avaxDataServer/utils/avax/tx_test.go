package avax

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_tx(t *testing.T){
	rawTx :="1111111112zs9UkarRq285UVQ1boTn5dMj9VvvaMV9TgjTSGrSq4fJXqJBf7BB9mcFtG98XYDn3nCTtUdPeNL8g6xc1yeSRv76puKCXQFCxvpWfmU3Vgs8sv3TCmKVnjw6WH8ye3ZVrzBaFckvwoBUS335xhTuNRM9xKvVuEH5JVNPCiBinQoAPhhp7oYM97af4PV6oYbsi4qvz9Ln44v7XVUXzJi19H4eoM3vhMBShM5XM8etUxgs5YpZs3haoB59YagEm55zoTwZZ46yM9G4jyKFxoeUfX3Xj5yZGEQ5PmuHidViQG5aETnx7qWjUXtLrmbEZKDoYHiqPg6upRt1D5J2BX5gZmP7cTLosonmnp1qdpBUi9JiLy4N4hHCpYT3W78uBgvN2UWbVdmkDPNWKNP6rpJLJJaytrSie6ABR2Ar4ofsD6bWbGwnMfnYxADh7wYMxQSexsqrymEj6UPXGjKoHHSEbk7x3yyfLT2yS1do5RdiLAxLTJuZyKo9Jtz8m9gXe8MCw5FXeF7QdaXUWyxXkxb4A4uFtCjheQPoNLArzxL24LvYQ7nYnXdpPQKQ6hPeoBFrjBveuttshhDLbZnMK4iQpzEUDfkFd9kEeeSc51YCu3Ny"
	avmtx, err := ParseTx(rawTx)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	//_ = tx
	txjson, _ := json.Marshal(avmtx)
	tx := new(Tx)
	err = json.Unmarshal(txjson,tx)
	fmt.Printf("%+v",tx)
	fmt.Println(string(txjson))
}
func Test_gettx(t *testing.T){
	cli := NewRpcClient("","","","")
	tx,err := cli.GetRawTransactionFromScan("KfGvJVBw346Nxe3eJegv8Kgmed5HhR9LrJLeZCB6kxBWHDmnb")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(tx))
}
//{"unsignedTx":{"networkID":1,"blockchainID":"2oYMBNV4eNHyqk2fjjV5nVQLDbtmNJzq5s3qs3Lo6ftnC6FByM","outputs":[{"assetID":"FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z","output":{"amount":101500001,"locktime":0,"threshold":1,"addresses":["CJbq1qfsJdkgqYdKMVNu3DHjWA8NuQjrW"]}}],"inputs":[{"txID":"Q5TnfEgrZH3omv4WGMQiXrYSva8kd8vvq4Pi1PReg5TjbmA7V","outputIndex":1,"assetID":"FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z","input":{"amount":2500001,"signatureIndices":[0]}},{"txID":"2WdTwj1KyK9No4G5vd8RBKRpJweC8Yb2ooC9okm1pcX8QBE6Ac","outputIndex":0,"assetID":"FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z","input":{"amount":100000000,"signatureIndices":[0]}}],"memo":""},"credentials":[{"signatures":["3WCsfDJidsxpcYBkZa6dyov5WRqa8cAXsrB6zRdSzQRvzB2rHRLzZnUSKkuDiR6LQg8Lev4tc2ByLBqGTEH3u7PKpCgy7W1"]},{"signatures":["3WckyT9mxwwN7wtw6yQ91xCUc42fiFWmJvjuZu2qF7cTajbLaKgRWi8V3kN6Qwm7ESpBcCUYSsoRFNPzPQw55wGUScoKgwz"]}]}