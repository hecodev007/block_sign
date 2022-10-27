package validator
//gin > 1.4.0
//将验证器错误翻译成中文
import (
	"github.com/go-playground/locales/zh"

	"github.com/gin-gonic/gin/binding"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	translation "github.com/go-playground/validator/v10/translations/zh"
)
var (
	trans ut.Translator
)
func init(){
	trans, _ = ut.New(zh.New()).GetTranslator("zh")
	translation.RegisterDefaultTranslations(binding.Validator.Engine().(*validator.Validate), trans)
}
func Error(err error) (ret string){
	if validationErrors ,ok := err.(validator.ValidationErrors);!ok {
	 	return err.Error()
	 } else {
		for _, e := range validationErrors{
			ret += e.Translate(trans)+";"
		}
	}
	return ret
}