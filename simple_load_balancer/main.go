package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	//"os"
)

func handleErr(err error){
	if err != nil {
		panic(err)
	}
}

type Server interface {

	Address() string

	IsAlive() bool

	Serve(rw http.ResponseWriter, req *http.Request)
}

type simpleServer struct{
	addr string
	proxy *httputil.ReverseProxy
}

func (s *simpleServer) Address() string { return s.addr}

func (s *simpleServer) IsAlive() bool { return true}

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

func newSimpleServer(addr string) *simpleServer{
	serverUrl,err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr: addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	 }
}

type loadBalancer struct{
	port string
	rrc int //roundRobinCount
	servers []Server
}

func (lb *loadBalancer) getNextAvailableServer() Server{

	server :=lb.servers[lb.rrc % len(lb.servers)]
	for !server.IsAlive(){
		lb.rrc++
		server = lb.servers[lb.rrc % len(lb.servers)]
	}
	lb.rrc++

	return server
}

func (lb *loadBalancer) serveTarget(rw http.ResponseWriter, req *http.Request){
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("forwarding request to address %q\n", targetServer.Address())
	targetServer.Serve(rw, req)
}

func newLoadBalancer(port string,servers []Server) *loadBalancer{
	return &loadBalancer{
		port: port,
		rrc: 0,
		servers: servers,
	}
}

func main(){
	servers := []Server{
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.bing.com"),
		//newSimpleServer("https://www.duckduckgo.com"),
		newSimpleServer("https://www.amazon.com"),
	}

	lb := newLoadBalancer("8000",servers)

	handleRedirect := func(rw http.ResponseWriter, req *http.Request){
		lb.serveTarget(rw,req)
	}

	http.HandleFunc("/", handleRedirect)

	fmt.Printf("serving requests at 'localhost:%s'\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)


}
