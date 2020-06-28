package sshtunnel

import (
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"sync"
)

type SSHTunnel struct {
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint
	Config *ssh.ClientConfig
	Log    *log.Logger
	close  chan interface{}
}

func (tunnel *SSHTunnel) logf(fmt string, args ...interface{}) {
	if tunnel.Log != nil {
		tunnel.Log.Printf(fmt, args...)
	}
}

func (tunnel *SSHTunnel) Start() error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return err
	}
	tunnel.Local.Port = listener.Addr().(*net.TCPAddr).Port
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				tunnel.logf("accept error", err)
				return
			}
			tunnel.logf("accepted connection")
			wg.Add(1)
			go tunnel.forward(conn, &wg)
			select {
			case <-tunnel.close:
				break
			default:

			}
		}
	}()

	wg.Wait()

	tunnel.logf("tunnel closed")
	err = listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (tunnel *SSHTunnel) forward(localConn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		tunnel.logf("server dial error: %s", err)
		return
	}
	tunnel.logf("connected to %s (1 of 2)\n", tunnel.Server.String())
	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		tunnel.logf("remote dial error: %s", err)
		return
	}
	tunnel.logf("connected to %s (2 of 2)\n", tunnel.Remote.String())
	copyConn := func(writer, reader net.Conn) {
		_, err := io.Copy(writer, reader)
		if err != nil {
			tunnel.logf("%v %v io.Copy error: %s", writer.LocalAddr(), reader.LocalAddr(), err)
			writer.Close()
			reader.Close()
		}
	}
	go copyConn(localConn, remoteConn)
	copyConn(remoteConn, localConn)
	//wg.Done()
	return
}

func (tunnel *SSHTunnel) Close() {
	tunnel.close <- struct{}{}
	return
}

// NewSSHTunnel creates a new single-use tunnel. Supplying "0" for localport will use a random port.
func NewSSHTunnel(tunnel string, auth ssh.AuthMethod, destination string, localport string) *SSHTunnel {

	localEndpoint := NewEndpoint("localhost:" + localport)

	server := NewEndpoint(tunnel)
	if server.Port == 0 {
		server.Port = 22
	}

	sshTunnel := &SSHTunnel{
		Config: &ssh.ClientConfig{
			User: server.User,
			Auth: []ssh.AuthMethod{auth},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// Always accept key.
				return nil
			},
		},
		Local:  localEndpoint,
		Server: server,
		Remote: NewEndpoint(destination),
		close:  make(chan interface{}),
	}

	return sshTunnel
}
