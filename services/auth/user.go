package auth

import (
	"fmt"

	"github.com/globalsign/mgo/bson"
)

// User structure in the auth info.
type User struct {
	ID                      bson.ObjectId `json:"id,omitempty"`
	Email                   string        `json:"email,omitempty"`
	EmailValidationRequired bool          `json:"email_validation_required,omitempty"`
	ImageURL                string        `json:"image_url,omitempty"`
	IsSiteAdmin             bool          `json:"is_site_admin,omitempty"`
	FirstName               string        `json:"first_name,omitempty"`
	LastName                string        `json:"last_name,omitempty"`
}

// FullName constructs user's full name from first & last name.
func (u *User) FullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}
