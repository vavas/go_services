package auth

import (
	"github.com/globalsign/mgo/bson"
)

// Administrator structure in the auth info.
type Administrator struct {
	ID bson.ObjectId `json:"id,omitempty"`

	Realm string   `json:"realm,omitempty"`
	Email string   `json:"email,omitempty"`
	Scope []string `json:"scope,omitempty"`
}
