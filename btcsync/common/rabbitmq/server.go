package rabbitmq

import (
	"fmt"
	"btcsync/common/log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type Server struct {
	conn *amqp.Connection
	url  string

	reconnAttempt int
	reconnLimit   int

	closes     []chan error
	errors     []chan error
	reconnects []chan bool
	// open communicates the current state of the connection. It's supposed to
	// be modified only by using atomic operations on it with 0 meaning false
	// and 1 meaning true as in standard binary. When the connection success it
	// will be 1, when there's an error or any NotifyClose event it will be 0
	// until a reconnection succeeds and then moves back to 1.
	opened bool
	mux    sync.Mutex
}

func NewServer(url string, limit int) (*Server, error) {
	r := &Server{
		url:           url,
		reconnAttempt: 0,
		reconnLimit:   limit,
		opened:        false,
	}

	if err := r.connect(); err != nil {
		return r, err
	}

	log.Info("new rabbitmq connect start")
	return r, nil
}

// Channel returns an *amqp.Channel that can be used to declare exchanges and
// publish/consume from queue.
func (r *Server) Channel() (*amqp.Channel, error) {
	conn, err := r.getConn()
	if err != nil {
		return nil, err
	}

	return conn.Channel()
}

// Shutdown returns an error if the shutdown process of the RabbitMQ connection
// returned an error. A reason is expected and it will be shown in the logs if
// *Server.log is not nil.
func (r *Server) Shutdown(reason string) error {
	//fmt.Printf("shutting down RabbitMQ connection to %s. Reason: %s", r.url, reason)
	return r.close(fmt.Errorf("shutting down RabbitMQ connection to %s. Reason: %s", r.url, reason))
}

// IsOpen returns a boolean that communicates the current state of the
// connection against the RabbitMQ server. As soon as the connection is
// established with a Server.connect() call it will be return true. During a
// amqp.NotifyClose or Server.Shutdown it will be set to false. During a
// reconnect it will be false until the reconnect succeeds.
//
// If you want to be sure the connection is open, check with Server.Open before
// making any operation against the RabbitMQ server.
func (r *Server) IsOpen() bool {
	r.mux.Lock()
	defer r.mux.Unlock()

	return r.opened
}

// Loop returns true if the connection is open or there's an active attempt to
// reconnect and get the Server to a working condition. This is specially
// useful to keep for-loops listening to the channels generated by Server as
// long as they actually exist.
func (r *Server) Loop() bool {
	if r.IsOpen() {
		return true
	}
	if r.reconnAttempt <= r.reconnLimit {
		return true
	}
	return false
}

// NotifyClose returns a receiving-only channel with the error interface that
// will be only be called once, if the RabbitMQ connection is closed. The error
// returned will be nil if the close was created by the Server.Shutdown()
// method, else it will return an error.
func (r *Server) NotifyClose() <-chan error {
	r.mux.Lock()
	defer r.mux.Unlock()

	ch := make(chan error)
	r.closes = append(r.closes, ch)
	return ch
}

// NotifyReconnect returns a receiving-only channel with true when a reconnect
// succeeds. In case of a reconnection event, the channels and queues linked to
// the Server connection need to be rebuilt.
func (r *Server) NotifyReconnect() <-chan bool {
	r.mux.Lock()
	defer r.mux.Unlock()

	ch := make(chan bool)
	r.reconnects = append(r.reconnects, ch)
	return ch
}

// NotifyErrors returns a receiving-only channel with the error interface that
// will be receive messages each time something throws an error within the
// Server representation. Receiving an error does not means something fatal
// happened. If something fatal happens the error will be received in a closing
// channel. (Check the NotifyClose's method documentation.)
func (r *Server) NotifyErrors() <-chan error {
	r.mux.Lock()
	defer r.mux.Unlock()

	ch := make(chan error)
	r.errors = append(r.errors, ch)
	return ch
}

// notifyHandler takes care of looping over the notify channel and calling the
// reconnect method.
func (r *Server) notifyHandler(notify chan *amqp.Error) {
	for {
		select {
		case err := <-notify:
			fmt.Printf("received NotifyClose with error value %v", err)
			// If the error is nil that means r.conn.Close was called so we
			// should not try to reconnect.
			if err == nil && !r.opened {
				return
			}
			// The reconnection will only be called if r.open is not already 0,
			// if it's already 0 then reconnect is very likely already running.
			//
			// As an extra precaution this section is locked with a sync.Mutex
			// so no other call would run it.
			r.mux.Lock()
			if err != nil && r.opened {
				go r.sendErrors(err)
				r.reconnect()
			}
			r.opened = false
			r.mux.Unlock()

			break
		}
	}
}

// connect returns an error if the connection process had any issue during the
// amqp.Dial call.
func (r *Server) connect() error {
	conn, err := amqp.Dial(r.url)
	if err != nil {
		go r.sendErrors(err)
		return err
	}

	r.conn = conn
	r.opened = true

	notify := conn.NotifyClose(make(chan *amqp.Error))
	go r.notifyHandler(notify)

	return nil
}

// reconnect will handle all the reconnection logic by receiving as argument
// the *amqp.Error value handed by the *amqp.Connection.NotifyClose method. If
// the argument passed is nil it won't try to reconnect since it will assume
// the nil was sent because *amqp.Connection.Close was called.
func (r *Server) reconnect() error {
	r.reconnAttempt++

	if r.reconnAttempt > r.reconnLimit {
		return r.close(fmt.Errorf("rabbit: can't reconnect to server %s, tried %d times", r.url, r.reconnLimit))
	}

	time.Sleep(5 * time.Second)
	if err := r.connect(); err != nil {
		fmt.Printf(`couldn't reconnect to RabbitMQ server.Address:	%s Attempt:	%d Error: %v`, r.url, r.reconnAttempt, err)
		go r.sendErrors(err)
		r.reconnect()
	}

	for _, ch := range r.reconnects {
		ch <- true
	}

	r.reconnAttempt = 0

	return nil
}

// connection returns an *amqp.Connection pointer that can be passed to other
// methods and take advantage of the reconnection handling implemented by the
// rabbit package.
func (r *Server) getConn() (*amqp.Connection, error) {
	if r.conn == nil {
		return nil, fmt.Errorf(ENOCONN)
	}
	return r.conn, nil
}

// sendErrors is a handy wrapper for sending errors to all the error channels
// opened up.
func (r *Server) sendErrors(err error) {
	for _, ch := range r.errors {
		ch <- err
	}
}

// close returns an error if Server.conn.Close() method goes wrong in any way.
// This method will be called by either the Server.reconnect() method in case
// it can't reconnect after the limit of retries or if Server.Shutdown() was
// called externally.
func (r *Server) close(err error) error {

	r.mux.Lock()
	defer r.mux.Unlock()

	r.opened = false
	// closing the close channels after notifying of the closure.
	for _, ch := range r.closes {
		close(ch)
	}

	conn, err := r.getConn()
	if err != nil {
		return err
	}

	cerr := conn.Close()
	if cerr != nil {
		return cerr
	}

	// closing the other channels.
	for _, ch := range r.reconnects {
		close(ch)
	}
	for _, ch := range r.errors {
		close(ch)
	}

	r.conn = nil

	return nil
}
