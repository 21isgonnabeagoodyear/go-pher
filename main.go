package main
import "gopher"
import "flag"
import "strconv"
import "fmt"
import "path/filepath"

var myrqh gopher.RequestHandler = func(r *gopher.Request) error{
	//_, err := r.Write([]byte("nothing to see here"))
	fmt.Println("request for:", r.Selector)
	if r.Selector == "/helloworld"{
		_, err := r.Write([]byte("jello world"))
		return err
	}else{
		err := r.WriteInfo("This server is not complete")
		if err != nil{return err}
		errentry := gopher.Direntry{Code:gopher.FILE, Selector:"/helloworld", Description:"example file", Host:r.MyDomain, Port:r.MyPort}//"localhost", Port:7070}
		err  = r.WriteEntry(errentry)
		return err
	}
}
var generalrqh gopher.RequestHandler = func(r *gopher.Request) error{
	return r.WriteInfo("this is the default handler")
}
var generalrqh2 gopher.RequestHandler = func(r *gopher.Request) error{
	return r.WriteInfo("this is the last handler")
}
var otherrqh gopher.RequestHandler = func(r *gopher.Request) error{
	return r.WriteInfo("this is the other one")
}
var fileserverheader gopher.RequestHandler = func(r *gopher.Request) error{
	r.WriteInfo(       "Welcome to this go-pher server")
	return r.WriteInfo("==============================")
}

func main(){
	var host = flag.String("host", "localhost", "internet accessible domain name to serve local documents on")
	var port = flag.Int("port", 70, "port to listen on")
	flag.Parse()

	var m gopher.Mux
/*
	m.Add("/hella", otherrqh)
	m.Add("/hello", myrqh)
	m.Add("", generalrqh)
	m.Add("", generalrqh2)
*/
	thisdir, err := filepath.Abs(".")
	if err != nil{fmt.Println(err);return}
	m.Add("", gopher.NewFileServerWithHeader(thisdir, fileserverheader))



	fmt.Println(gopher.ListenAndServe(*host+":"+strconv.Itoa(*port), m.GetRequestHandler()))
}
