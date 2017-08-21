// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.
// 2017.08.19: Adapted from Bookshelf app - any errors are probably mine.

// Sample contacts is a fully-featured app demonstrating several Google Cloud APIs, including Datastore, Cloud SQL, Cloud Storage.
// See https://cloud.google.com/go/getting-started/tutorial-app
package main

import (
	_ "encoding/json"
	_ "errors"
	"fmt"
	_ "io"
	"log"
	"net/http"
	"os"
	_ "path"
	"strconv"

	//"cloud.google.com/go/pubsub"
	//"cloud.google.com/go/storage"

	_ "golang.org/x/net/context"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/satori/go.uuid"

	"google.golang.org/appengine"

	"github.com/rjj-work/yum-contacts"
)

var (
	// See template.go
	listTmpl   = parseTemplate("list.html")
	editTmpl   = parseTemplate("edit.html")
	detailTmpl = parseTemplate("detail.html")
)

func main() {
	registerHandlers()
	appengine.Main()
}

func registerHandlers() {
	// Use gorilla/mux for rich routing.
	// See http://www.gorillatoolkit.org/pkg/mux
	r := mux.NewRouter()

	r.Handle("/", http.RedirectHandler("/contacts", http.StatusFound))

	r.Methods("GET").Path("/contacts").
		Handler(appHandler(listHandler))
	r.Methods("GET").Path("/contacts/mine").
		Handler(appHandler(listMineHandler))
	r.Methods("GET").Path("/contacts/{id:[0-9]+}").
		Handler(appHandler(detailHandler))
	r.Methods("GET").Path("/contacts/add").
		Handler(appHandler(addFormHandler))
	r.Methods("GET").Path("/contacts/{id:[0-9]+}/edit").
		Handler(appHandler(editFormHandler))

	r.Methods("POST").Path("/contacts").
		Handler(appHandler(createHandler))
	r.Methods("POST", "PUT").Path("/contacts/{id:[0-9]+}").
		Handler(appHandler(updateHandler))
	r.Methods("POST").Path("/contacts/{id:[0-9]+}:delete").
		Handler(appHandler(deleteHandler)).Name("delete")

	// The following handlers are defined in auth.go and used in the
	// "Authenticating Users" part of the Getting Started guide.
	r.Methods("GET").Path("/login").
		Handler(appHandler(loginHandler))
	r.Methods("POST").Path("/logout").
		Handler(appHandler(logoutHandler))
	r.Methods("GET").Path("/oauth2callback").
		Handler(appHandler(oauthCallbackHandler))

	// Respond to App Engine and Compute Engine health checks.
	// Indicate the server is healthy.
	r.Methods("GET").Path("/_ah/health").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

	// Following handlers are defined in webhook.go
	// support for API.AI fulfillment
	r.Methods("POST").Path("/contactsWebhook").
		Handler( appHandler(webhookHandler) )

	// [START request_logging]
	// Delegate all of the HTTP routing and serving to the gorilla/mux router.
	// Log all requests using the standard Apache format.
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, r))
	// [END request_logging]
}

// listHandler displays a list with summaries of contacts in the database.
func listHandler(w http.ResponseWriter, r *http.Request) *appError {
	contacts, err := contacts.DB.ListContacts()
	if err != nil {
		return appErrorf(err, "could not list contacts: %v", err)
	}

	return listTmpl.Execute(w, r, contacts)
}

// listMineHandler displays a list of contacts created by the currently
// authenticated user.
func listMineHandler(w http.ResponseWriter, r *http.Request) *appError {
	user := profileFromSession(r)
	if user == nil {
		http.Redirect(w, r, "/login?redirect=/contacts/mine", http.StatusFound)
		return nil
	}

	contacts, err := contacts.DB.ListContactsCreatedBy(user.ID)
	if err != nil {
		return appErrorf(err, "could not list contacts: %v", err)
	}

	return listTmpl.Execute(w, r, contacts)
}

