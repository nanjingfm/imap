package mail

import (
	"fmt"
	"github.com/emersion/go-imap"
	"testing"
)

var testConfig = Config{
	Addr:     "imap.exmail.qq.com",
	Port:     "993",
	Username: "xxx",
	Password: "xxxx",
}

func TestMail_Scan(t *testing.T) {
	mail := NewMail(testConfig)
	mail.AddSaver(&LocalSaver{BasePath: "/Users/xman/code/data"})
	ch := mail.Scan(imap.InboxName, 0)
	count := 0
	for range ch {
		count++
		fmt.Println(count)
	}
}
