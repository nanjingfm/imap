package mail

import "github.com/davecgh/go-spew/spew"

var dftLogger Logger = &FmtLogger{}

type Logger interface {
	Errorf(...interface{})
}

type FmtLogger struct {
}

func (f *FmtLogger) Errorf(i ...interface{}) {
	spew.Dump(i...)
}

