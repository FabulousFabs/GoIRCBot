package main

import (
    "fmt"
    "bufio"
    "os"
    )

func main () {
    // setup IRC
    IRC := IRCHandler{}
    IRC.Setup("#EmiliaIRC", "irc.snoonet.org", 6667, "EmiliaIRC", "EmiliaIRCBot912956")
    
    // connect IRC
    go IRC.Connect()
    for !IRC.PollReadyState() {}
    fmt.Printf("Done.\n")
    
    // console
    reader := bufio.NewReader(os.Stdin)
    for {
        text, _ := reader.ReadString('\n')
        
        if text[0:1] == "/" {
            switch comm := text[1:len(text)-1]; comm {
                case "go":
                    if IRC.PollReadyState() {
                        continue
                    }
                    go IRC.Connect()
                    for !IRC.PollReadyState() {}
                    fmt.Printf("Done.\n")
                case "quit":
                    IRC.Disconnect()
                    fmt.Println("Bye.")
                default:
                    fmt.Println("Unknown command.")
            }
        } else {
            IRC.SendAll(text)
        }
    }
}