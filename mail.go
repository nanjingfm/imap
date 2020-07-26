package mail

import (
	"context"
	"fmt"
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
	Addr      string `json:"addr" binding:"required"`     // 服务器地址
	Port      string `json:"port" binding:"required"`     // 端口号
	Username  string `json:"username" binding:"required"` // 账号
	Password  string `json:"password" binding:"required"` // 密码
	StorePath string `json:"store_path"`                  // eml文件保存路径
}

type Filter func(*imap.Message) bool

func NewMail(c Config) *Mail {
	mail := &Mail{}
	mail.c = c
	//mail.saveUid = func(uid uint32) {
	//	fmt.Println(uid)
	//}
	return mail
}

type Mail struct {
	c       Config
	filters []Filter
	savers  []Saver
	saveUid func(uid uint32)
}

func (p *Mail) AddFilter(filter ...Filter) {
	p.filters = append(p.filters, filter...)
}

func (p *Mail) connect() (c *client.Client, err error) {
	c, err = client.DialTLS(fmt.Sprintf("%s:%s", p.c.Addr, p.c.Port), nil)
	if err != nil {
		dftLogger.Errorf("<connectMailServer> client dial err:%v", err)
		return
	}

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

func (p *Mail) Scan(pCtx context.Context, boxName string) <-chan *imap.Message {
	messagesCh := make(chan *imap.Message)
	ctx, cancel := context.WithCancel(pCtx)
	newestUid := uint32(0)
	msgChan := make(chan *imap.Message)

	// 收邮件
	go func() {
		err := p.fetch(ctx, boxName, 10, messagesCh)
		if err == nil && p.saveUid != nil {
			p.saveUid(newestUid) // fixme:data race
		}
		close(messagesCh)
	}()

	// 处理邮件
	go func() {
		firstFlag := false
		for message := range messagesCh {
			if !firstFlag {
				newestUid = message.Uid
				firstFlag = true
			}
			select {
			case <-ctx.Done():
				cancel()
				break
			default:
				// nothing
			}
			if p.filter(message) {
				go p.save(message)
				msgChan <- message
			}
		}
		close(msgChan)
	}()
	return msgChan
}

func (p *Mail) AddSaver(saver Saver) {
	p.savers = append(p.savers, saver)
}

func (p *Mail) save(msg *imap.Message) {
	for _, saver := range p.savers {
		if saver != nil {
			_, _ = saver.Save(GetBody(msg))
		}
	}
}

func (p *Mail) fetch(stop context.Context, boxName string, limitNum uint32, messagesCh chan *imap.Message) (err error) {
	var c *client.Client
	var sw sync.WaitGroup
	defer func() {
		sw.Wait()
	}()

	// 如果退出登录，尝试重新登录
	c, err = p.connect()
	if err != nil {
		dftLogger.Errorf("<StartReceivedMail> connect mail server err:%v", err)
		return
	}
	defer c.Logout()
	// 3. 读取收件箱
	mbox, err := c.Select(boxName, false)
	if err != nil {
		dftLogger.Errorf("<StartReceivedMail> select INBOX err:%v", err)
		return
	}
	// 获取邮件总数量
	count := mbox.Messages
	for count > 0 {
		select {
		case <-stop.Done(): // 监听退出信号
			return
		default:
			// nothing
		}

		seqSet := new(imap.SeqSet)
		if count > limitNum {
			seqSet.AddRange(count-limitNum+1, count)
			count = count - limitNum
		} else {
			seqSet.AddRange(1, count)
			count = 0
		}
		tempCh := make(chan *imap.Message)
		sw.Add(1)
		go func() {
			defer sw.Done()
			bf := make([]*imap.Message, 0, limitNum)
			for message := range tempCh {
				bf = append(bf, message)
			}
			l := len(bf)
			for i := 0; i < l; i++ {
				messagesCh <- bf[l-i-1]
			}
		}()
		var section imap.BodySectionName
		err = c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, section.FetchItem()}, tempCh)
		if err != nil {
			dftLogger.Errorf("mail fetch failed, err: %v", err)
			return
		}
		time.Sleep(time.Second * 1)
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
