package camera

import (
	"github.com/andriantp/camera"
)

func (c *icamera) EncryptCredential(credential camera.Credential) (*camera.Credential, error) {
	host, err := c.cipher.Encrypt(credential.Host)
	if err != nil {
		return nil, err
	}

	username, err := c.cipher.Encrypt(credential.Username)
	if err != nil {
		return nil, err
	}

	password, err := c.cipher.Encrypt(credential.Password)
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
	host, err := c.cipher.Decrypt(credential.Host)
	if err != nil {
		return nil, err
	}

	username, err := c.cipher.Decrypt(credential.Username)
	if err != nil {
		return nil, err
	}

	password, err := c.cipher.Decrypt(credential.Password)
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
	return c.cipher.Encrypt(plainText)
}
