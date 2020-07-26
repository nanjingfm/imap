package mail

import (
	"bytes"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestEml_Parse(t *testing.T) {
	testEml := "/Users/hades/Downloads/444444.eml"
	file, err := ioutil.ReadFile(testEml)
	assert.Nil(t, err)
	f := bytes.NewReader(file)
	email, err := Parse(f)
	assert.Nil(t, err)
	spew.Dump(email)
}
