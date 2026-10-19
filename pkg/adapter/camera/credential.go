package camera

import (
	"camera_api/pkg/utils/crypto"

	"github.com/andriantp/camera"
)

func (c *icamera) EncryptCredential(credential camera.Credential) (*camera.Credential, error) {
	host, err := crypto.Encrypt(credential.Host, c.key)
	if err != nil {
		return nil, err
	}

	username, err := crypto.Encrypt(credential.Username, c.key)
	if err != nil {
		return nil, err
	}

	password, err := crypto.Encrypt(credential.Password, c.key)
	if err != nil {
		return nil, err
	}

	return &camera.Credential{
		Host:     host,
		Username: username,
		Password: password,
	}, nil
}

func (c *icamera) DecryptCredential(credential camera.Credential) (*camera.Credential, error) {
	host, err := crypto.Decrypt(credential.Host, c.key)
	if err != nil {
		return nil, err
	}

	username, err := crypto.Decrypt(credential.Username, c.key)
	if err != nil {
		return nil, err
	}

	password, err := crypto.Decrypt(credential.Password, c.key)
	if err != nil {
		return nil, err
	}

	return &camera.Credential{
		Host:     host,
		Username: username,
		Password: password,
	}, nil
}

func (c *icamera) Encrypt(plainText string) (string, error) {
	return crypto.Encrypt(plainText, c.key)
}
