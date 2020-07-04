package mail

import (
	"bytes"
	"fmt"
	"github.com/qiniu/iconv"
	"io"
	"io/ioutil"
	"mime"
	"regexp"
	"strings"
	"time"
)
const (
	envelopeDateTimeLayout = "Mon, 02 Jan 2006 15:04:05 -0700"
)

// Permutations of the layouts defined in RFC 5322, section 3.3.
var envelopeDateTimeLayouts = [...]string{
	envelopeDateTimeLayout, // popular, try it first
	"_2 Jan 2006 15:04:05 -0700",
	"_2 Jan 2006 15:04:05 MST",
	"_2 Jan 2006 15:04 -0700",
	"_2 Jan 2006 15:04 MST",
	"_2 Jan 06 15:04:05 -0700",
	"_2 Jan 06 15:04:05 MST",
	"_2 Jan 06 15:04 -0700",
	"_2 Jan 06 15:04 MST",
	"Mon, _2 Jan 2006 15:04:05 -0700",
	"Mon, _2 Jan 2006 15:04:05 MST",
	"Mon, _2 Jan 2006 15:04 -0700",
	"Mon, _2 Jan 2006 15:04 MST",
	"Mon, _2 Jan 06 15:04:05 -0700",
	"Mon, _2 Jan 06 15:04:05 MST",
	"Mon, _2 Jan 06 15:04 -0700",
	"Mon, _2 Jan 06 15:04 MST",
}

var commentRE = regexp.MustCompile(`[ \t]+\(.*\)$`)
var whitespaceRe = regexp.MustCompile("[ \n\r]+")

// DecodeRFC2047WordUtf8 解析标题为utf8
func DecodeRFC2047WordUtf8(text string) string {
	return DecodeRFC2047Word(text, "utf8")
}

// DecodeRFC2047Word 解析标题
// RFC 2047：https://www.ietf.org/rfc/rfc2047.txt
func DecodeRFC2047Word(text string, targetCharset string) string {
	if targetCharset == "" {
		targetCharset = "utf8"
	}
	dec := new(mime.WordDecoder)
	dec.CharsetReader = iconvFun(targetCharset)
	decodedText, err := dec.Decode(text)
	if err != nil {
		return text
	}
	return decodedText
}

// iconvFun 编码转换
func iconvFun(targetCharset string) func(charset string, input io.Reader) (io.Reader, error) {
	return func(charset string, input io.Reader) (io.Reader, error) {
		cd, err := iconv.Open(targetCharset, charset)
		if err != nil {
			return input, err
		}
		defer cd.Close()

		c, err := ioutil.ReadAll(input)
		if err != nil {
			return input, err
		}
		var outbuf [512]byte
		output, _, err := cd.Conv(c, outbuf[:])
		if err != nil {
			return input, err
		}
		return bytes.NewReader(output), nil
	}
}

// DecodeDateTime 解析邮件时间
func DecodeDateTime(maybeDate string) (time.Time, error) {
	maybeDate = commentRE.ReplaceAllString(maybeDate, "")
	for _, layout := range envelopeDateTimeLayouts {
		parsed, err := time.Parse(layout, maybeDate)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("date %s could not be parsed", maybeDate)
}

// DecodeSubject 解析主题
// 主题有拆分的情况，中间用空格或者换行符分割，比如
// =?utf-8?B?UmU6IOWJjeerr+eUs+ivt+a0u+WKqOi3r+eUseacjeWKoS3nvo4=?= =?utf-8?B?5Lq65rS75Yqo?=
func DecodeSubject(subject string) (string) {
	split := whitespaceRe.Split(subject, -1)
	var strList []string
	for _, item := range split {
		strList = append(strList, DecodeRFC2047WordUtf8(item))
	}
	return strings.Join(strList, "")
}

