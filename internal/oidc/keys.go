package oidc

import (
	"context"
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	fositejwt "github.com/ory/fosite/token/jwt"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// NewKeyManager when provided a schema.OpenIDConnectConfiguration creates a new KeyManager and initializes the Strategy
// for use with Fosite.
func NewKeyManager(configuration *schema.OpenIDConnectConfiguration) (manager *KeyManager, err error) {
	manager = new(KeyManager)
	manager.keys = map[string]*rsa.PrivateKey{}
	manager.keySet = new(jose.JSONWebKeySet)

	key, webKey, err := manager.AddActiveKeyData(configuration.IssuerPrivateKey)
	if err != nil {
		return nil, err
	}

	manager.strategy, err = NewRS256JWTStrategy(webKey.KeyID, key)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// KeyManager keeps track of all of the active/inactive rsa keys and provides them to services requiring them.
// It additionally allows us to add keys for the purpose of key rotation in the future.
type KeyManager struct {
	activeKeyID string
	keys        map[string]*rsa.PrivateKey
	keySet      *jose.JSONWebKeySet
	strategy    *RS256JWTStrategy
}

// Strategy returns the RS256JWTStrategy.
func (m KeyManager) Strategy() (strategy *RS256JWTStrategy) {
	return m.strategy
}

// GetKeySet returns the joseJSONWebKeySet containing the rsa.PublicKey types.
func (m KeyManager) GetKeySet() (keySet *jose.JSONWebKeySet) {
	return m.keySet
}

// GetActiveWebKey obtains the currently active jose.JSONWebKey.
func (m KeyManager) GetActiveWebKey() (webKey *jose.JSONWebKey, err error) {
	webKeys := m.keySet.Key(m.activeKeyID)
	if len(webKeys) == 1 {
		return &webKeys[0], nil
	}

	if len(webKeys) == 0 {
		return nil, errors.New("could not find a key with the active key id")
	}

	return &webKeys[0], errors.New("multiple keys with the same key id")
}

// GetActiveKeyID returns the key id of the currently active key.
func (m KeyManager) GetActiveKeyID() (keyID string) {
	return m.activeKeyID
}

// GetActiveKey returns the rsa.PublicKey of the currently active key.
func (m KeyManager) GetActiveKey() (key *rsa.PublicKey, err error) {
	if key, ok := m.keys[m.activeKeyID]; ok {
		return &key.PublicKey, nil
	}

	return nil, errors.New("failed to retrieve active key")
}

// GetActivePrivateKey returns the rsa.PrivateKey of the currently active key.
func (m KeyManager) GetActivePrivateKey() (key *rsa.PrivateKey, err error) {
	if key, ok := m.keys[m.activeKeyID]; ok {
		return key, nil
	}

	return nil, errors.New("failed to retrieve active key")
}

// AddActiveKeyData adds a rsa.PublicKey given the key in the PEM string format, then sets it to the active key.
func (m *KeyManager) AddActiveKeyData(data string) (key *rsa.PrivateKey, webKey *jose.JSONWebKey, err error) {
	key, err = utils.ParseRsaPrivateKeyFromPemStr(data)
	if err != nil {
		return nil, nil, err
	}

	webKey, err = m.AddActiveKey(key)

	return key, webKey, err
}

// AddActiveKey adds a rsa.PublicKey, then sets it to the active key.
func (m *KeyManager) AddActiveKey(key *rsa.PrivateKey) (webKey *jose.JSONWebKey, err error) {
	wk := jose.JSONWebKey{
		Key:       &key.PublicKey,
		Algorithm: "RS256",
		Use:       "sig",
	}

	keyID, err := wk.Thumbprint(crypto.SHA256)
	if err != nil {
		return nil, err
	}

	strKeyID := string(keyID)

	if _, ok := m.keys[strKeyID]; ok {
		return nil, fmt.Errorf("key id %s already exists", strKeyID)
	}

	// TODO: Add Mutex here when implementing key rotation.
	wk.KeyID = strKeyID
	m.keySet.Keys = append(m.keySet.Keys, wk)
	m.keys[strKeyID] = key

	return &wk, nil
}

// NewRS256JWTStrategy returns a new RS256JWTStrategy.
func NewRS256JWTStrategy(id string, key *rsa.PrivateKey) (strategy *RS256JWTStrategy, err error) {
	strategy = new(RS256JWTStrategy)
	strategy.JWTStrategy = new(fositejwt.RS256JWTStrategy)

	strategy.SetKey(id, key)

	return strategy, nil
}

// RS256JWTStrategy is a decorator struct for the fosite RS256JWTStrategy.
type RS256JWTStrategy struct {
	JWTStrategy *fositejwt.RS256JWTStrategy

	keyID string
}

// KeyID returns the key id.
func (s RS256JWTStrategy) KeyID() (id string) {
	return s.keyID
}

// SetKey sets the provided key id and key as the active key (this is what triggers fosite to use it).
func (s *RS256JWTStrategy) SetKey(id string, key *rsa.PrivateKey) {
	s.keyID = id
	s.JWTStrategy.PrivateKey = key
}

// Hash is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) Hash(ctx context.Context, in []byte) ([]byte, error) {
	return s.JWTStrategy.Hash(ctx, in)
}

// GetSigningMethodLength is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) GetSigningMethodLength() int {
	return s.JWTStrategy.GetSigningMethodLength()
}

// GetSignature is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) GetSignature(ctx context.Context, token string) (string, error) {
	return s.JWTStrategy.GetSignature(ctx, token)
}

// Generate is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) Generate(ctx context.Context, claims jwt.Claims, header fositejwt.Mapper) (string, string, error) {
	return s.JWTStrategy.Generate(ctx, claims, header)
}

// Validate is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) Validate(ctx context.Context, token string) (string, error) {
	return s.JWTStrategy.Validate(ctx, token)
}

// Decode is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) Decode(ctx context.Context, token string) (*jwt.Token, error) {
	return s.JWTStrategy.Decode(ctx, token)
}

// GetPublicKeyID is a decorator func for the underlying fosite RS256JWTStrategy.
func (s *RS256JWTStrategy) GetPublicKeyID(_ context.Context) (string, error) {
	return s.keyID, nil
}