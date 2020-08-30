package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
)

/**
 *@Author tudou
 *@Date 2020/8/30
 **/

var (
	publicPort int
	remotePort int
)

func init() {
	flag.IntVar(&publicPort, "l", 4100, "the public port")
	flag.IntVar(&remotePort, "r", 4200, "the client port")
}

func main() {
	flag.Parse()
	defer func() {
		err:=recover()
		if err!=nil{
			panic(err)
		}
	}()

	//监听client端
	clientListener,err:=net.Listen("tcp",fmt.Sprintf(":%d",remotePort))
	if err!=nil{
		panic(err)
	}
	log.Printf("[client]listen the port:%d, waiting for client connecting",remotePort)

	//监听user端
	userListener,err:=net.Listen("tcp",fmt.Sprintf(":%d",publicPort))
	if err!=nil{
		panic(err)
	}
	log.Printf("[user]listen the port:%d, waiting for user connecting",publicPort)

	for {
		//接收client端连接
		clientConn, err := clientListener.Accept()
		if err != nil {
			panic(err)
		}

		log.Printf("[client]%s connected\n", clientConn.RemoteAddr())

		client := &client{
			conn:   clientConn,
			read:   make(chan []byte),
			write:  make(chan []byte),
			exit:   make(chan error),
			reConn: make(chan bool),
		}

		//协程接收user端连接
		userConnChan := make(chan net.Conn)
		go acceptUserConn(userListener, userConnChan)

		//client端与user端交互
		go handleClient(client, userConnChan)

		//重新连接阻塞
		<-client.reConn
		log.Println("[client]trying connect again")
	}


}


//连接user端
func acceptUserConn(userListener net.Listener,connChan chan net.Conn){
	userConn,err:=userListener.Accept()
	if err!=nil{
		panic(err)
	}
	log.Printf("[user]%s connected\n",userConn.RemoteAddr())
	connChan <- userConn
	return
}

//处理client端与user端交互
func handleClient(client *client, userConnChan chan net.Conn) {
	go client.Read()
	go client.Write()

	for {
		select {
		case err := <-client.exit:
			log.Println("Failed In Client;Err:" + err.Error())
			//若client端出现错误则尝试重连
			client.reConn <- true
			//结束当前goroutine
			runtime.Goexit()
		case userConn := <-userConnChan:
				user:=&user{
					conn:  userConn,
					read:  make(chan []byte),
					write: make(chan []byte),
					exit:  make(chan error),
				}
				go user.Read()
				go user.Write()

				go handleUser(client,user)
		}
	}
}


func handleUser(client *client,user *user){
	for{
		select{
		case userResp:=<-user.read:
			//写入client端
			client.write<-userResp
		case clientResp:=<-client.write:
			//写入user端
			user.write<-clientResp
		case err:=<-client.exit:
			log.Println("Failed In Client;Err:"+err.Error())
			//同时关闭user端和client端
			_ = client.conn.Close()
			_ = user.conn.Close()
			//尝试重连
			client.reConn<-true
			//结束当前goroutine
			runtime.Goexit()
		case err:=<-user.exit:
			log.Println("Failed In User;Err:"+err.Error())
			_ = user.conn.Close()
		}
	}
}


