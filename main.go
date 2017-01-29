package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var dir string
var port int
var s *Server

func main() {
	flag.IntVar(&port, "port", 8000, "port number")
	flag.IntVar(&port, "p", 8000, "port number")
	flag.StringVar(&dir, "d", "", "directory name include template files")
	flag.Parse()

	s = NewServer()
	go s.Start()

	listener, ch := server(":" + fmt.Sprint(port))
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	go func() {
		log.Println(<-sig)
		listener.Close()
	}()
	log.Println(<-ch)
}

func server(addr string) (listener net.Listener, ch chan error) {
	ch = make(chan error)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", chatRoom)
		mux.HandleFunc("/signup", signUp)
		mux.HandleFunc("/signin", signIn)
		mux.Handle("/ws", s.WebsocketHandler())
		ch <- http.Serve(listener, mux)
	}()
	return
}

func signIn(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	hash, err := GetUserHash(email, password)
	if err == nil {
		c := &http.Cookie{
			Name:  "username",
			Value: hash,
		}
		http.SetCookie(w, c)
		http.Redirect(w, r, "/room", http.StatusSeeOther)
		return
	}
	fmt.Println(err)
	if f, e := os.Open("assets/signin.html"); e == nil {
		io.Copy(w, f)
	}
}

func signUp(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	if username != "" && email != "" && password != "" {
		if err := RegisterUser(username, email, password); err != nil {
			fmt.Println(err)
			if f, err := os.Open("assets/signup.html"); err == nil {
				io.Copy(w, f)
			}
		}
		c := &http.Cookie{
			Name:  "username",
			Value: fmt.Sprintf("%x", sha256.Sum256([]byte(email+password))),
		}
		http.SetCookie(w, c)
		http.Redirect(w, r, "/room", http.StatusSeeOther)
		return
	}
	if f, e := os.Open("assets/signup.html"); e == nil {
		io.Copy(w, f)
	}
}

func chatRoom(w http.ResponseWriter, r *http.Request) {
	info := "Time:" + fmt.Sprint(time.Now()) + "\t"
	info = info + "Url:" + r.Host + r.URL.String() + "\t"
	info = info + "Proto:" + r.Proto + "\t"
	info = info + "RemoteAddr:" + r.RemoteAddr + "\t"
	info = info + "Header:" + fmt.Sprint(r.Header) + "\t"
	if r.PostForm != nil {
		info = info + "PostForm:" + r.PostForm.Encode() + "\t"
	}
	if r.Trailer != nil {
		info = info + "Trailer:" + fmt.Sprint(r.Trailer) + "\t"
	}
	fmt.Println(info)
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/html")
	var u string
	c, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	} else {
		u = c.Value
	}
	var m []Message
	if _, err := os.Stat("message.json"); err == nil {
		if f, err := os.Open("message.json"); err == nil {
			if b, err := ioutil.ReadAll(f); err == nil {
				err := json.Unmarshal(b, &m)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	data := Data{
		Port:     fmt.Sprint(port),
		HostName: r.Host,
		UserName: u,
		Messages: m,
	}
	t := template.Must(template.ParseFiles("assets/room.html"))
	if err := t.Execute(w, data); err != nil {
		Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Error(e error) {
	fmt.Fprintln(os.Stderr, "Time:"+fmt.Sprint(time.Now())+"\t"+fmt.Sprint(e))
}
