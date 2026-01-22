package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

// Timing settings
const (
	DefaultMaxConnections      = 10
	DefaultMaxIdle             = time.Minute
	DefaultSweepInterval       = time.Second * 70
	DefaultScalingFactor  uint = 2
	DefaultMaxRetries     int  = 3
)

const (
	DefaultHostname = "dv.net"
	DefaultIdentity = "localhost-client"
)

var (
	ErrNoAvailableConnections = errors.New("no available SMTP connections")
	ErrMaxConnectionsReached  = errors.New("pool maximum SMTP connections reached")
	ErrScalingDisabled        = errors.New("pool scaling is disabled")
	ErrShrinkThresholdReached = errors.New("pool shrink threshold reached. Default being 10")
	ErrPoolClosed             = errors.New("pool is closed")
)

type IPool interface {
	Send(string, []string, io.Reader) error
}

var _ IPool = (*SMTPPool)(nil)

type PoolOptions struct {
	Address       string        `json:"address"` // SMTP server address ex. being (smtp.example.com:1025)
	MaxConn       uint          `json:"max_conn"`
	Hostname      string        `json:"hostname"`
	SSL           bool          `json:"ssl"`
	Username      *string       `json:"username"`
	Password      *string       `json:"password"`
	Identity      string        `json:"identity"`
	Auth          *sasl.Client  `json:"auth"` // Used to authenticate clients to the server
	TLSConfig     *tls.Config   `json:"tls_config"`
	MaxIdle       time.Duration `json:"max_idle"`
	MaxRetries    int           `json:"max_retries"`
	SweepInterval time.Duration `json:"sweep_interval"` // Time interval to sweep idle connections
	Scaling       bool          `json:"scaling"`        // If true, pool will dynamically grow and shrink itself
	ScalingFactor uint          `json:"scaling_factor"` // Scaling factor for dynamic pool size
}

type SMTPPool struct {
	opt         *PoolOptions
	cons        chan *conn
	activeConn  atomic.Int64
	createdConn atomic.Int64
	stopCh      chan bool
	closed      atomic.Bool
}

func NewSMTPPool(ctx context.Context, o *PoolOptions) (*SMTPPool, error) {
	pool := &SMTPPool{
		opt:    o,
		stopCh: make(chan bool),
	}
	if o.Identity == "" {
		o.Identity = DefaultIdentity
	}
	if o.MaxRetries == 0 {
		o.MaxRetries = DefaultMaxRetries
	}
	if o.Hostname == "" {
		o.Hostname = DefaultHostname
	}
	if o.MaxConn == 0 {
		o.MaxConn = DefaultMaxConnections
	}
	if o.MaxIdle == 0 {
		o.MaxIdle = DefaultMaxIdle
	}
	if o.SweepInterval == 0 {
		o.SweepInterval = DefaultSweepInterval
	}
	if o.ScalingFactor == 0 {
		o.ScalingFactor = DefaultScalingFactor
	}

	pool.cons = make(chan *conn, o.MaxConn)

	if o.Username != nil && o.Password != nil {
		authClient := sasl.NewPlainClient("", *o.Username, *o.Password)
		o.Auth = &authClient
	}

	go pool.run(ctx)

	return pool, nil
}

func (o *SMTPPool) Close() error {
	o.closed.Store(true)
	close(o.stopCh)
	return o.close()
}

func (o *SMTPPool) run(ctx context.Context) {
	sweepTicker := time.NewTicker(o.opt.SweepInterval)
	defer sweepTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stopCh:
			return
		case <-sweepTicker.C:
			_ = o.releaseDangling()
		}
	}
}

func (o *SMTPPool) Send(from string, to []string, body io.Reader) error {
	connection, err := o.get()
	if err != nil {
		return err
	}
	defer func() {
		if err := o.put(connection); err != nil {
			_ = connection.Close()
		}
	}()

	if err := connection.Mail(from); err != nil {
		return err
	}

	if err := connection.Rcpt(to); err != nil {
		return err
	}

	w, err := connection.Data()
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, body); err != nil {
		return err
	}

	return w.Close()
}

func (o *SMTPPool) close() error {
	var err error
	close(o.cons)
	for connection := range o.cons {
		if closeErr := connection.Close(); closeErr != nil {
			err = closeErr
		}
	}
	return err
}

func (o *SMTPPool) get() (*conn, error) {
	if o.closed.Load() {
		return nil, ErrPoolClosed
	}

	select {
	case connection := <-o.cons:
		o.activeConn.Add(1)
		return connection, nil
	default:
		if int(o.activeConn.Load()) >= int(o.opt.MaxConn) { //nolint:gosec
			return nil, ErrNoAvailableConnections
		}
		connection, err := o.newConn()
		if err != nil {
			return nil, err
		}
		o.activeConn.Add(1)
		return connection, nil
	}
}

func (o *SMTPPool) put(connection *conn) error {
	if o.closed.Load() {
		return ErrPoolClosed
	}
	select {
	case o.cons <- connection:
		o.activeConn.Add(-1)
		return nil
	default:
		o.activeConn.Add(-1)
		return connection.Close()
	}
}

func (o *SMTPPool) newConn() (*conn, error) {
	var (
		client *smtp.Client
		err    error
	)
	if o.opt.SSL && o.opt.TLSConfig != nil {
		client, err = smtp.DialStartTLS(o.opt.Address, o.opt.TLSConfig)
	} else {
		client, err = smtp.Dial(o.opt.Address)
	}
	if err != nil {
		return nil, err
	}

	localName := o.opt.Hostname + "-" + strconv.FormatInt(o.createdConn.Add(1), 10)
	if err := client.Hello(localName); err != nil {
		_ = client.Close()
		return nil, err
	}

	if o.opt.Auth != nil {
		if err := client.Auth(*o.opt.Auth); err != nil {
			_ = client.Close()
			return nil, err
		}
	}

	return &conn{
		client:     client,
		lastActive: time.Now(),
		localName:  localName,
	}, nil
}

func (o *SMTPPool) releaseDangling() error {
	for len(o.cons) > 0 {
		select {
		case connection := <-o.cons:
			if time.Since(connection.lastActive) > o.opt.MaxIdle {
				if err := connection.Close(); err != nil {
					o.cons <- connection
					return err
				}
				o.activeConn.Add(-1)
			} else {
				o.cons <- connection
			}
		default:
			return nil
		}
	}
	return nil
}

type conn struct {
	client     *smtp.Client
	lastActive time.Time
	localName  string
}

func (o *conn) Mail(from string) error {
	o.lastActive = time.Now()
	return o.client.Mail(from, nil)
}

func (o *conn) Rcpt(to []string) error {
	for _, rcpt := range to {
		if err := o.client.Rcpt(rcpt, nil); err != nil {
			return err
		}
	}
	o.lastActive = time.Now()
	return nil
}

func (o *conn) Data() (io.WriteCloser, error) {
	o.lastActive = time.Now()
	return o.client.Data()
}

func (o *conn) Close() error {
	return o.client.Close()
}

func (o *conn) Quit() error {
	return o.client.Quit()
}

func (o *conn) Rst() error {
	o.lastActive = time.Now()
	return o.client.Reset()
}
