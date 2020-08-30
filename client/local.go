package main

import "net"

/**
 *@Author tudou
 *@Date 2020/8/30
 **/

type local struct {
	conn net.Conn
	// 数据传输通道
	read  chan []byte
	write chan []byte
	// 有异常退出通道
	exit chan error
}

//读取local端
func (l *local)Read(){
	for {
		data := make([]byte, 10240)
		n, err := l.conn.Read(data)
		if err != nil {
			l.exit <- err
		}
		l.read <- data[:n]
	}
}

//写入local端
func (l *local) Write() {
	for {
		select {
		case data := <-l.write:
			_, err := l.conn.Write(data)
			if err != nil {
				l.exit <- err
			}
		}
	}
}