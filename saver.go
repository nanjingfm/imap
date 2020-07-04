package mail

import (
	"crypto/md5"
	"fmt"
	"github.com/emersion/go-imap"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var DftSaver = &LocalSaver{}

type Saver interface {
	Save(data []byte) (filePath string, err error) // 保存邮件
}

type LocalSaver struct {
	BasePath string
}

func (d LocalSaver) Save(data []byte) (filePath string, err error) {
	dirPath, filePath := d.parse(data)
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return
	}
	// 保存到本地文件
	err = ioutil.WriteFile(filePath, data, 0644)
	return
}

func (d LocalSaver) parse(data []byte) (dirPath, filePath string){
	key := hashKey(data)
	dirPath = strings.TrimRight(d.BasePath, "/") + "/" + key[0:2] + "/" + key[2:4]
	filePath = dirPath + "/" + key + ".eml"
	return
}

func hashKey(data []byte) string {
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has) //将[]byte转成16进制
}

func GetBody(msg *imap.Message) []byte {
	var body []byte
	for _, value := range msg.Body {
		len := value.Len()
		buf := make([]byte, len)
		n, err := value.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if n != len {
			log.Fatal("Didn't read correct length")
		}
		body = append(body, buf...)
	}
	return body
}
