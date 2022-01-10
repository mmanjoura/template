package user

type UserSerializer interface {
	DecodeUser(input []byte) (*User, error)
	EncodeUser(input *User) ([]byte, error)
}

type UserUpdateSerializer interface {
	DecodeUserUpdate(input []byte) (*UserUpdate, error)
	EncodeUserUpdate(input *UserUpdate) ([]byte, error)
}

type AuthSerializer interface {
	DecodeAuth(input []byte) (*Auth, error)
	EncodeAuth(input *Auth) ([]byte, error)
}
