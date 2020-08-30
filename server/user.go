package main

import (
	"io"
	"net"
	"time"
)

/**
 *@Author tudou
 *@Date 2020/8/30
 **/


type user struct{
	conn net.Conn
	read chan []byte
	write chan []byte
	exit chan error
}


//读取user端
func (u *user)Read(){
	//设置读取数据过时时间
	_ = u.conn.SetReadDeadline(time.Now().Add(time.Second * 100))
	for{
		data:=make([]byte,10240)
		n,err:=u.conn.Read(data)
		if err!=nil && err !=io.EOF{
			u.exit<-err
		}
		u.read<-data[:n]
	}

}

//写入user端
func (u *user)Write(){
	for{
		select{
		case data:=<-u.write:
			_,err:=u.conn.Write(data)
			if err!=nil && err!=io.EOF{
				u.exit<-err
			}
		}
	}
}