package main

import (
	"container/list"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

var (
	port         = flag.Uint("port", 0, "监听端口")
	onlineServer = flag.String("online", "", "线上服务地址")
	testServer   = flag.String("test", "", "测试服务地址")
	bufferSize   = flag.Uint("buffer", 65536, "缓存大小")
)

func exit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args)
	os.Exit(1)
}

func main() {
	flag.Parse()
	if *port == 0 || *onlineServer == "" || *testServer == "" || *bufferSize == 0 {
		exit("参数错误")
	}
	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", *port))
	onlineAddr, err := net.ResolveTCPAddr("tcp", *onlineServer)
	if err != nil {
		exit("解析线上服务地址出错: addr=%q, error=%q", onlineAddr, err.Error())
	}
	testAddr, err := net.ResolveTCPAddr("tcp", *testServer)
	if err != nil {
		exit("解析测试服务地址出错: addr=%q, error=%q", testAddr, err.Error())
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		exit("打开监听出错: port=%d, error=%q", *port, err.Error())
	}

	fmt.Println("启动服务")
	nextConnID := 0
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Fprintf(os.Stderr, "接受连接出错: port=%d, error=%q\n", *port, err.Error())
			continue
		}
		connID := nextConnID
		nextConnID++
		fmt.Fprintf(os.Stderr, "建立连接成功: id=%d, port=%d\n", connID, *port)
		go func() {
			// 所有同步锁
			locks := list.New()
			defer conn.Close()
			// 请求通道，最多两个，至少1个
			inStreams := make([]chan []byte, 0, 2)
			// 建立线上服务连接
			if onlineConn, err := net.DialTCP("tcp", nil, onlineAddr); err != nil {
				fmt.Fprintf(os.Stderr, "建立线上服务连接出错: id=%d, addr=%q, error=%q\n",
					connID, *onlineServer, err.Error())
				return
			} else {
				fmt.Fprintf(os.Stderr, "建立线上服务连接成功: id=%d, addr=%q\n",
					connID, *onlineServer)
				defer func() {
					if err := onlineConn.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "关闭线上服务连接出错: id=%d, addr=%q, error=%q\n",
							connID, *onlineServer, err.Error())
					}
				}()
				// 转发请求到线上服务器
				onlineInStream := make(chan []byte)
				inStreams = append(inStreams, onlineInStream)
				// 读协程
				go func() {
					lock := new(sync.Mutex)
					lock.Lock()
					defer lock.Unlock()
					locks.PushBack(lock)
					totalBytes := 0
					for content := range onlineInStream {
						n, err := onlineConn.Write(content)
						totalBytes += n
						if err != nil {
							fmt.Fprintf(os.Stderr, "转发线上请求出错: id=%d, bytes=%d, totalBytes=%d, error=%q\n",
								connID, n, totalBytes, err.Error())
							return
						} else {
							fmt.Fprintf(os.Stderr, "转发线上请求: id=%d, bytes=%d, totalBytes=%d\n",
								connID, n, totalBytes)
						}
					}
					fmt.Fprintf(os.Stderr, "转发线上请求完成: id=%d, totalBytes=%d\n",
						connID, totalBytes)
				}()
				// 写协程
				go func() {
					lock := new(sync.Mutex)
					lock.Lock()
					defer lock.Unlock()
					locks.PushBack(lock)
					totalBytes := 0
					buffer := make([]byte, *bufferSize)
					for {
						n, err := onlineConn.Read(buffer)
						if n > 0 {
							m, err := conn.Write(buffer[:n])
							totalBytes += m
							if err != nil {
								fmt.Fprintf(os.Stderr, "转发线上应答出错: id=%d, addr=%q, bytes=%d, totalBytes=%d, error=%q\n",
									connID, *onlineServer, m, totalBytes, err.Error())
								return
							} else {
								fmt.Fprintf(os.Stderr, "转发线上应答: id=%d, addr=%q, bytes=%d, totalBytes=%d\n",
									connID, *onlineServer, m, totalBytes)
							}
						}
						if err == io.EOF {
							fmt.Fprintf(os.Stderr, "读取线上服务器应答完成: id=%d, addr=%q, totalBytes=%d\n",
								connID, *onlineServer, totalBytes)
							return
						} else if err != nil {
							fmt.Fprintf(os.Stderr, "读取线上服务器应答出错: id=%d, addr=%q, totalBytes=%d, error=%q\n",
								connID, *onlineServer, totalBytes, err.Error())
							return
						}
					}
				}()
			}

			// 建立测试服务连接
			if testConn, err := net.DialTCP("tcp", nil, testAddr); err != nil {
				fmt.Fprintf(os.Stderr, "建立测试服务连接出错: id=%d, addr=%q, error=%q\n",
					connID, *testServer, err.Error())
			} else {
				defer func() {
					if err := testConn.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "关闭测试服务连接出错: id=%d, addr=%q, error=%q\n",
							connID, *testServer, err.Error())
					}
				}()
				// 转发请求到测试服务器
				testInStream := make(chan []byte)
				inStreams = append(inStreams, testInStream)
				go func() {
					lock := new(sync.Mutex)
					lock.Lock()
					defer lock.Unlock()
					locks.PushBack(lock)
					totalBytes := 0
					for content := range testInStream {
						testConn.Write(content)
						totalBytes += len(content)
						fmt.Fprintf(os.Stderr, "转发测试请求: id=%d, bytes=%d, totalBytes=%d\n",
							connID, len(content), totalBytes)
					}
					fmt.Fprintf(os.Stderr, "转发测试请求完成: id=%d, totalBytes=%d\n",
						connID, totalBytes)
				}()
				// 抛弃测试服务应答
				go func() {
					lock := new(sync.Mutex)
					lock.Lock()
					defer lock.Unlock()
					locks.PushBack(lock)
					if n, err := io.Copy(nullWriter, testConn); err != nil {
						fmt.Fprintf(os.Stderr, "抛弃测试服务器应答出错: id=%d, addr=%q, size=%d, error=%q\n",
							connID, *testServer, n, err.Error())
					}
				}()
			}
			buffer := make([]byte, *bufferSize)
			totalBytes := 0
			for {
				n, err := conn.Read(buffer)
				if n > 0 {
					for _, inStream := range inStreams {
						inStream <- buffer[:n]
					}
					totalBytes += n
					fmt.Fprintf(os.Stderr, "读取请求: id=%d, bytes=%d, totalBytes=%d\n",
						connID, n, totalBytes)
				}
				if err != nil {
					if err != io.EOF {
						fmt.Fprintf(os.Stderr, "读取请求出错: id=%d, totalBytes=%d, error=%q\n",
							connID, totalBytes, err.Error())
					} else {
						fmt.Fprintf(os.Stderr, "请求读取完成: id=%d, totalBytes=%d\n",
							connID, totalBytes)
					}
					break
				}
			}
			for lock := locks.Front(); lock != nil; lock = lock.Next() {
				lock.Value.(*sync.Mutex).Lock()
			}
		}()
	}
}

type NullWriter struct{}

func (writer NullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

var (
	nullWriter = new(NullWriter)
)
