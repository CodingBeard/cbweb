package cbweb

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

type ErrorHandler interface {
	Error(e error)
	Recover()
}

type DefaultErrorHandler struct{}

func (d DefaultErrorHandler) Error(e error) {
	buf := make([]byte, 1000000)
	runtime.Stack(buf, false)
	buf = bytes.Trim(buf, "\x00")
	stack := string(buf)
	stackParts := strings.Split(stack, "\n")
	newStackParts := []string{stackParts[0]}
	newStackParts = append(newStackParts, stackParts[3:]...)
	stack = strings.Join(newStackParts, "\n")
	log.Println("ERROR", e.Error()+"\n"+stack)
}

func (d DefaultErrorHandler) Recover() {
	e := recover()

	if e != nil {
		err, ok := e.(error)

		if ok {
			d.Error(err)
		} else {
			d.Error(errors.New(fmt.Sprint(e)))
		}
	}
}

type CacheProvider interface {
	Get(key string) (interface{}, bool)
	Delete(key string)
	Set(key string, value interface{}, ttl time.Duration)
}