package gopher
import "errors"
//import "fmt"

type callbackholder struct{
	prefix string
	callback RequestHandler
}
//A Mux allows you to use multiple RequestHandlers in the same server.  
//For each request the longest matching prefix will be used to select which RequestHandler to call.
type Mux struct{
	callbacks []callbackholder
}
//Add adds a RequestHandler to the Mux and associates it with a prefix.
//When multiple RequestHandlers have the same prefix, the last specified will be used.
func (m *Mux)Add(prefix string, callback RequestHandler){
	m.callbacks = append(m.callbacks, callbackholder{prefix, callback})
}
//GetRequestHandler returns a closured callback which uses the Mux's prefixes and RequestHandlers.  
//Add should not be called after calling this function.
func (m *Mux)GetRequestHandler() RequestHandler{
	return func(r *Request) error{
		var besthandler RequestHandler = nil
		besthandlerlen := -1
		for _, cbh := range m.callbacks{
			if len(r.Selector) >= len(cbh.prefix){
				if len(r.Selector) == 0{besthandlerlen=0;besthandler=cbh.callback;continue}//the last default handler (prefix "") always goes with the empty selector
				if len(cbh.prefix) < besthandlerlen{continue}//don't use shorter prefixes
				//at this point r.Selector is at least one character long and at least as long as cbh.prefix
				if r.Selector[:len(cbh.prefix)] == cbh.prefix{besthandler=cbh.callback;besthandlerlen=len(cbh.prefix)}

			}
/*			if len(r.Selector) == 0 && besthandlerlen <= 0{
				besthandler = cbh.callback
				besthandlerlen = 0
			//}else if cbh.prefix[:len(r.Selector)-1]{
			}else if r.Selector[:len(cbh.prefix)-1] == cbh.prefix{
				besthandler = cbh.callback
				besthandlerlen = len(cbh.prefix)
			}*/
		}
		//fmt.Printf("bhl:%d prefix: %s Selector: %s\n", besthandlerlen, "???", r.Selector)
		if besthandler != nil{return besthandler(r)}
		r.WriteInfo("Internal server error")
		return errors.New("no handler available")

	}

}
