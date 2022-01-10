package user

import "context"

// UserService represents a service for managing users.
type UserService interface {
	// Retrieves a user by ID along with their associated auth objects.
	// Returns ENOTFOUND if user does not exist.
	FindUserByID(ctx context.Context, id int) (*User, error)

	// Retrieves a list of users by filter. Also returns total count of matching
	// users which may differ from returned results if filter.Limit is specified.
	FindUsers(ctx context.Context, filter UserFilter) ([]*User, int, error)

	// Creates a new user. This is only used for testing since users are typically
	// created during the OAuth creation process in AuthService.CreateAuth().
	CreateUser(ctx context.Context, user *User) error

	// Updates a user object. Returns EUNAUTHORIZED if current user is not
	// the user that is being updated. Returns ENOTFOUND if user does not exist.
	UpdateUser(ctx context.Context, id int, upd UserUpdate) (*User, error)

	// Permanently deletes a user and all owned shops. Returns EUNAUTHORIZED
	// if current user is not the user being deleted. Returns ENOTFOUND if
	// user does not exist.
	DeleteUser(ctx context.Context, id int) error
}

// AuthService represents a service for managing auths.
type AuthService interface {
	// Looks up an authentication object by ID along with the associated user.
	// Returns ENOTFOUND if ID does not exist.
	FindAuthByID(ctx context.Context, id int) (*Auth, error)

	// Retrieves authentication objects based on a filter. Also returns the
	// total number of objects that match the filter. This may differ from the
	// returned object count if the Limit field is set.
	FindAuths(ctx context.Context, filter AuthFilter) ([]*Auth, int, error)

	// Creates a new authentication object If a User is attached to auth, then
	// the auth object is linked to an existing user. Otherwise a new user
	// object is created.
	//
	// On success, the auth.ID is set to the new authentication ID.
	CreateAuth(ctx context.Context, auth *Auth) error

	// Permanently deletes an authentication object from the system by ID.
	// The parent user object is not removed.
	DeleteAuth(ctx context.Context, id int) error
}
