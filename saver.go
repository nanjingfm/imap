package mail

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var DftSaver = &LocalSaver{}

type Saver interface {
	Save(message *imap.Message) error // 保存邮件
}

type LocalSaver struct {

}

func (d LocalSaver) Save(message *imap.Message) error {
	data, _ := json.Marshal(message.Format())
	dirPath, filePath := parse(data)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}
	// 保存到本地文件
	err = ioutil.WriteFile(filePath, data, 0644)
	return err
}

func parse(data []byte) (dirPath, filePath string){
	key := hashKey(data)
	dirPath = key[0:2] + "/" + key[2:4]
	filePath = dirPath + "/" + key + ".eml"
	return
}

func hashKey(data []byte) string {
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has) //将[]byte转成16进制
}

func aa(msg *imap.Message)  {
	var section imap.BodySectionName
	r := msg.GetBody(&section)
	//fmt.Print(message.Format())
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Fatal(err)
	}
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			b, _ := ioutil.ReadAll(p.Body)
			log.Printf("Got text: %s\n", string(b))
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			log.Printf("Got attachment: %s\n", filename)
		}
	}
}
