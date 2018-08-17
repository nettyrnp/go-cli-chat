package main


import (
	"bufio"
	"net"
	"os"
	"time"
	"sync"
	"io"
	"log"
	"errors"
	"strings"
)

const (
	delay = 100 * time.Millisecond
)

type Server struct {
	clients map[int]*Client
	lastId int
	entrance chan net.Conn
	incoming chan string
	outgoing chan string
	serverMutex sync.Mutex
}

func NewServer() *Server {
	server := &Server{
		clients: make(map[int]*Client),
		lastId: -1,
		entrance: make(chan net.Conn),
		incoming: make(chan string),
		outgoing: make(chan string),
	}
	return server
}

func (server *Server) Broadcast(data string) {
	for _, client := range server.clients {
		client.outgoing <- data
	}
}

func (server *Server) Join(connection net.Conn) {
	server.serverMutex.Lock()
	defer server.serverMutex.Unlock()
	newClientId := server.lastId + 1
	server.lastId = newClientId
	client := NewClient(newClientId, server, connection)

	name, err := getParticipantName(client, time.Second * 5)
	if err!=nil {
		if err.Error() == "timed out" {
			log.Println("Timed out. Destroy client, id=", client.id)
			client.LeaveChat()
		}
		log.Fatalln(err)
	}
	client.name = name

	client.Listen()

	_, keyExist := server.clients[newClientId]
	if ! keyExist {
		server.clients[newClientId] = client
	}

	server.Broadcast("*** " + client.name + " is online\n\n")

	// Here client can write to other clients
	go func() {
		for {
			select {
			case <-client.stopServerWriting:
				return
			case data := <-client.incoming:
				data = client.name + ": \t" + data
				server.incoming <- data
			}
		}
	}()
}

func (server *Server) Listen() {
	// No need for a channel to stop itself
	go func() {
		for {
			select {
			case data := <-server.incoming:
				server.Broadcast(data)
			case conn := <-server.entrance:
				server.Join(conn)
			}
		}
	}()
}

type Client struct {
	id               int
	name              string
	server            *Server
	connection        net.Conn
	incoming          chan string
	outgoing          chan string
	reader            *bufio.Reader
	writer            *bufio.Writer
	stopServerWriting chan bool
	stopClientReading chan bool
	stopClientWriting chan bool
	clientMutex       sync.Mutex
}

func NewClient(id int, server *Server, connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	client := &Client{
		id:               id,
		name:              "",
		server:            server,
		connection:        connection,
		incoming:          make(chan string),
		outgoing:          make(chan string),
		reader:            reader,
		writer:            writer,
		stopServerWriting: make(chan bool),
		stopClientReading: make(chan bool),
		stopClientWriting: make(chan bool),
	}
	return client
}

func (client *Client) Read() {
	for {
		select {
		case <-client.stopClientReading:
			return
		default:
			line, err := client.ReadOnce()
			if err!=nil {
				time.Sleep(delay * time.Millisecond)
				continue
			}
			if line != "" {
				client.incoming <- line
			}
		}
	}
}

func (client *Client) ReadOnce() (string, error) {
	line, err := client.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			client.server.Broadcast("*** " + client.name + " is offline\n\n")
			client.LeaveChat()
		}
		log.Println(err)
		return "", err
	}
	return line, nil
}

func getParticipantName(client *Client, timeout time.Duration) (string, error) {
	client.WriteOnce(">> Please type in your name\n")

	timeoutCh := time.After(timeout)
	tickCh := time.Tick(delay)
	for {
		select {
		case <-timeoutCh:
			client.WriteOnce(">> Client timed out\n")
			return "", errors.New("timed out") // TODO spec error
		case <-tickCh:
			name, err := client.ReadOnce()
			if err!=nil {
				return "", err
			}
			name = strings.TrimSpace(name)
			if name == "" {
				client.WriteOnce(">> A name should be none-empty!\n")
				continue
			}
			return name, nil
		}
	}
	panic("Something went wrong. Should never reach this line")
}

func (client *Client) Write() {
	for {
		select {
		case <-client.stopClientWriting:
			return
		case data := <-client.outgoing:
			client.WriteOnce(data)
		}
	}
}

func (client *Client) WriteOnce(data string) {
	client.writer.WriteString(data)
	client.writer.Flush()
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

func (client *Client) LeaveChat() {
	// Unregister client from server
	server := *client.server
	id := client.id
	server.serverMutex.Lock()
	defer server.serverMutex.Unlock()
	delete(server.clients, id)

	// Destroy client
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	client.stopClientReading <- true
	client.stopClientWriting <- true
	client.reader = nil
	client.writer = nil
	client.connection.Close()
	client.connection = nil
	client.stopServerWriting <- true
}



func main() {
	port := "50505"
	listener, err_listen := net.Listen("tcp", ":" + port)
	if err_listen != nil {
		log.Println("Server failed")
		os.Exit(1)
	}

	server := NewServer()
	server.Listen()
	log.Println("Server is listening on port " + port)

	for {
		conn, err_ac := listener.Accept()
		if err_ac != nil {
			log.Println("Connection failed")
			conn.Close()
			time.Sleep(delay)
			continue
		}
		conn.SetDeadline(time.Now().Add(time.Hour*24))
		server.entrance <- conn
	}
}

