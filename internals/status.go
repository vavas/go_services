// Package internals is a collection of common internal services.
package internals

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/vavas/go_services/db"
	"github.com/vavas/go_services/logger"
	"github.com/vavas/go_services/services/intsrv"
)

// SetupStatus adds "Status" internal service. It will return a json `"OK"`.
func SetupStatus(sub *intsrv.Subscription) {
	sub.AddHandler("Status", status)
}

func status(_ *mongo.Database, _ *intsrv.Request) (*intsrv.Response, error) {
	log := logger.Logger
	status := "OK"

	if db.HasClient() {
		db := db.DB()

		result := primitive.M{}
		if err := db.RunCommand(nil, primitive.D{{"isMaster", 1}}).Decode(&result); err != nil {
			log.Error("Status check reported DBDOWN",
				zap.Error(err))
			status = "DBDOWN"
		} else {
			isMaster, _ := result["ismaster"].(bool)
			primary, _ := result["primary"].(string)
			if !isMaster && len(primary) == 0 {
				log.Error("Status check reported DBDOWN",
					zap.Any("isMaster", result))
				status = "PRIMARYDBDOWN"
			}
		}
	}

	if status == "OK" {
		log.Info("Status check reported OK")
	}

	return &intsrv.Response{Body: status}, nil
}
