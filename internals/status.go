// Package internals is a collection of common internal services.
package internals

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/vavas/go_services/db"
	"github.com/vavas/go_services/services/intsrv"
	"bitbucket.org/telemetryapp/go_services/blackbox"
	"bitbucket.org/telemetryapp/go_services/logger"
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
			log.Ctx(blackbox.Ctx{"error": err}).Error("Status check reported DBDOWN")
			status = "DBDOWN"
		} else {
			isMaster, _ := result["ismaster"].(bool)
			primary, _ := result["primary"].(string)
			if !isMaster && len(primary) == 0 {
				log.Ctx(blackbox.Ctx{"isMaster": result}).Error("Status check reported DBDOWN")
				status = "PRIMARYDBDOWN"
			}
		}
	}

	if status == "OK" {
		log.Info("Status check reported OK")
	}

	return &intsrv.Response{Body: status}, nil
}
