package user

import (
	"context"
	"errors"
)

var (
	ErrUserNotFound = errors.New("User Not Found")
	ErrUserInvalid  = errors.New("User Invalid")
)

type userService struct {
	userRepo UserRepository
}

type authService struct {
	authRepo AuthRepository
}

func NewUserService(userRepo UserRepository) UserService {
	return &userService{
		userRepo,
	}
}

// Retrieves a user by ID along with their associated auth objects.
// Returns ENOTFOUND if user does not exist.
func (r *userService) FindUserByID(ctx context.Context, id int) (*User, error) {
	return r.userRepo.FindUserByID(ctx, id)
}

// Retrieves a list of users by filter. Also returns total count of matching
// users which may differ from returned results if filter.Limit is specified.
func (r *userService) FindUsers(ctx context.Context, filter UserFilter) ([]*User, int, error) {
	return r.userRepo.FindUsers(ctx, filter)
}

// Creates a new user. This is only used for testing since users are typically
// created during the OAuth creation process in AuthService.CreateAuth().
func (r *userService) CreateUser(ctx context.Context, user *User) error {
	return r.userRepo.CreateUser(ctx, user)
}

// Updates a user object. Returns EUNAUTHORIZED if current user is not
// the user that is being updated. Returns ENOTFOUND if user does not exist.
func (r *userService) UpdateUser(ctx context.Context, id int, upd UserUpdate) (*User, error) {
	return r.userRepo.UpdateUser(ctx, id, upd)
}

// Permanently deletes a user and all owned shops. Returns EUNAUTHORIZED
// if current user is not the user being deleted. Returns ENOTFOUND if
// user does not exist.
func (r *userService) DeleteUser(ctx context.Context, id int) error {
	return r.userRepo.DeleteUser(ctx, id)
}

func (r *authService) FindAuthByID(ctx context.Context, id int) (*Auth, error) {
	return r.authRepo.FindAuthByID(ctx, id)
}

// Retrieves authentication objects based on a filter. Also returns the
// total number of objects that match the filter. This may differ from the
// returned object count if the Limit field is set.
func (r *authService) FindAuths(ctx context.Context, filter AuthFilter) ([]*Auth, int, error) {
	return r.authRepo.FindAuths(ctx, filter)
}

// Creates a new authentication object If a User is attached to auth, then
// the auth object is linked to an existing user. Otherwise a new user
// object is created.
//
// On success, the auth.ID is set to the new authentication ID.
func (r *authService) CreateAuth(ctx context.Context, auth *Auth) error {
	return r.authRepo.CreateAuth(ctx, auth)
}

// Permanently deletes an authentication object from the system by ID.
// The parent user object is not removed.
func (r *authService) DeleteAuth(ctx context.Context, id int) error {
	return r.authRepo.DeleteAuth(ctx, id)
}

// Validate returns an error if the user contains invalid fields.
// This only performs basic validation.
func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("User Invalid")
	}
	return nil
}

// AvatarURL returns a URL to the avatar image for the user.
// This loops over all auth providers to find the first available avatar.
// Currently only GitHub is supported. Returns blank string if no avatar URL available.
func (u *User) AvatarURL(size int) string {
	for _, auth := range u.Auths {
		if s := auth.AvatarURL(size); s != "" {
			return s
		}
	}
	return ""
}
