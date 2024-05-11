package models

import (
	"database/sql"
	"errors"
	"time"
)

// Contact is a struct representing a contact document.
type Contact struct {
	ID      int
	First   string
	Last    string
	Phone   string
	Email   string
	Created time.Time
	Version int32
}

// ContactModel is a wrapper for our sql.DB connection pool.
// Contains methods for interacting with the Contacts collection.
type ContactModel struct {
	DB *sql.DB
}

type ContactModelInterface interface {
	Insert(first string, last string, phone string, email string) (int, error)
	Get(id int) (Contact, error)
	GetAll() ([]Contact, error)
	Update(contact *Contact) error
}

// Insert adds a new contact into the DB.
// Returns the ID of the inserted record or an error.
func (m *ContactModel) Insert(
	first string, last string, phone string, email string) (int, error) {
	query := `
		INSERT INTO contacts (first, last, phone, email, created)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		RETURNING id;`

	var lastInsertId int
	err := m.DB.QueryRow(query, first, last, phone, email).Scan(&lastInsertId)
	if err != nil {
		return 0, err
	}

	return int(lastInsertId), nil
}

// The Get method retrieves a contact by its ID.
// If no matching Contact is found, a models.ErrNoRecord error is returned.
func (m *ContactModel) Get(id int) (Contact, error) {
	query := `SELECT id, first, last, phone, email, version FROM contacts
	WHERE id = $1`

	// Executes a query statement that will return no more than one row.
	// Accepts the query statement and a variadic list of placeholder values.
	row := m.DB.QueryRow(query, id)

	// Declare an empty Contact and populate it from the row returned by QueryRow.
	// If no rows were found, an sql.ErrNoRows error is returned.
	// If multiple rows were found, the first row is used.
	var s Contact
	err := row.Scan(&s.ID, &s.First, &s.Last, &s.Phone, &s.Email, &s.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Contact{}, ErrNoRecord
		} else {
			return Contact{}, err
		}
	}

	return s, nil
}

// Update updates a specific record in the contacts table. The caller should
// check for the existence of the record to be updated before calling Update.
// The record's version field is incremented by 1 after update.
//
// Prevents edit conflicts by verifying that the version of the record in the
// UPDATE query is the same as the version of the contact argument. In case of
// an edit conflict, an ErrEditConflict error is returned.
func (m *ContactModel) Update(contact *Contact) error {
	query := `
		UPDATE contacts
		SET first = $1, last = $2, phone = $3, email = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args := []any{contact.First, contact.Last, contact.Phone, contact.Email, contact.ID, contact.Version}

	err := m.DB.QueryRow(query, args...).Scan(&contact.Version)
	if err != nil {
		switch {
		// An sql.ErrNoRows is returned if there are no matching records. Since we
		// know that the record exists already, this can be assumed to be due to a
		// version mismatch (hence an edit conflict).
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// GetAll retrieves all contacts from the DB.
func (m *ContactModel) GetAll() ([]Contact, error) {
	query := `SELECT id, first, last, phone, email FROM contacts
	ORDER BY first`

	// Query will return an sql.Rows result set containing 10 latest entries.
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // don't defer closing until after handling the error

	// Iterate through result set, calling rows.Scan on each row. Create a contact
	// Create a contact for each row and add it to the contacts slice.
	var contacts []Contact

	for rows.Next() {
		var s Contact
		err = rows.Scan(&s.ID, &s.First, &s.Last, &s.Phone, &s.Email)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, s)
	}

	// rows.Err() contains any errors that occurred during iteration, including
	// including errors that wouldn't be returned by rows.Scan().
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return contacts, nil
}
