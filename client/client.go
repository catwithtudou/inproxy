package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
	"strconv"
)

/**
 *@Author tudou
 *@Date 2020/8/30
 **/

var (
	remoteHost       string
	localPort  int
	remotePort int
)

func init() {
	flag.StringVar(&remoteHost, "h", "127.0.0.1", "remote server ip")
	flag.IntVar(&localPort, "l", 8080, "the local port")
	flag.IntVar(&remotePort, "r", 4200, "remote server port")
}


func main(){
	flag.Parse()
	defer func() {
		err:=recover()
		if err!=nil{
			panic(err)
		}
	}()

	for{
		//连接server端
		serverConn,err:=net.Dial("tcp",net.JoinHostPort(remoteHost,strconv.Itoa(remotePort)))
		if err!=nil{
			panic(err)
		}

		log.Printf("[server]%s connected\n",serverConn.RemoteAddr())
		server := &server{
			conn:   serverConn,
			read:   make(chan []byte),
			write:  make(chan []byte),
			exit:   make(chan error),
			reConn: make(chan bool),
		}

		go server.Read()
		go server.Write()

		go handle(server)

		//尝试重连
		<-server.reConn
		log.Println("[server]trying connect again")
	}
}


//处理server端与local端交互
func handle(server *server) {
	//得到server端写入的数据
	data := <-server.read

	//连接local端
	localConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		panic(err)
	}

	local := &local{
		conn:  localConn,
		read:  make(chan []byte),
		write: make(chan []byte),
		exit:  make(chan error),
	}

	go local.Read()
	go local.Write()

	//连接后直接将data写入local端
	local.write <- data

	for {
		select {
		case data := <-server.read:
			//写入local端
			local.write <- data

		case data := <-local.read:
			//写入server端
			server.write <- data

		case err:=<-server.exit:
			log.Println("Failed In Server;Err:"+err.Error())
			//同时关闭user端和client端
			_ = server.conn.Close()
			_ = local.conn.Close()
			//尝试重连
			server.reConn<-true
			//结束当前goroutine
			runtime.Goexit()
		case err:=<-local.exit:
			log.Println("Failed In Local;Err:"+err.Error())
			_ = local.conn.Close()
		}
	}
}
