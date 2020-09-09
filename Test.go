package main

import (
	"fmt"
	jmtp "github.com/jmtp/jmtp-client-go"
	"github.com/jmtp/jmtp-client-go/jmtpclient"
)

func Hello() {
	fmt.Println("abc")
}

type jmtpClientAdapter struct {
	jmtpClient *jmtpclient.JmtpClient
}

func main() {
	//var c *jmtpclient.JmtpClient
	//a:=jmtpclient.JmtpClient{}
	var client *jmtpclient.JmtpClient
	client, _ = jmtpclient.NewJmtpClient(nil, func(packet jmtp.JmtpPacket, err error) {
	})
	fmt.Println("a")
	fmt.Println(client)
}
