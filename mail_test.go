package mail

import (
	"fmt"
	"github.com/emersion/go-imap"
	"testing"
)

var testConfig = Config{
	Addr:     "imap.exmail.qq.com",
	Port:     "993",
	Username: "xxxx",
	Password: "xxx",
}

func TestMail_Scan(t *testing.T) {
	mail := NewMail(testConfig)
	ch := mail.Scan(imap.InboxName, 0)
	for range ch {
		fmt.Println("*")
	}
}
