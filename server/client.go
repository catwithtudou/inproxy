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

type client struct{
	conn net.Conn
	//读写通道
	read chan []byte
	write chan []byte
	//退出通道
	exit chan error
	//重连通道
	reConn chan bool
}

//Client端读取
func (c *client)Read(){
	//设置Client端连接过期时间
	_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	for{
		data := make([]byte,10240)
		n,err:=c.conn.Read(data)
		if err!=nil && err!=io.EOF{
			if strings.Contains(err.Error(),"timeout"){
				//设置Client端心跳包过期时间
				_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * 5))
				_,_=c.conn.Write([]byte("ping"))
				continue
			}
			log.Println("Failed In Read")
			c.exit <- err
		}

		//收到心跳包则跳过
		if string(data[:4]) == "ping"{
			log.Println("server accept ping")
			continue
		}

		c.read <- data[:n]
	}

}

//写入Client端
func (c *client)Write(){
	for{
		select {
		case data:=<-c.write:
			_,err:=c.conn.Write(data)
			if err!=nil && err!= io.EOF{
				c.exit<-err
			}
		}
	}
}