// contactFromRequest retrieves a contact from the database given a contact ID in the
// URL's path.
func contactFromRequest(r *http.Request) (*contacts.Contact, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad contact id: %v", err)
	}
	contact, err := contacts.DB.GetContact(id)
	if err != nil {
		return nil, fmt.Errorf("could not find contact: %v", err)
	}
	return contact, nil
}

// detailHandler displays the details of a given contact.
func detailHandler(w http.ResponseWriter, r *http.Request) *appError {
	contact, err := contactFromRequest(r)
	if err != nil {
		return appErrorf(err, "%v", err)
	}

	return detailTmpl.Execute(w, r, contact)
}

// addFormHandler displays a form that captures details of a new contact to add to
// the database.
func addFormHandler(w http.ResponseWriter, r *http.Request) *appError {
	return editTmpl.Execute(w, r, nil)
}

// editFormHandler displays a form that allows the user to edit the details of
// a given contact.
func editFormHandler(w http.ResponseWriter, r *http.Request) *appError {
	contact, err := contactFromRequest(r)
	if err != nil {
		return appErrorf(err, "%v", err)
	}

	return editTmpl.Execute(w, r, contact)
}

// contactFromForm populates the fields of a Contact from form values
// (see templates/edit.html).
func contactFromForm(r *http.Request) (*contacts.Contact, error) {
	/* imageURL, err := uploadFileFromForm(r)
	if err != nil {
		return nil, fmt.Errorf("could not upload file: %v", err)
	}
	if imageURL == "" {
		imageURL = r.FormValue("imageURL")
	} */

	contact := &contacts.Contact{
		FirstName:      r.FormValue("firstname"),
		LastName:       r.FormValue("lastname"),
		Address:        r.FormValue("address"),
		Email:          r.FormValue("email"),
		Phone:          r.FormValue("phone"),
		CreatedBy:      r.FormValue("createdBy"),
		CreatedByID:    r.FormValue("createdByID"),
	}

	// If the form didn't carry the user information for the creator, populate it
	// from the currently logged in user (or mark as anonymous).
	if contact.CreatedByID == "" {
		user := profileFromSession(r)
		if user != nil {
			// Logged in.
			contact.CreatedBy = user.DisplayName
			contact.CreatedByID = user.ID
		} else {
			// Not logged in.
			contact.SetCreatorAnonymous()
		}
	}

	return contact, nil
}

// createHandler adds a contact to the database.
func createHandler(w http.ResponseWriter, r *http.Request) *appError {
	contact, err := contactFromForm(r)
	if err != nil {
		return appErrorf(err, "could not parse contact from form: %v", err)
	}
	id, err := contacts.DB.AddContact(contact)
	if err != nil {
		return appErrorf(err, "could not save contact: %v", err)
	}
	//go publishUpdate(id)
	http.Redirect(w, r, fmt.Sprintf("/contacts/%d", id), http.StatusFound)
	return nil
}

// updateHandler updates the details of a given contact.
func updateHandler(w http.ResponseWriter, r *http.Request) *appError {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return appErrorf(err, "bad contact id: %v", err)
	}

	contact, err := contactFromForm(r)
	if err != nil {
		return appErrorf(err, "could not parse contact from form: %v", err)
	}
	contact.ID = id

	err = contacts.DB.UpdateContact(contact)
	if err != nil {
		return appErrorf(err, "could not update contact: %v", err)
	}
	//go publishUpdate(contact.ID)
	http.Redirect(w, r, fmt.Sprintf("/contacts/%d", contact.ID), http.StatusFound)
	return nil
}

// deleteHandler deletes a given contact.
func deleteHandler(w http.ResponseWriter, r *http.Request) *appError {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return appErrorf(err, "bad contact id: %v", err)
	}
	err = contacts.DB.DeleteContact(id)
	if err != nil {
		return appErrorf(err, "could not delete contact: %v", err)
	}
	http.Redirect(w, r, "/contacts", http.StatusFound)
	return nil
}


// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error)

		http.Error(w, e.Message, e.Code)
	}
}

func appErrorf(err error, format string, v ...interface{}) *appError {
	return &appError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}
