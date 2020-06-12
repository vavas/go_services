package db

import (
	"context"
	"os"
	"time"

	"github.com/vavas/go_services/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

// Session mongo session
type Client struct {
	client     *mongo.Client
	collection *mongo.Collection
	filter     interface{}
	fOpts      *options.FindOptions
}

var client *mongo.Client

// Database name.
var database string

// Config structure.
type Config struct {
	URL string
	DB  string
}

// Connect mongo client
func Connect(db string, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database = db
	opt := options.Client().ApplyURI(url)

	for {
		client, _ = mongo.NewClient(opt)
		err := client.Connect(ctx)
		if err != nil {
			logger.Logger.Warn("MongoDB Connection Error (retrying in 5sec)",
				zap.NamedError("error", err),
			)
			time.Sleep(10 * time.Second)
		} else {
			logger.Logger.Debug("MongoDB Connected",
				zap.String("url", url),
				zap.String("database", database),
			)
			break
		}
	}
}

// Ping verifies that the client can connect to the topology.
// If readPreference is nil then will use the client's default read
// preference.
func Ping() error {
	return client.Ping(context.Background(), readpref.Primary())
}

// DB returns a value representing the named database.
func DB() *mongo.Database {
	return client.Database(database)
}

// HasClient returns true if client is not nil.
func HasClient() bool {
	return client != nil
}

// ------------------------------------------------------------------------------------------------------------------ //

//TestConnect returns a test DB
func TestConnect() {
	port := os.Getenv("MONGODB_PORT")
	if len(port) == 0 {
		port = "27017"
	}

	// Init logging
	logger.InitLogging("db_testing")

	Connect("testing", "mongodb://127.0.0.1:" + port)
}
