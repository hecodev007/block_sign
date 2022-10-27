package main

//
//import (
//	"iotasign/utils/btc"
//	"fmt"
//	"log"
//
//	"github.com/shopspring/decimal"
//
//	"github.com/tealeg/xlsx"
//)
//
//func main() {
//	file := "/Users/one/Desktop/bcha_amount.xlsx"
//	ret := AnalyzeExcel(file)
//	client := btc.NewRpcClient("http://rylink:4CpmLnbOiaTbD20gPdsRYY6WiMFDyF8N8QzGGYrAfIw=@13.114.44.225:31407", "", "")
//	for _, v := range ret {
//		utxos, err := client.GetUnSpends(v[0])
//		if err != nil {
//			panic(err.Error())
//		}
//		value := decimal.NewFromInt(0)
//		for _, v := range utxos {
//			value = value.Add(v.Amount)
//		}
//		fmt.Println(v[0], v[1], value.String())
//	}
//
//}
//
//func AnalyzeExcel(path string) [][2]string {
//	xlFile, err := xlsx.OpenFile(path) //打开文件
//	if err != nil {
//		log.Println(err)
//	}
//	result := make([][2]string, 0)
//	for _, sheet := range xlFile.Sheets { //遍历sheet层
//		for rowIndex, row := range sheet.Rows { //遍历row层
//			if rowIndex > 0 {
//				if len(row.Cells) < 2 {
//					break
//				}
//				c := [2]string{"", ""}
//				result = append(result, c)
//				for cellIndex, cell := range row.Cells { //遍历cell层
//					text := cell.String() //把单元格的内容转成string
//					if len(text) == 0 {
//						break
//					}
//					result[rowIndex-1][cellIndex] = text //为数组赋值
//				}
//			}
//		}
//		break //这里直接break是因为我这个测试用的文件只有1个sheet
//	}
//
//	return result
//}
