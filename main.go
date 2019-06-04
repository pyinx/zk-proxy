package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/pyinx/zk-proxy/zk"
	"golang.org/x/net/context"
)

var (
	buildstamp string
	githash    string
	goversion  string
)

var (
	backendAddrs = flag.String("backend_addr", "", "zk server address: 1.1.1.1:2181,2.2.2.2:2181")
	httpAddr     = flag.String("http_addr", "0.0.0.0:8000", "http address")
	proxyAddr    = flag.String("proxy_addr", "0.0.0.0:2182", "proxy address")
	cpuNum       = flag.Int("cpu_num", 1, "max cpu num")
	ipAcl        = flag.Bool("ip_acl", false, "enable ip acl for zk path")
	limitNum     = flag.Int("limit_num", -1, "limit num for request rate")
	version      = flag.Bool("version", false, "show proxy version")
)

func main() {
	flag.Parse()
	if *version {
		showVersion()
		return
	}

	if len(*backendAddrs) == 0 {
		fmt.Println("help to get usage")
		return
	}

	setCpuNum(*cpuNum)

	ln, err := net.Listen("tcp", *proxyAddr)
	if err != nil {
		panic(err)
	}
	ctx, cancle := context.WithCancel(context.Background())

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go stopProc(cancle, c)

	go zk.StartHttp(*httpAddr)
	if *ipAcl {
		zk.InitAcl(zk.GetZkServers(*backendAddrs))
	}
	if *limitNum > 0 {
		zk.SetLimit(*limitNum)
	}
	// go cpuProfile()
	// go heapProfile()

	serv := zk.Serve
	serv(ctx, ln, zk.NewAuth(zk.GetZkServers(*backendAddrs)), zk.NewZK())
}

func setCpuNum(cpuNum int) {
	if cpuNum > 4 {
		runtime.GOMAXPROCS(4)
	} else if cpuNum < 1 {
		runtime.GOMAXPROCS(1)
	} else {
		runtime.GOMAXPROCS(cpuNum)
	}
}

func stopProc(cancle context.CancelFunc, c chan os.Signal) {
	s := <-c
	fmt.Printf("receive signal %s, exiting...\n", s)
	cancle()
	time.Sleep(10 * time.Millisecond)
	os.Exit(0)
}

func showVersion() {
	fmt.Println("buildstamp: " + buildstamp)
	fmt.Println("githash: " + githash)
	fmt.Println("goversion: " + goversion)
}

func cpuProfile() {
	f, _ := os.OpenFile("cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
	defer f.Close()

	fmt.Println("CPU Profile started")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	time.Sleep(120 * time.Second)
	fmt.Println("CPU Profile stopped")
}

func heapProfile() {
	f, _ := os.OpenFile("heap.prof", os.O_RDWR|os.O_CREATE, 0644)
	defer f.Close()

	fmt.Println("Heap Profile started")
	time.Sleep(120 * time.Second)

	pprof.WriteHeapProfile(f)
	fmt.Println("Heap Profile generated")
}
