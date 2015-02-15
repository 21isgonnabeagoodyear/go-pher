package main
import "gopher"
import "fmt"
import "io/ioutil"
import "flag"
import "strconv"
import "os"

func main(){
	var host = flag.String("host", "localhost", "host to connect to first")
	var port = flag.Int("port", 70, "port to connect to first")
	var selector = flag.String("selector", "", "first selector to load")
	flag.Parse()






	var first gopher.Direntry
	first.Host = *host
	first.Port = *port
	first.Selector= *selector
	first.Code = gopher.DIRECTORY
	var dirstack []gopher.Direntry
	dirstack = append(dirstack, first)

	for{
		if len(dirstack) <= 0{break}
		if dirstack[len(dirstack)-1].Code == gopher.INDEXSEARCH{
			fmt.Println("Enter a search string followed by END:")
			var searchstring string
			for{
				var b [1]byte
				n, err := os.Stdin.Read(b[:])//fmt is such a piece of shit
				if b[0] == byte('\n') || b[0] == byte('\r') || err != nil || n != 1{break}
				searchstring = searchstring + string(b[0])
			}
			dirstack[len(dirstack)-1] = dirstack[len(dirstack)-1].SearchToDir(searchstring)
		}
		if dirstack[len(dirstack)-1].Code == gopher.DIRECTORY{
			fmt.Println("=============================================================================")
			fmt.Println("listing contents of", dirstack[len(dirstack)-1].Host+":"+strconv.Itoa(dirstack[len(dirstack)-1].Port)+" "+dirstack[len(dirstack)-1].Selector)
			dircontents, err := dirstack[len(dirstack)-1].FetchDir()
			if err != nil{
				fmt.Println("ERROR:",err)
				dirstack = dirstack[:len(dirstack)-1]
			}else{
				fmt.Println("   0         Quit\n   1         Back")
				for i, thing := range dircontents{
					if thing.Code == gopher.DIRECTORY{fmt.Printf("\x1b[32m")}
					if thing.Code == gopher.FILE{fmt.Printf("\x1b[33m")}
					if thing.Code == gopher.INDEXSEARCH{fmt.Printf("\x1b[34m")}
					if thing.Code == gopher.ERROR{fmt.Printf("\x1b[31m")}
					fmt.Printf("%4d %12s %30s  |  %s\n", i+2, gopher.CodeName(thing.Code),thing.Selector,  thing.Description)
					fmt.Printf("\x1b[0m")
				}
				dirindex := 1
				fmt.Scanf("%d", &dirindex)
				dirindex -=2
				if dirindex <= -2{
					return
				}else if dirindex == -1{
					dirstack = dirstack[:len(dirstack)-1]
				}else if dirindex >= len(dircontents){fmt.Println("invalid index!")}else{
					dirstack = append(dirstack, dircontents[dirindex])
				}
			}
		}else if dirstack[len(dirstack)-1].Code == gopher.FILE{
			filereader, err := dirstack[len(dirstack)-1].FetchFile()
			if err != nil{fmt.Println(err);continue}
			buf, err := ioutil.ReadAll(filereader)
			fmt.Println(string(buf))
			dirindex := 1
			fmt.Scanf("%d", &dirindex)
			dirindex -=2
			if dirindex <= -2{
				return
			}//else if dirindex == -1{dirstack = dirstack[:len(dirstack)-1]}
			dirstack = dirstack[:len(dirstack)-1]

		}else{fmt.Println("unsupported code: ", string(dirstack[len(dirstack)-1].Code));dirstack = dirstack[:len(dirstack)-1]}
	}
}
