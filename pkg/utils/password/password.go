package password

import "golang.org/x/crypto/bcrypt"

// Cost is the bcrypt cost for new hashes. 12 takes about 0.2s per check on a
// modern core (cost 14 took about 0.8s), which keeps logins responsive while
// staying well above OWASP's minimum of 10. Hashes made with another cost
// still verify; NeedsRehash reports them so they can be upgraded on login.
const Cost = 12

var CreatePassword = func(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		Cost,
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

// NeedsRehash reports whether a stored hash uses a different cost than Cost,
// so it should be re-created after a successful login. Unparseable hashes
// return false: they can't be verified anyway.
var NeedsRehash = func(hashedPassword string) bool {
	cost, err := bcrypt.Cost([]byte(hashedPassword))
	if err != nil {
		return false
	}

	return cost != Cost
}
