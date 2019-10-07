package socks

import "time"

var readTimeout time.Duration

// Disguised as normal HTTP traffic to bypass gfw
var preface = []byte("GET /up/2016-12/20161227101716.jpg HTTP/1.1\nHost: www.tonychan.me\nConnection: keep-alive\nUser-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36\nAccept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3\nAccept-Encoding: gzip, deflate\nAccept-Language: en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,zh-TW;q=0.6,pl;q=0.5\n\n")
var prefaceLen = len(preface)
