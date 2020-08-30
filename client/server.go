package main

import (
	"io"
	"log"
	"net"
	"strings"
	"time"
)

/**
 *@Author tudou
 *@Date 2020/8/30
 **/

type server struct {
	conn net.Conn
	//读写通道
	read chan []byte
	write chan []byte
	//退出通道
	exit chan error
	//重连通道
	reConn chan bool
}

//读取server端
func (s *server)Read(){
	//设置Server端连接超时时间
	_ = s.conn.SetReadDeadline(time.Now().Add(time.Second*10))
	for{
		//与服务端处理相似
		data:=make([]byte,10240)
		n,err:=s.conn.Read(data)
		if err != nil && err != io.EOF {
			// 超时
			if strings.Contains(err.Error(), "timeout") {
				//设置心跳包超时时间
				_ = s.conn.SetReadDeadline(time.Now().Add(time.Second * 3))
				_,_=s.conn.Write([]byte("ping"))
				continue
			}
			log.Println("Failed In Read")
			s.exit <- err
		}

		//收到心跳包则跳过
		if string(data[:4]) == "ping"{
			log.Println("server accept ping")
			continue
		}

		s.read <- data[:n]
	}
}

//写入server端
func (s *server)Write(){
	for{
		select {
		case data:=<-s.write:
			_,err:=s.conn.Write(data)
			if err!=nil && err!= io.EOF{
				s.exit<-err
			}
		}
	}
}