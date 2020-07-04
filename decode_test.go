package mail

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeRFC2047Word(t *testing.T) {
	testCases := []struct {
		give string
		want string
	}{
		{
			give: "=?utf-8?B?5rWL6K+V5paH5pys?=",
			want: "测试文本",
		},
		{
			give: "=?utf-8?B?5rWL6K+V5paH5pys7aC87b+g?=",
			want: "测试文本\xed\xa0\xbc\xed\xbf\xa0",
		},
		{
			give: "=?UTF-8?q?=E6=B5=8B=E8=AF=95=E6=96=87=E6=9C=AC?=",
			want: "测试文本",
		},
		{
			give: "=?gb2312?B?xOPU2r2ty9XKodDeuMTBy9PKz+TD3MLr?=",
			want: "你在江苏省修改了邮箱密码",
		},
	}
	for _, testCase := range testCases {
		got := DecodeRFC2047Word(testCase.give, "utf8")
		assert.Equal(t, testCase.want, got)
	}
}

func TestDecodeDateTime(t *testing.T) {
	testCases := []struct {
		give string
		want string
	}{
		{
			give: "Wed, 1 Jul 2020 09:51:31 +0800 (GMT+08:00)",
		},
	}
	for _, testCase := range testCases {
		got, err := DecodeDateTime(testCase.give)
		assert.Nil(t, err)
		assert.True(t, !got.IsZero())
		//fmt.Println(got.Format(time.RFC3339))
	}
}

func TestDecodeSubject(t *testing.T) {
	testCases := []struct {
		give string
		want string
	}{
		{
			give: "=?utf-8?B?UmU6IOWJjeerr+eUs+ivt+a0u+WKqOi3r+eUseacjeWKoS3nvo4=?= =?utf-8?B?5Lq65rS75Yqo?=",
			want: "Re: 前端申请活动路由服务-美人活动",
		},
		{
			give: `=?utf-8?B?UmU6IOWJjeerr+eUs+ivt+a0u+WKqOi3r+eUseacjeWKoS3nvo4=?=
 =?utf-8?B?5Lq65rS75Yqo?=`,
			want: "Re: 前端申请活动路由服务-美人活动",
		},
		{
			give: `测试测试`,
			want: "测试测试",
		},
	}
	for _, testCase := range testCases {
		got := DecodeSubject(testCase.give)
		assert.Equal(t, got, testCase.want)
	}
}
