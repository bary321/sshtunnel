package main

import (
	"fmt"
	"github.com/elliotchance/sshtunnel"
	"log"
	"os"
	"time"
)

func main() {

	//REMOTE_SERVER_IP := "139.159.245.39"
	//PRIVATE_SERVER_IP := "220.195.127.166"
	tunnel := sshtunnel.NewSSHTunnel(
		// User and host of tunnel server, it will default to port 22
		// if not specified.
		"pxy1@139.159.245.39:56566",

		// Pick ONE of the following authentication methods:
		sshtunnel.PrivateKeyFile("E:\\ProgramFile\\supervisord\\sshtunnelT\\p1"), // 1. private key
		//ssh.Password("password"),                            // 2. password

		// The destination host and port of the actual server.
		"220.195.127.166:7788",

		// The local port you want to bind the remote port to.
		// Specifying "0" will lead to a random port.
		"6677",
	)
	ll := log.New(os.Stdout, "", 1)
	tunnel.Log = ll

	err := tunnel.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		time.Sleep(3600 * time.Second)
	}
}
