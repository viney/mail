package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"path/filepath"
)

const (
	Addr     = "smtp.exmail.qq.com"
	Host     = Addr + ":25"
	AuthName = "viney@qq.com"
	AuthPwd  = "viney"
)

type Message struct {
	SenderName string
	// 发送者
	Sender string
	// 结算用户列表
	To     []string
	ToName []string
	// 主题
	Subject string
	// 内容
	Body   string
	Marker string
}

func NewMessage() *Message {
	return &Message{
		SenderName: "viney",
		Sender:     "viney@qq.com",
		To:         []string{"test@qq.com"},
		// To:         []string{"viney.chow@gmail.com", "kf.ye@ot24.net"},
		ToName:  []string{"Viney"},
		Subject: "Test",
		Body:    "Test",
		Marker:  "ACUSTOMANDUNIQUEBOUNDARY",
	}
}

// mail headers
func (m *Message) Head() string {
	return fmt.Sprintf("From: %s <%s>\r\nTo: %s <%s>\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n--%s",
		m.SenderName, m.Sender, m.ToName[0], m.To[0], m.Subject, m.Marker, m.Marker)
}

// body (text or HTML)
func (m *Message) Bodys() string {
	return fmt.Sprintf("\r\nContent-Type: text/html\r\nContent-Transfer-Encoding:8bit\r\n\r\n%s\r\n--%s", m.Body, m.Marker)
}

var ContentType = map[string]string{
	".gif":  "image/gif",
	".doc":  "application/msword",
	".docx": "application/msword",
	"jpg":   "image/jpeg",
}

func (m *Message) Encode(filename string) (string, error) {
	var buf bytes.Buffer

	name := filepath.Base(filename)
	contentType := ContentType[filepath.Ext(filename)]

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(content)

	lineMaxLength := 500
	nbrLines := len(encoded) / lineMaxLength

	//append lines to buffer
	for i := 0; i < nbrLines; i++ {
		buf.WriteString(encoded[i*lineMaxLength:(i+1)*lineMaxLength] + "\n")
	}

	//append last line in buffer
	buf.WriteString(encoded[nbrLines*lineMaxLength:])

	return fmt.Sprintf("\r\nContent-Type: %s; name=\"%s\"\r\nContent-Transfer-Encoding:base64\r\nContent-Disposition: attachment; filename=\"%s\"\r\n\r\n%s\r\n--%s--",
		contentType, name, name, buf.String(), m.Marker), nil
}

func main() {
	filename := "golang.jpg"

	m := NewMessage()
	encodeStr, err := m.Encode(filename)
	if err != nil {
		fmt.Println("Encode: ", err)
		return
	}

	msg := m.Head() + m.Bodys() + encodeStr

	auth := smtp.PlainAuth(m.Sender, AuthName, AuthPwd, Addr)

	// go
	count := 100
	finish := make(chan bool)

	for i := 0; i < count; i++ {
		go func() {
			defer func() { finish <- true }()
			//send the email
			if err := smtp.SendMail(Host, auth, m.Sender, m.To, []byte(msg)); err != nil {
				fmt.Println("SendMail: ", err)
				return
			}
		}()
	}

	for i := 0; i < count; i++ {
		<-finish
	}
}
