package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	server      = "irc.freenode.net:6667"
	userName    = "ajray15"
	channel     = "#bottest"
	maxMessages = 100
	maxMesgSize = 513
	sleepTime   = 1e6
)
// OBJECT
type IRCConn struct { // IRC server connnection, a TCP connection
	net.Conn                             // TCP Connection to write on
	*bufio.Reader                        // Buffered reads
	user          map[string]chan string // Map Username to messages for user
}
// CONSTRUCTOR
func DialIRC(addr string) (c *IRCConn, err os.Error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	} // Parse TCP dial error, ignore bufio.NewReader error below
	buf, _ := bufio.NewReaderSize(conn, maxMesgSize)
	user := make(map[string]chan string)
	return &IRCConn{conn, buf, user}, nil
}
// METHOD ON SAID OBJECT
func (irc *IRCConn) Mesg(msg string) (n int, err os.Error) {
	log.Println("Write:", msg)
	return irc.Write([]byte(msg + "\r\n"))
}
// YOU GET THE IDEA
func (irc *IRCConn) Handle(s []string, write chan string) {
	if s[1] == "PRIVMSG" && s[2] == channel {
		if len(s) < 6 {
			return
		} // no message, ignore silently
		log.Println("Handling:" + strings.Join(s, ""))
		if s[3] == ":"+userName && s[4] == "tell" { // leave message
			usr := s[5]
			if _, ok := irc.user[usr]; ok { // already seen user
				irc.user[usr] <- ":" + strings.Join(s[5:len(s)], " ")
			} else { // new user
				irc.user[usr] = make(chan string, maxMessages)
				irc.user[usr] <- ":" + strings.Join(s[5:len(s)], " ")
			}
			log.Println("Wraiting:" + strings.Join(s[4:len(s)], " "))
		}
	} else if s[1] == "JOIN" {
		usr := s[0][1:strings.Index(s[0], "!")]
		log.Println("Found: " + usr)
		if _, ok := irc.user[usr]; ok { // if we have messages
			for i := 0; i < len(irc.user[usr]); i++ {
				write <- usr + " " + <-irc.user[usr]
			}
		}
	} else if s[1] == "PING" {
        write <- "PONG" + strings.Join(s[2:len(s)], " ")
    }
}
// OMG MAIN PROGRAM
func main() {
	irc, err := DialIRC(server)
	if err != nil {
		log.Fatal("Error connecting:", err)
	}
	// identify to server
	irc.Mesg("NICK " + userName)
	irc.Mesg("USER bot * * :...")
	irc.Mesg("JOIN " + channel)
	write := make(chan string)
	// writer loop (ANONYMOUS FUNCTION SPAWNED IN A NEW GOROUTINE)
	go func() {
		for {
			irc.Mesg("PRIVMSG " + <-write)
			time.Sleep(sleepTime)
		}
	}()
	// reader loop (SPINS FOREVER IN THE MAIN GOROUTINE)
	for {
		line, _, err := irc.ReadLine()
		if err != nil {
			log.Println("Error reading:", err)
		}
		log.Println("Read:", string(line))
		go irc.Handle(strings.Split(string(line), " ", -1), write)
	}
}
