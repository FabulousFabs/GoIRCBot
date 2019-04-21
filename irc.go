package main

import (
    "fmt"
    "net"
    "net/textproto"
    "bufio"
    "time"
    "strings"
    )

const (
       // out commands
       IRC_OL = "%s\r\n"
       IRC_NICK = "NICK %s"
       IRC_USER = "USER %s 0 * :EmiliaIRC Bot"
       IRC_AUTH = "NS IDENTIFY %s"
       IRC_JOIN = "JOIN :%s"
       IRC_PONG = "PONG %s"
       IRC_MSG = "PRIVMSG %s %s"
       // in commands
       IRC_OP_PING = "PING"
       IRC_OP_PRIVMSG = "PRIVMSG"
       IRC_OP_JOIN = "JOIN"
       IRC_OP_QUIT = "QUIT"
       // exit codes
       IRC_EXIT_QUIT = "EXIT_REGULAR"
      )

type IRCEventCallback func(*IRCHandler, []string)

type IRCHandler struct {
    Channel string
    Server string
    Port int
    User string
    Password string
    Socket net.Conn
    PingPong bool
    EventListeners map[string]IRCEventCallback
    Events Events
    PoisonPill bool
    ReadyState bool
}

func (IRC *IRCHandler) Setup (ch string, se string, po int, us string, pw string) {
    (*IRC).Channel = ch
    (*IRC).Server = se
    (*IRC).Port = po
    (*IRC).User = us
    (*IRC).Password = pw
    (*IRC).PingPong = false
    (*IRC).EventListeners = make(map[string]IRCEventCallback)
    (*IRC).Events = Events{}
    (*IRC).Events.Prime(&(*IRC))
    (*IRC).ReadyState = false
}

func (IRC *IRCHandler) Send (message string) {
    message = fmt.Sprintf(IRC_OL, message)
    fmt.Fprintf((*IRC).Socket, message)
}

func (IRC *IRCHandler) SendAll (message string) {
    message = fmt.Sprintf(IRC_MSG, (*IRC).Channel, message)
    (*IRC).Send(message)
}

func (IRC *IRCHandler) SendPriv (message string, user string) {
    message = fmt.Sprintf(IRC_MSG, user, message)
    (*IRC).Send(message)
}

func IsOpCode(input string, code string) bool {
    if len(input) >= len(code) {
        if input[0:len(code)] == code {
            return true
        }
    }
    return false
}

func (IRC *IRCHandler) Listen () {
    defer func(){
        (*IRC).Trigger(IRC_OP_QUIT, []string{IRC_EXIT_QUIT})
        (*IRC).Socket.Close()
    }()
    
    reader := bufio.NewReader((*IRC).Socket)
    tp := textproto.NewReader(reader)
    
    for {
        if (*IRC).PoisonPill {
            break
        }
        
        line, _ := tp.ReadLine()
        
        if IsOpCode(line, IRC_OP_PING) {
            args := strings.Split(line, fmt.Sprintf("%s ", IRC_OP_PING))
            (*IRC).Send(fmt.Sprintf(IRC_PONG, args[1]))
            (*IRC).PingPong = true
            (*IRC).Trigger(IRC_OP_PING, []string{args[1]})
        } else {
            a := strings.Fields(line)
            
            if len(a) > 1 {
                comm := a[1]
                var event string
                parameters := []string{}
                
                if IsOpCode(comm, IRC_OP_PRIVMSG) {
                    host := strings.Split(a[0], "!")
                    author := host[0][1:]
                    channel := a[2]
                    message := strings.Join(a[3:], " ")[1:]
                    event = IRC_OP_PRIVMSG
                    parameters = []string{channel, author, message}
                }
                
                (*IRC).Trigger(event, parameters)
            }
        }
    }
}

func (IRC *IRCHandler) Trigger (op string, parameters []string) {
    if _, ok := (*IRC).EventListeners[op]; ok {
        (*IRC).EventListeners[op](&(*IRC), parameters)
    }
}

func (IRC *IRCHandler) On (op string, fn IRCEventCallback) {
    (*IRC).EventListeners[op] = fn
}

func (IRC *IRCHandler) Connect () {
    (*IRC).PoisonPill = false
    
    fmt.Println("Attempting to connect...")
    (*IRC).Socket, _ = net.Dial("tcp", fmt.Sprintf("%s:%d", (*IRC).Server, (*IRC).Port))
    go (*IRC).Listen()
    
    
    fmt.Println("Sending first init...")
    (*IRC).Send(fmt.Sprintf(IRC_USER, (*IRC).User))
    (*IRC).Send(fmt.Sprintf(IRC_NICK, (*IRC).User))
    
    
    fmt.Println("Waiting for pong...")
    for !(*IRC).PingPong {
        time.Sleep(250 * time.Millisecond)
    }
    time.Sleep(3500 * time.Millisecond)
    
    
    fmt.Println("Authenticating...")
    (*IRC).Send(fmt.Sprintf(IRC_AUTH, (*IRC).Password))
    
    
    fmt.Println("Joining channel...")
    (*IRC).Send(fmt.Sprintf(IRC_JOIN, (*IRC).Channel))
    time.Sleep(2500 * time.Millisecond)
    (*IRC).Trigger(IRC_OP_JOIN, []string{})
    (*IRC).ReadyState = true
}

func (IRC *IRCHandler) PollReadyState () bool {
    return (*IRC).ReadyState
}

func (IRC *IRCHandler) Disconnect () {
    (*IRC).PoisonPill = true
    (*IRC).ReadyState = false
    (*IRC).PingPong = false
}