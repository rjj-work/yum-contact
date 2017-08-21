// Adapted from code for: https://cloud.google.com/go/getting-started/tutorial-app

package contacts

// Contact definds the basic data collected for each entry.
type Contact struct {
	ID           int64
	FirstName    string
	LastName     string
	Address      string
	Email        string
	Phone        string
	CreatedBy    string
	CreatedByID  string
	CreatedDate  string
	LastEdited   string
}

// CreatedByDisplayName returns a string appropriate for displaying the name of
// the user who created this contact object.
func (b *Contact) CreatedByDisplayName() string {
	if b.CreatedByID == "anonymous" {
		return "Anonymous"
	}
	return b.CreatedBy
}

// SetCreatorAnonymous sets the CreatedByID field to the "anonymous" ID.
func (b *Contact) SetCreatorAnonymous() {
	b.CreatedBy = ""
	b.CreatedByID = "anonymous"
}

// ContactDatabase provides thread-safe access to a database of contacts.
type ContactDatabase interface {
	// ListContacts returns a list of contacts, ordered by title.
	ListContacts() ([]*Contact, error)

	// ListContactsCreatedBy returns a list of contacts, ordered by title, filtered by
	// the user who created the contact entry.
	ListContactsCreatedBy(userID string) ([]*Contact, error)

	// GetContact retrieves a contact by its ID.
	GetContact(id int64) (*Contact, error)

	// AddContact saves a given contact, assigning it a new ID.
	AddContact(b *Contact) (id int64, err error)

	// DeleteContact removes a given contact by its ID.
	DeleteContact(id int64) error

	// UpdateContact updates the entry for a given contact.
	UpdateContact(b *Contact) error

	// TallyContacts provides a count of contacts
	TallyContacts() (int64, error)

	// FindByNameContacts looks up contact by first and last name
	//	Note the name of function indicates we expect to only find 1 (or zero)
	FindContactByName(string, string) ([]*Contact, error)

	// Close closes the database, freeing up any available resources.
	// TODO(cbro): Close() should return an error.
	Close()
}
