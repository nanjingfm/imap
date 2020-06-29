package mail

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
	"sync"
	"time"
)

// 收取邮件
// 邮件解析
// 邮件过滤
// pipeline
// TODO增量更新

type Config struct {
	Addr     string `json:"addr" binding:"required"`     // 服务器地址
	Port     string `json:"port" binding:"required"`     // 端口号
	Username string `json:"username" binding:"required"` // 账号
	Password string `json:"password" binding:"required"` // 密码

}

type Filter func(*imap.Message) bool

func NewMail(c Config) *Mail {
	mail := &Mail{}
	mail.c = c
	return mail
}

type Mail struct {
	c       Config
	filters []Filter
	saveUid func(uid uint32)
}

func (p *Mail) connect() (c *client.Client, err error) {
	c, err = client.DialTLS(fmt.Sprintf("%s:%s", p.c.Addr, p.c.Port), nil)
	if err != nil {
		dftLogger.Errorf("<connectMailServer> client dial err:%v", err)
		return
	}
	c.Timeout = time.Second *5

	// 登录
	err = c.Login(p.c.Username, p.c.Password)
	if err != nil {
		dftLogger.Errorf("<connectMailServer> Login err:%v", err)
		return
	}

	// 设置字符集
	imap.CharsetReader = charset.Reader
	return
}

func (p *Mail) Scan(boxName string, lastUid uint32) <-chan *imap.Message {
	messagesCh := make(chan *imap.Message)
	ctx, cancel := context.WithCancel(context.Background())
	newestUid := uint32(0)
	msgChan := make(chan *imap.Message)

	// 收邮件
	go func() {
		err := p.fetch(ctx, boxName, 10, messagesCh)
		if err != nil {
			close(messagesCh)
		} else if p.saveUid != nil {
			p.saveUid(newestUid) // fixme:data race
		}
	}()

	// 处理邮件
	go func() {
		firstFlag := false
		for message := range messagesCh {
			if !firstFlag {
				newestUid = message.Uid
				firstFlag = true
			}
			if lastUid > 0 && message.Uid == lastUid {
				cancel()
				return
			}
			if p.filter(message) {
				msgChan <- message
			}
		}
		close(msgChan)
	}()
	return msgChan
}

func (p *Mail) fetch(stop context.Context, boxName string, limitNum uint32, messagesCh chan *imap.Message) (err error) {
	var c *client.Client
	var sw sync.WaitGroup
	defer func() {
		sw.Wait()
	}()
	seqSet := new(imap.SeqSet)

	// 如果退出登录，尝试重新登录
	c, err = p.connect()
	if err != nil {
		dftLogger.Errorf("<StartReceivedMail> connect mail server err:%v", err)
		return
	}
	// 3. 读取收件箱
	mbox, err := c.Select(boxName, true)
	if err != nil {
		dftLogger.Errorf("<StartReceivedMail> select INBOX err:%v", err)
		return
	}
	// 获取邮件总数量
	count := mbox.Messages
	for i := uint32(1); i <= count; i += limitNum {
		select {
		case <-stop.Done(): // 监听退出信号
			return
		default:
			// nothing
		}
		seqSet.Clear()
		seqSet.AddRange(i, i+limitNum)
		spew.Dump(seqSet)
		tempCh := make(chan *imap.Message)
		sw.Add(1)
		go func() {
			defer sw.Done()
			for message := range tempCh {
				messagesCh <- message
			}
		}()
		go func() {
			spew.Dump(c.State())
			time.Sleep(time.Second)
		}()
		err = c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope}, tempCh)
		if err != nil {
			dftLogger.Errorf("mail fetch failed, err: %v", err)
			return
		}
		time.Sleep(time.Second * 2)
	}
	return
}

func (p *Mail) filter(message *imap.Message) bool {
	for _, filter := range p.filters {
		if filter != nil && !filter(message) {
			return false
		}
	}
	return true
}
