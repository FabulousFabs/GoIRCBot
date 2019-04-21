package main

import (
    "fmt"
    )

type Events struct {
    IRC *IRCHandler
}

func (Ev *Events) Prime (IRC *IRCHandler) {
    (*IRC).On(IRC_OP_PING, EventPing)
    (*IRC).On(IRC_OP_PRIVMSG, EventMessage)
    (*IRC).On(IRC_OP_JOIN, EventJoin)
    (*IRC).On(IRC_OP_QUIT, EventQuit)
}

func EventPing(IRC *IRCHandler, args []string) {
    fmt.Printf("Ping: %s.\n", args[0])
}

func EventQuit(IRC *IRCHandler, args []string) {
    (*IRC).SendAll(args[0])
}

func EventJoin(IRC *IRCHandler, args []string) {
    greetings := fmt.Sprintf("Hi, %s! I'm %s.", (*IRC).Channel, (*IRC).User)
    (*IRC).SendAll(greetings)
}

func EventMessage(IRC *IRCHandler, args []string) {
    fmt.Println(args)
}