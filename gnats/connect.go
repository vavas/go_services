// Package gnats is a library for handling gnatsd connection & sessions.
package gnats

import (
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/pkg/errors"

	"bitbucket.org/telemetryapp/go_services/logger"
)

var connection = &lockedConnection{}
var onActiveHandlers = map[string]OnActiveHandler{}

// OnActiveHandler is a function that can passed to AddOnActiveHandler. When
// A nats connection is established this function will be executed.
type OnActiveHandler func() error

// Config contains configuration options that can be used when connecting to
// nats using this package.
type Config struct {
	Name string
	Urls []string
}

// IsConnected returns true if connection to gnatsd server is OK.
func IsConnected() bool {
	return connection.getBare().IsConnected()
}

// JSONConn returns the JSON encoded connection.
func JSONConn() (*nats.EncodedConn, error) {
	if encodedConnection := connection.getEncoded(); encodedConnection != nil {
		return encodedConnection, nil
	}

	encodedConnection, err := nats.NewEncodedConn(connection.getBare(), nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	}

	connection.setEncoded(encodedConnection)

	return encodedConnection, nil
}

// AddOnActiveHandler adds a function to run when connection to gnatsd server becomes active
// (after connected or re-connected).
func AddOnActiveHandler(name string, onActiveHandler OnActiveHandler) {
	if _, ok := onActiveHandlers[name]; ok {
		logger.Logger.Fatalf("OnActiveHandler for '%s' has been registered", name)
	}
	onActiveHandlers[name] = onActiveHandler
}

// ClearOnActiveHandlers removes all onActiveHandlers
func ClearOnActiveHandlers() {
	// set to initial condition (empty map)
	onActiveHandlers = map[string]OnActiveHandler{}
}

// Connect takes a config object and creates a new nats connection, using it for all nats
// communication required by this package.
func Connect(conf *Config) error {
	log := logger.Logger

	opts := nats.GetDefaultOptions()
	opts.Servers = conf.Urls

	conn, err := nats.Connect(
		strings.Join(conf.Urls, ","),
		nats.Name(conf.Name),
		nats.MaxReconnects(-1),
		// TODO: define these handlers outside of Connect
		// How to insert handlers
		// nc, err = nats.Connect(servers,
		// 	nats.DisconnectHandler(func(_ *nats.Conn) {
		// 		fmt.Printf("Got disconnected!\n")
		// 	}),
		// 	nats.ReconnectHandler(func(nc *nats.Conn) {
		// 		fmt.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
		// 	}),
		// nats.ClosedHandler(func(nc *nats.Conn) {
		// fmt.Printf("Connection closed. Reason: %q\n", nc.LastError())
		// })
		// )
		nats.ClosedHandler(func(conn *nats.Conn) {
			log.Debugf("nats connection closed: %s", conn.LastError())
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Debug("nats connection re-established")
		}),
		nats.DiscoveredServersHandler(func(conn *nats.Conn) {
			log.Debugf("nats found new server at url %s", conn.ConnectedUrl())
		}),
		nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
			log.Errorf(
				"nats communication error on subject %s: with server at %s",
				subscription.Subject,
				conn.ConnectedUrl(),
			)
		}),
	)
	if err != nil {
		return errors.Wrap(err, "nats.Connect")
	}

	for name, onActiveHandler := range onActiveHandlers {
		if err := onActiveHandler(); err != nil {
			log.WithError(err).Errorf("Error running onActiveHandler '%s'", name)
		}
	}

	connection.setBare(conn)

	return nil
}

// TestConnect is a connect method used for testing purposes. Do not use this
// method in production code.
func TestConnect() {
	conf := &Config{Urls: []string{"nats://127.0.0.1:4222"}}
	_ = Connect(conf)
}

// SetConnection is a function that allows setting an already established
// connection for use.
func SetConnection(conn *nats.Conn) error {
	if !connection.getBare().IsConnected() {
		return errors.New("Connection must already be established")
	}

	connection.setBare(conn)

	return nil
}

// Disconnect disconnects the current nats connection.
func Disconnect() {
	logger.Logger.Warn("called disconnect")
	connection.getBare().Close()
	UnsetConnection()
}

// UnsetConnection clears the connection without disconnecting. This is meant
// for use with SetConnection. It's here so if you need to run teardown logic
// on your connection you can do so without this package interfering.
func UnsetConnection() {
	connection.setBare(nil)
	connection.setEncoded(nil)
}

// StartReconnectMonitor was used to start a reconnect loop for nats.
// Nats has it's own reconnect logic, so this has been converted to a noop.
func StartReconnectMonitor() {
	// NOOP
}
