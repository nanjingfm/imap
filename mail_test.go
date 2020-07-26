package mail

import (
	"context"
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
	mail.AddSaver(&LocalSaver{BasePath: "/Users/hades/code/data"})
	ch := mail.Scan(context.TODO(), imap.InboxName)
	count := 0
	for range ch {
		count++
		fmt.Println(count)
	}
}
