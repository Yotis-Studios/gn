package gn

type ServerError struct {
	err  error
	conn *Connection
	serv *Server
}
