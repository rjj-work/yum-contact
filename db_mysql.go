// Adapted from Contacts
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package contacts

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

var createTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS yum_contacts DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
	`USE yum_contacts;`,
	`CREATE TABLE IF NOT EXISTS contacts (
		id INT UNSIGNED NOT NULL AUTO_INCREMENT,
		firstName VARCHAR(255) NULL,
		lastName VARCHAR(255) NULL,
		address VARCHAR(255) NULL,
		email VARCHAR(255) NULL,
		phone TEXT NULL,
		createdBy VARCHAR(255) NULL,
		createdById VARCHAR(255) NULL,
		createdDate datetime DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id)
	)`,
}

// mysqlDB persists contacts to a MySQL instance.
type mysqlDB struct {
	conn *sql.DB

	list   *sql.Stmt
	listBy *sql.Stmt
	insert *sql.Stmt
	get    *sql.Stmt
	update *sql.Stmt
	delete *sql.Stmt
	// Added as part of the API.AI Fulfillement
	tally       *sql.Stmt
	findByName  *sql.Stmt
}

// Ensure mysqlDB conforms to the ContactDatabase interface.
var _ ContactDatabase = &mysqlDB{}

type MySQLConfig struct {
	// Optional.
	Username, Password string

	// Host of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Host string

	// Port of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Port int

	// UnixSocket is the filepath to a unix socket.
	//
	// If set, Host and Port should be unset.
	UnixSocket string
}

// dataStoreName returns a connection string suitable for sql.Open.
func (c MySQLConfig) dataStoreName(databaseName string) string {
	var cred string
	// [username[:password]@]
	if c.Username != "" {
		cred = c.Username
		if c.Password != "" {
			cred = cred + ":" + c.Password
		}
		cred = cred + "@"
	}

	if c.UnixSocket != "" {
		return fmt.Sprintf("%sunix(%s)/%s", cred, c.UnixSocket, databaseName)
	}
	return fmt.Sprintf("%stcp([%s]:%d)/%s", cred, c.Host, c.Port, databaseName)
}

// newMySQLDB creates a new ContactDatabase backed by a given MySQL server.
func newMySQLDB(config MySQLConfig) (ContactDatabase, error) {
	// Check database and table exists. If not, create it.
	if err := config.ensureTableExists(); err != nil {
		return nil, err
	}

	conn, err := sql.Open("mysql", config.dataStoreName("yum_contacts"))
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("mysql: could not establish a good connection: %v", err)
	}

	db := &mysqlDB{
		conn: conn,
	}

	// Prepared statements. The actual SQL queries are in the code near the
	// relevant method (e.g. addContact).
	if db.list, err = conn.Prepare(listStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare list: %v", err)
	}
	if db.listBy, err = conn.Prepare(listByStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare listBy: %v", err)
	}
	if db.get, err = conn.Prepare(getStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare get: %v", err)
	}
	if db.insert, err = conn.Prepare(insertStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare insert: %v", err)
	}
	if db.update, err = conn.Prepare(updateStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare update: %v", err)
	}
	if db.delete, err = conn.Prepare(deleteStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare delete: %v", err)
	}
	// Additions for APIAI
	if db.tally, err = conn.Prepare(tallyStatement); err != nil {
	return nil, fmt.Errorf("mysql: prepare tally: %v", err)
	}
	if db.findByName, err = conn.Prepare(findByNameStatement); err != nil {
	return nil, fmt.Errorf("mysql: prepare findByName: %v", err)
	}

	return db, nil
}

// Close closes the database, freeing up any resources.
func (db *mysqlDB) Close() {
	db.conn.Close()
}

// rowScanner is implemented by sql.Row and sql.Rows
type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanContact reads a contact from a sql.Row or sql.Rows
func scanContact(s rowScanner) (*Contact, error) {
	var (
		id          int64
		firstName   sql.NullString
		lastName    sql.NullString
		address     sql.NullString
		email       sql.NullString
		phone       sql.NullString
		createdBy   sql.NullString
		createdByID sql.NullString
		createdDate sql.NullString
	)
	if err := s.Scan(&id, &firstName, &lastName, &address, &email, &phone,
		// &imageURL,
		&createdBy, &createdByID, &createdDate); err != nil {
		return nil, err
	}

	contact := &Contact{
		ID:            id,
		FirstName:   firstName.String,
		LastName:    lastName.String,
		Address:     address.String,
		Email:       email.String,
		Phone:       phone.String,
		CreatedBy:   createdBy.String,
		CreatedByID: createdByID.String,
		CreatedDate: createdDate.String,
	}
	return contact, nil
}

const listStatement = `SELECT * FROM contacts ORDER BY lastname, firstname`

