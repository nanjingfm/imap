package mail

import (
	"fmt"
	"github.com/emersion/go-imap"
	"testing"
	"time"
)

var testConfig = Config{
	Addr:     "imap.exmail.qq.com",
	Port:     "993",
	Username: "fuming@lanjingren.com",
	Password: "NNJfm1",
}

func TestMail_Scan(t *testing.T) {
	mail := NewMail(testConfig)
	ch := mail.Scan(imap.InboxName, 0)
	for _ = range ch {
		fmt.Print("1")
	}
	time.Sleep(time.Second * 3)
}
