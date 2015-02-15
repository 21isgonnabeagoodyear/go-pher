package gopher
import "net"
import "strconv"
import "bytes"
import "errors"
import "fmt"
import "bufio"
import "io"
import "io/ioutil"

const(
/*
from rfc
   0   Item is a file
   1   Item is a directory
   2   Item is a CSO phone-book server
   3   Error
   4   Item is a BinHexed Macintosh file.
   5   Item is DOS binary archive of some sort.
       Client must read until the TCP connection closes.  Beware.
   6   Item is a UNIX uuencoded file.
   7   Item is an Index-Search server.
   8   Item points to a text-based telnet session.
   9   Item is a binary file!
       Client must read until the TCP connection closes.  Beware.
   +   Item is a redundant server
   T   Item points to a text-based tn3270 session.
   g   Item is a GIF format graphics file.
   I   Item is some kind of image file.  Client decides how to display.

*/
	FILE = byte(iota+'0')// =text document with no funny characters?
	DIRECTORY
	PHONEBOOK
	ERROR
	BINHEX
	DOSARCHIVE
	UUENCODED
	INDEXSEARCH
	TELNET
	BINARY
	REDUNDANT = byte('+')
	TN3270 = byte('T')
	GIF = byte('g')
	IMAGE = byte('I')

	END = byte('.')
//nonstandard
	INFO = byte('i')
	HTML = byte('h')
	AUDIO = byte('s')
)
const CRLF = "\r\n"
//CodeName returns a string representation of a gopher type code (e.g. FILE)
func CodeName(b byte) string{
	switch{
		case b == FILE:
			return "FILE"
		case b == DIRECTORY:
			return "DIRECTORY"
		case b == PHONEBOOK:
			return "PHONEBOOK"
		case b == ERROR:
			return "ERROR"
		case b == BINHEX:
			return "BINHEX"
		case b == DOSARCHIVE:
			return "DOSARCHIVE"
		case b == UUENCODED:
			return "UUENCODED"
		case b == INDEXSEARCH:
			return "INDEXSEARCH"
		case b == TELNET:
			return "TELNET"
		case b == BINARY:
			return "BINARY"
		case b == REDUNDANT:
			return "REDUNDANT"
		case b == TN3270:
			return "TN3270"
		case b == GIF:
			return "GIF"
		case b == IMAGE:
			return "IMAGE"
		case b == END:
			return "END"
		case b == INFO:
			return "INFO"
		case b == HTML:
			return "HTML"
		case b == AUDIO:
			return "AUDIO"
		default:
			return "UNKNOWN"
	}

}
//A Direntry describes an entry in a directory listing.  
type Direntry struct{
	Code byte
	Description string
	Selector string
	Host string
	Port int
	Extras []string//nonstandard extensions (ignored by standard clients)

}
func (d Direntry)serialize() ([]byte, error){
	line := []byte{}
	line = append(line, d.Code)
	line = append(line, []byte(d.Description)...)
	line = append(line,   byte('\t'))
	line = append(line, []byte(d.Selector)...)
	line = append(line,   byte('\t'))
	line = append(line, []byte(d.Host)...)
	line = append(line,   byte('\t'))
	line = append(line, []byte(strconv.Itoa(d.Port))...)
	for _, s := range d.Extras{
		line = append(line,   byte('\t'))
		line = append(line, []byte(s)...)
	}
	line = append(line, []byte(CRLF)...)
	return line, nil
}
func (d *Direntry)unserialize(text []byte) error{
	parts := bytes.Split(bytes.Trim(text, "\r\n"/*FIXME:this in order*/), []byte{'\t'})
	if len(parts) < 4{return errors.New("truncated directory line :"+string(text))}
	if len(parts[0]) < 1{return errors.New("no code part")}
	d.Code = parts[0][0]
	d.Description = string(parts[0][1:])
	d.Selector = string(parts[1])
	d.Host = string(parts[2])
	port, err := strconv.Atoi(string(parts[3]))
	if err != nil{return err}
	d.Port = port
	if len(parts) >= 4{
		for _, val := range parts[4:]{d.Extras = append(d.Extras, string(val))}
	}
	return nil
}
//FetchFile fetches data, not directory information.  Calling this on a directory or unsupported type will return an error.
//TODO:decode special encodings
func (d *Direntry)FetchFile() (io.Reader, error){
	if d.Code == DIRECTORY{return nil, errors.New("cannot fetch a directory as a file (use FetchDir)")}
	if d.Code != FILE{return nil, errors.New("non-plaintext encodings not supported")}
	//return nil, errors.New("FetchFile not implemented")
	conn, err := net.Dial("tcp", d.Host+":"+strconv.Itoa(d.Port))
	if err != nil{return nil, err}
	_, err  = conn.Write([]byte(d.Selector+CRLF))
	if err != nil{conn.Close();return nil, err}
	return conn, nil
}
//FetchDir fetches directory information, not data.  Calling this on a Direntry whose code is not DIRECTORY will return an error.
func (d *Direntry)FetchDir() ([]Direntry, error){
	if d.Code != DIRECTORY{return nil, errors.New("cannot fetch a file as a directory (use FetchFile)")}
	conn, err := net.Dial("tcp", d.Host+":"+strconv.Itoa(d.Port))
	if err != nil{return nil, err}
	_, err  = conn.Write([]byte(d.Selector+CRLF))
	if err != nil{return nil, err}
	data, err := ioutil.ReadAll(conn)
	if err != nil{return nil, err}//TODO:figure out if this triggers on server closing connection
	parts := bytes.Split(data, []byte("\n"))//FIXME:should actually split on \r\n
	var rv []Direntry
	for _, line := range parts{
		if len(line) <=3{continue}//ignore short lines (xkcd seems to have these)
		if string(line) == ".\r"{break}//???
		var d Direntry
		err := d.unserialize(line)
		if err != nil{return rv, err}else{rv = append(rv, d)}
	}
	return rv, nil
}
//SearchToDir creates a directory entry from a search (INDEXSEARCH) entry, given a search string.  
//INDEXSEARCH entries are selectors which may be partly specified by the client, e.g. search engines.
func (d *Direntry)SearchToDir(searchstring string) Direntry{
	var rv = *d
	rv.Selector += "\t"
	rv.Selector += searchstring
	rv.Code = DIRECTORY
	return rv
}

