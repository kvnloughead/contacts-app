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
}

// ContactModel is a wrapper for our sql.DB connection pool.
// Contains methods for interacting with the Contacts collection.
type ContactModel struct {
	DB *sql.DB
}

type ContactModelInterface interface {
	Insert(first string, last string, phone string, email string) (int, error)
	Get(id int) (Contact, error)
	Latest() ([]Contact, error)
}

// Insert adds a new contact into the DB.
// Returns the ID of the inserted record or an error.
func (m *ContactModel) Insert(
	first string, last string, phone string, email string) (int, error) {

	// The query to be executed. Query statements allow for '?' as placeholders.
	query := `
		INSERT INTO contacts (first, last, phone, email, created)
		VALUES(?, ?, ?, ?, UTC_TIMESTAMP()`

	// Execute query. Exec accepts variadic values for the query placeholders.
	result, err := m.DB.Exec(query, first, last, phone, email)
	if err != nil {
		return 0, err
	}

	// Get ID of the inserted record as an int64.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// The Get method retrieves a contact by its ID.
// If no matching Contact is found, a models.ErrNoRecord error is returned.
func (m *ContactModel) Get(id int) (Contact, error) {
	query := `SELECT id, first, last, phone, email FROM contacts
	WHERE id = ?`

	// Executes a query statement that will return no more than one row.
	// Accepts the query statement and a variadic list of placeholder values.
	row := m.DB.QueryRow(query, id)

	// Declare an empty Contact and populate it from the row returned by QueryRow.
	// If no rows were found, an sql.ErrNoRows error is returned.
	// If multiple rows were found, the first row is used.
	var s Contact
	err := row.Scan(&s.ID, &s.First, &s.Last, &s.Phone, &s.Email, &s.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Contact{}, ErrNoRecord
		} else {
			return Contact{}, err
		}
	}

	return s, nil
}

func (m *ContactModel) Latest() ([]Contact, error) {
	query := `SELECT id, first, last, phone, email FROM contacts
	ORDER BY id DESC LIMIT 10`

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
		err = rows.Scan(&s.ID, &s.First, &s.Last, &s.Phone, &s.Email, &s.Created)
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