// ListContacts returns a list of contacts, ordered by name.
func (db *mysqlDB) ListContacts() ([]*Contact, error) {
	rows, err := db.list.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		contact, err := scanContact(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

const listByStatement = `
  SELECT * FROM contacts
  WHERE createdById = ? ORDER BY lastName, firstName`

// ListContactsCreatedBy returns a list of contacts, ordered by name, filtered by
// the user who created the contact entry.
func (db *mysqlDB) ListContactsCreatedBy(userID string) ([]*Contact, error) {
	if userID == "" {
		return db.ListContacts()
	}

	rows, err := db.listBy.Query(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		contact, err := scanContact(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

const getStatement = "SELECT * FROM contacts WHERE id = ?"

// GetContact retrieves a contact by its ID.
func (db *mysqlDB) GetContact(id int64) (*Contact, error) {
	contact, err := scanContact(db.get.QueryRow(id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("mysql: could not find contact with id %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get contact: %v", err)
	}
	return contact, nil
}

const insertStatement = `
  INSERT INTO contacts (
    firstName, lastName, address, email, phone, createdBy, createdById
  ) VALUES (?, ?, ?, ?, ?, ?, ?)`

// AddContact saves a given contact, assigning it a new ID.
func (db *mysqlDB) AddContact(b *Contact) (id int64, err error) {
	r, err := execAffectingOneRow(db.insert, b.FirstName, b.LastName, b.Address, b.Email, b.Phone,
		b.CreatedBy, b.CreatedByID)
	if err != nil {
		return 0, err
	}

	lastInsertID, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("mysql: could not get last insert ID: %v", err)
	}
	return lastInsertID, nil
}

const deleteStatement = `DELETE FROM contacts WHERE id = ?`

// DeleteContact removes a given contact by its ID.
func (db *mysqlDB) DeleteContact(id int64) error {
	if id == 0 {
		return errors.New("memorydb: contact with unassigned ID passed into deleteContact")
	}
	_, err := execAffectingOneRow(db.delete, id)
	return err
}

const updateStatement = `
  UPDATE contacts
  SET firstName=?, lastName=?, address=?, email=?, phone=?,
      createdBy=?, createdById=?
  WHERE id = ?`

// UpdateContact updates the entry for a given contact.
func (db *mysqlDB) UpdateContact(b *Contact) error {
	if b.ID == 0 {
		return errors.New("mysql: contact with unassigned ID passed into updateContact")
	}

	_, err := execAffectingOneRow(db.update, b.FirstName, b.LastName, b.Address, b.Email, b.Phone,
		b.CreatedBy, b.CreatedByID, b.ID)
	return err
}

// ensureTableExists checks the table exists. If not, it creates it.
func (config MySQLConfig) ensureTableExists() error {
	conn, err := sql.Open("mysql", config.dataStoreName(""))
	if err != nil {
		return fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	defer conn.Close()

	// Check the connection.
	if conn.Ping() == driver.ErrBadConn {
		return fmt.Errorf("mysql: could not connect to the database. " +
			"could be bad address, or this address is not whitelisted for access.")
	}

	if _, err := conn.Exec("USE yum_contacts"); err != nil {
		// MySQL error 1049 is "database does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1049 {
			return createTable(conn)
		}
	}

	if _, err := conn.Exec("DESCRIBE contacts"); err != nil {
		// MySQL error 1146 is "table does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1146 {
			return createTable(conn)
		}
		// Unknown error.
		return fmt.Errorf("mysql: could not connect to the database: %v", err)
	}
	return nil
}

// createTable creates the table, and if necessary, the database.
func createTable(conn *sql.DB) error {
	for _, stmt := range createTableStatements {
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// execAffectingOneRow executes a given statement, expecting one row to be affected.
func execAffectingOneRow(stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	r, err := stmt.Exec(args...)
	if err != nil {
		return r, fmt.Errorf("mysql: could not execute statement: %v", err)
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return r, fmt.Errorf("mysql: could not get rows affected: %v", err)
	} else if rowsAffected != 1 {
		return r, fmt.Errorf("mysql: expected 1 row affected, got %d", rowsAffected)
	}
	return r, nil
}


// 2017.08.21 rjj: Counting capability
const tallyStatement = `SELECT count(1) FROM contacts`

// TallyContacts returns the number of contacts, if we had owners of contacts, it should be in that context.
// Note if tally can not be determined, -1 is returned as tally value.
func (db *mysqlDB) TallyContacts() (int64, error) {

	tallyError := int64( -1 )	// Default value
	tally := tallyError

	rows, err := db.tally.Query()
	if err != nil {
		return tallyError, err
	}
	defer rows.Close()

	// Should be exactly 1 Row in *Rows
	if rows.Next() {
		if err := rows.Scan(&tally); err != nil {
			return tallyError, err
		}
		// tally should now have the correct value
	} else {
		// No rows found, return zero
		tally = 0
	}

	return tally, nil
}


// HACK: Forcing at most 1 contact to be found with this name.
const findByNameStatement = `
  SELECT * FROM contacts
  WHERE firstname = ? and lastname = ? LIMIT 1`

// ListContacts returns a list of contacts, ordered by name.
// There are several design choices to be made here:
//	- What if there is more than 1 contact with this name ?
//	- Should this procedure only return a single Contact, in which ase we do not need to return a slice
//	- Should we allow returning multiple contacts, but only take the first one, i.e. really return at most 1 ?
//		This is enforces a design decision at a very low level.
//		I don't think this is a good long term solution, but for this first cut I'm going to use it.
//	- At some point in the evolution of this it will be necessary to face this issue directly.
func (db *mysqlDB) FindContactByName( fn, ln string ) ([]*Contact, error) {
	rows, err := db.findByName.Query(fn, ln)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		contact, err := scanContact(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql::FindContactByName could not read row: %v", err)
		}

		// With the LIMIT 1, we should only have 1 row
		contacts = append(contacts, contact)
	}

	return contacts, nil
}
