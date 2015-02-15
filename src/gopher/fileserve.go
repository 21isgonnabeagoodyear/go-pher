package gopher
import "os"
import "path"
import "path/filepath"
import "errors"
import "strings"
import "io/ioutil"//has ReadDir for some reason
import "io"//has ReadDir for some reason


func filereqhandler(r *Request, basepath string, prefix RequestHandler) error{
	fullpath := path.Join(basepath, r.Selector)
	fullpath, err := filepath.EvalSymlinks(fullpath)
	if err != nil{return err}
	fullpath = filepath.Clean(fullpath)
	fullpath, err = filepath.Abs(fullpath)
	if err != nil{return err}
	if !strings.HasPrefix(fullpath, basepath){return errors.New("shenanigans")}
	info, err := os.Lstat(fullpath)
	if err != nil{return err}
	if info.Mode() & os.ModeDir != 0{
		r.isdir = byte('y')//cheating
		if prefix != nil{prefix(r)}

		parentpath, err := filepath.Rel(basepath, path.Dir(fullpath))
		if err != nil{return err}
		r.WriteEntry(Direntry{Code:DIRECTORY, Description:"..", Selector:parentpath, Host:r.MyDomain, Port:r.MyPort})

		files, err := ioutil.ReadDir(fullpath)
		if err != nil{return err}

		for _, file :=range files{
			if strings.HasSuffix(file.Name(), "~"){continue}//ignore backup files (~)
			if file.Name()[0] == '.'{continue}
			if file.Mode() & os.ModeDir != 0{//TODO:check if it's a symlink?
				pathname, err := filepath.Rel(basepath, path.Join(fullpath, file.Name()))
				if err != nil{return err}
				r.WriteEntry(Direntry{Code:DIRECTORY, Description:file.Name(), Selector:pathname, Host:r.MyDomain, Port:r.MyPort})
			}else if file.Mode() & os.ModeType ==0{
				pathname, err := filepath.Rel(basepath, path.Join(fullpath, file.Name()))
				if err != nil{return err}
				r.WriteEntry(Direntry{Code:FILE, Description:file.Name(), Selector:pathname, Host:r.MyDomain, Port:r.MyPort})
			}
		}
	}else if info.Mode() & os.ModeType == 0{//no ModeType bits set = regular file
		//fmt.Println(mime.TypeByExtension(filepath.Ext(r.Selector)))
		//fmt.Println(filepath.Ext(r.Selector))
		//TODO:use http.DetectContentType to get filetype and choose between FILE and binary types
		f, err:= os.Open(fullpath)
		if err != nil{return err}
		defer f.Close()
		_, err  = io.Copy(r, f)
		return err
	}else{
		r.WriteError("no such file or directory")
		return errors.New("no such file or directory")
	}
	return nil
}

//NewFileServer returns a RequestHandler which may be passed to ListenAndServe to serve files in a directory.  
//Currently all files are served with type FILE, so binary files may be interpretted incorrectly by clients.
func NewFileServer(base string) RequestHandler{
	return func(r *Request) error{return filereqhandler(r,base, nil)}
}
//NewFileServerWithHeader returns a RequestHandler which may be passed to ListenAndServe to serve files in a directory.  
//The RequestHandler prefix will be used to generate headers for the directory listings.
func NewFileServerWithHeader(base string, prefix RequestHandler) RequestHandler{
	return func(r *Request) error{return filereqhandler(r,base, prefix)}
}
