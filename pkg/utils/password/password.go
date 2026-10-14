package password

import "golang.org/x/crypto/bcrypt"

var CreatePassword = func(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		14,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

var CheckPassword = func(password string, hashedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
	if err != nil {
		return false, err
	}

	return true, nil
}
