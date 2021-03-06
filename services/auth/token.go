// Package auth is the auth info that is passed from auth service to other services.
package auth

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/vavas/go_services/datetime"
)

// Token structure in the auth info.
type Token struct {
	ID          primitive.ObjectID `json:"id,omitempty"`
	Scope       []string           `json:"scope,omitempty"`
	Name        string             `json:"name,omitempty"`
	Expires     *datetime.DT       `json:"expires,omitempty"`
	Permissions []string           `json:"permissions,omitempty"`
}
