package models

import "errors"

// Occurs when no matching record is found in the DB.
var ErrNoRecord = errors.New("models: no matching record found")

// Occurs when a user with the given email already exists.
var ErrDuplicateEmail = errors.New("models: duplicate email")

// Occurs when login credentials are invalid.
var ErrInvalidCredentials = errors.New("models: invalid credentials")