//A Request is given to the RequestHandler callback to describe a request from a client.
//MyDomain and MyPort may be used by the RequestHandler for local links.  These are specified in the call to ListenAndServe.
type Request struct{
	connection net.Conn
	isdir byte//'y':is directory, 'n':is file
	Selector string
	MyDomain string
	MyPort int
}
//WriteEntry sends a Direntry to the client.  Calling WriteEntry after calling Write on the same Request is an error.
func (r *Request)WriteEntry(d Direntry) error{
	if r.isdir == byte('n'){return errors.New("cannot write directory data to a document!")}
	r.isdir = byte('y')
	text, err := d.serialize()
	if err != nil{return err}
	_, err  = r.connection.Write(text)
	if err != nil{return err}
	return nil
}
//WriteError sends a Direntry representing an error message to the client.  Calling WriteError after calling Write on the same Request is an error.
func (r *Request)WriteError(err string) error{
	if r.isdir == byte('n'){
		_, e := r.Write([]byte(err))
		return e
	}
	infoentry := Direntry{Code:ERROR, Selector:"fake", Description:err}
	return r.WriteEntry(infoentry)
}
//WriteError sends a Direntry representing an informational message to the client.  Calling WriteInfo after calling Write on the same Request is an error.  This is a nonstandard code and may be ignored by some clients.
func (r *Request)WriteInfo(info string) error{
	infoentry := Direntry{Code:INFO, Selector:"fake", Description:info}
	return r.WriteEntry(infoentry)
}
//Write sends (usually ascii encoded plaintext) data to the client.  Calling Write after calling WriteInfo, WriteError or WriteEntry on the same Request is an error.  
func (r *Request)Write(tw []byte) (int, error){
	if r.isdir == byte('y'){return 0, errors.New("cannot write document data to a directory!")}
	r.isdir = byte('n')
	return r.connection.Write(tw)
}
//sends the closing period and newline
func (r *Request)endentries() (error){
	_, err := r.connection.Write([]byte(".\r\n"))
	return err
}
//A RequestHandler is a callback used by a server to fullfill requests.  The RequestHandler should call Write* functions on the request and return nil on success.  For security reasons, the client will receive a generic error message if this returns any error.
//RequestHandlers are called inside a separate goroutine so locks or channels must be used whenever accessing shared data.
type RequestHandler func(r *Request) (error)

//the default request handler if none is specified
var defaultrequesthandler RequestHandler = func(r *Request) error{
	//_, err := r.Write([]byte("nothing to see here"))
	fmt.Println("request for:", r.Selector)
	if r.Selector == "/helloworld"{
		_, err := r.Write([]byte("hello world"))
		return err
	}else{
		err := r.WriteInfo("This server is functioning but is only using the default request handler")
		if err != nil{return err}
		errentry := Direntry{Code:FILE, Selector:"/helloworld", Description:"example file", Host:r.MyDomain, Port:r.MyPort}//"localhost", Port:7070}
		err  = r.WriteEntry(errentry)
		return err
	}
}

//ListenAndServe starts serving gopher requests using the given RequestHandler.
//The address passed to ListenAndServe should be an internet-accessable domain name, optionally followed by a colon and the port number.
//If the address is not a FQDN, MyDomain as passed to the RequestHandler may not be accessible to clients, so links may not work.
func ListenAndServe(addr string, r RequestHandler) error{
	if addr == ""{addr = ":70"}
	if r == nil{r = defaultrequesthandler}
	listener, err := net.Listen("tcp", addr)
	if err != nil{return err}
	for{
		conn, err := listener.Accept()
		if err != nil{return err}
		go func(){
			defer conn.Close()
			req := Request{connection:conn}
			if tcpa, ok := listener.Addr().(*net.TCPAddr); ok{
				req.MyDomain = tcpa.IP.String()//FIXME:this won't work so good on the interwebs (use addr+net.SplitHostPort  to get domain name?)
				req.MyPort = tcpa.Port
			}
			//read selector
			bufreader := bufio.NewReader(conn)
			sel, err := bufreader.ReadBytes(byte('\n'/*FIXME:should be \r\n sequence, could be \ns in there alone*/))
			if len(sel) < 2 || sel[len(sel)-2] != '\r'{fmt.Errorf("malformed or unusual query, newline without carriage return\n");return}
			req.Selector = string(sel[:len(sel)-2])
			if err != nil{fmt.Printf("failed to retrieve selector line:%s\n",err);return}
			err = r(&req)
			if err != nil{fmt.Printf("request handler encountered an error:%s\n",err); req.WriteError("internal server error")}
			if req.isdir == byte('y'){req.endentries()}//else{req.Write([]byte("\n.\r\n"))}
		}()
	}
}
