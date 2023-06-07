package session

import (
	"context"
	"errors"
	"time"

	"github.com/FayeZheng0/ask_pubmed/config"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/golang-jwt/jwt"
	"github.com/patrickmn/go-cache"
)

var (
	ErrKeystoreNotProvided = errors.New("keystore not provided, use --file or --stdin")
	ErrPinNotProvided      = errors.New("pin not provided, use --pin or include in keystore file")
)

type Session struct {
	Version string

	store *mixin.Keystore
	token string
	pin   string

	cache *cache.Cache
}

func (s *Session) WithKeystore(store *mixin.Keystore) *Session {
	s.store = store
	return s
}

func (s *Session) WithAccessToken(token string) *Session {
	s.token = token
	return s
}

func (s *Session) WithPin(pin string) *Session {
	s.pin = pin
	return s
}

func (s *Session) GetKeystore() (*mixin.Keystore, error) {
	if s.store != nil {
		return s.store, nil
	}

	return nil, ErrKeystoreNotProvided
}

func (s *Session) GetPin() (string, error) {
	if s.pin != "" {
		return s.pin, nil
	}

	return "", ErrPinNotProvided
}

func (s *Session) GetClient() (*mixin.Client, error) {
	store, err := s.GetKeystore()
	if err != nil {
		return mixin.NewFromAccessToken(s.token), nil
	}

	return mixin.NewFromKeystore(store)
}

func (s *Session) Login(ctx context.Context, accessToken string, refresh bool) (*mixin.User, error) {
	var claim struct {
		jwt.StandardClaims
		Scope string `json:"scp,omitempty"`
	}

	jwt.ParseWithClaims(accessToken, &claim, nil)
	// @TODO: in some cases, the Valid() will return `Token used before issued` error.
	// I have no idea about it.
	if err := claim.Valid(); err != nil {
		return nil, err
	}

	if s.cache == nil {
		s.cache = cache.New(time.Hour, time.Hour)
	}

	uID, found := s.cache.Get(accessToken)
	if found && !refresh {
		u, found := s.cache.Get(uID.(string))
		if found {
			return u.(*mixin.User), nil
		}
	}

	cli := mixin.NewFromAccessToken(accessToken)
	user, err := cli.UserMe(ctx)
	if err != nil {
		return nil, err
	}

	isAdmin := false
	for _, adminID := range config.C().Admins {
		if user.UserID == adminID {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return nil, errors.New("not admin")
	}

	s.cache.Set(user.UserID, user, time.Hour)
	s.cache.Set(accessToken, user.UserID, time.Hour)

	return user, nil
}
