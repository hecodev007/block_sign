package xkutils

import (
	"bytes"
	"fmt"
)

type StringBuilder struct {
	Builder bytes.Buffer
}

func NewStringBuilder() *StringBuilder {
	build := new(StringBuilder)
	return build
}

func (s *StringBuilder) StringBuild(format string, values ...interface{}) *StringBuilder {
	str := fmt.Sprintf(format, values...)
	s.Builder.WriteString(str)
	return s
}

func (s *StringBuilder) AddString(format string) *StringBuilder {
	s.Builder.WriteString(format)
	return s
}

func (s *StringBuilder) ToString() string {
	return s.Builder.String()
}
