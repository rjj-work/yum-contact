// 2017.08.20 rjj.work@gmail.com: Provide support for API.AI Fulfillment
//

package main

import(
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rjj-work/yum-contacts"
)

// =====================================================================================================
// Credit for APIAIRequest and APIAIMessage data structure to:
//	https://github.com/rominirani/api-ai-workshop/blob/master/webhook%20implementations/golang/server.go

//APIAIRequest : Incoming request format from APIAI
type APIAIRequest struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Result    struct {
		Parameters map[string]string `json:"parameters"`
		Contexts   []interface{}     `json:"contexts"`
		Metadata   struct {
			IntentID                  string `json:"intentId"`
			WebhookUsed               string `json:"webhookUsed"`
			WebhookForSlotFillingUsed string `json:"webhookForSlotFillingUsed"`
			IntentName                string `json:"intentName"`
		} `json:"metadata"`
		Score float32 `json:"score"`
	} `json:"result"`
	Status struct {
		Code      int    `json:"code"`
		ErrorType string `json:"errorType"`
	} `json:"status"`
	SessionID       string      `json:"sessionId"`
	OriginalRequest interface{} `json:"originalRequest"`
}

//APIAIMessage : Response Message Structure
type APIAIMessage struct {
	Speech      string `json:"speech"`
	DisplayText string `json:"displayText"`
	Source      string `json:"source"`
}
// =====================================================================================================

// Need to think about how to pass in ID from DB
type APIAIContact struct {
	GivenName string
	LastName  string
	Address   string
	Email     string
	Phone     string
}


// This handler is to be used by the filfullment for the my_contacts api.aid
// Basic flow
//	- hook invoked, verify POST
//	- extract json request
//	- determine what is being requested, either by INTENT or the action
//	- pass processing off to func per request type
func webhookHandler(w http.ResponseWriter, r *http.Request) *appError {
	// Verify POST
	if "POST" != r.Method  {
		http.Error( w, "Error, expected POST, received: " + r.Method, http.StatusMethodNotAllowed )
	}

	// OK, so if here it is a POST method
	// Extract the body which should be json data
	decoder := json.NewDecoder(r.Body)

	var ar APIAIRequest
	err := decoder.Decode(&ar)
	if nil != err {
		return appErrorf( err, "Decode of request failed: %v", err )
	}

	log.Printf( "Contact Params: %+v", extractContactFromAPIAIRequest(&ar) )

	// Extract the INTENT, not sure where Action is at or if included
	// 	Use INTENT to process correctly
	respJson := APIAIMessage{
		Speech: "Unprocessed Speech value",
		DisplayText: "Unprocessed DisplayText value",
		Source: "rjj-work@gmail.com yum-contacts programming exercise",
		}

	intent := ar.Result.Metadata.IntentName
	switch intent {
		case "number_of_contacts" : err = tallyContacts( &ar, &respJson )
		case "find_contact"       : err = findContact( &ar, &respJson )
		case "add_contact"        : err = addContact( &ar, &respJson )
		case "update_contact"     : err = updateContact( &ar, &respJson )
		case "delete_contact"     : err = deleteContact( &ar, &respJson )
		default                   : err = unhandledIntent( &ar, &respJson )
	}

	// Hopefully no errors, but check anyway
	if nil != err {
		return appErrorf( err, "Processing of INTENT: %s failed: %v", intent, err )
	}

	// Encode the response and done
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode( respJson )

	return nil
}

func tallyContacts( ar *APIAIRequest, rj *APIAIMessage ) error {
	// Hit the DB and get the count
	tally, err := contacts.DB.TallyContacts()
	if nil != err {
		rj.Speech = fmt.Sprintf( "Error: tallying contacts, %v", err )
		rj.DisplayText = rj.Speech
		return err
	}
	// Should have an actual count
	rj.Speech = fmt.Sprintf( "Tally %d for contacts as of %v", tally, time.Now() )
	rj.DisplayText = rj.Speech

	return err
}

// A more sophisticated implementation would supprt a find using a combination of contract attributes
//	For now we will just use first and last name
func findContact( ar *APIAIRequest, rj *APIAIMessage ) error {
	var err error
	var cts []*contacts.Contact
	t := extractContactFromAPIAIRequest( ar )

	cts, err = contacts.DB.FindContactByName( t.GivenName, t.LastName )
	if nil != err {
		rj.Speech = fmt.Sprintf( "Error looking up contact %s %s, %v", t.GivenName, t.LastName, err )
		rj.DisplayText = rj.Speech
		return err
	}
	// HACK, we will just use the first contact found.
	if 0 == len(cts) {
		// No contacts found
		rj.Speech =  fmt.Sprintf( "No contact found for first name %s, last name: %s", t.GivenName, t.LastName )
		rj.DisplayText = rj.Speech
		return nil
	}
	fn := cts[0].FirstName
	ln := cts[0].LastName
	address := cts[0].Address
	email := cts[0].Email
	phone := cts[0].Phone

	// Assume all there for now
	rj.Speech =  fmt.Sprintf( "Found: %s %s at address: %s, with phone number: %s and email: %s",
			fn, ln, address, phone, email )
	rj.DisplayText = rj.Speech

	// TODO:
	//	- [ ] Set outgoing context with atleast the current contacts DB ID
	//	- [ ] Allow for multiple contats found, and some indication of that condition
	return err
}

func addContact( ar *APIAIRequest, rj *APIAIMessage ) error {
	var err error
	numContacts := 0

	// Do something to get this value

	rj.Speech = fmt.Sprintf( "You have %d contacts as of %v", numContacts, time.Now() )
	rj.DisplayText = rj.Speech

	return err
}

func updateContact( ar *APIAIRequest, rj *APIAIMessage ) error {
	var err error
	numContacts := 0

	// Do something to get this value

	rj.Speech = fmt.Sprintf( "You have %d contacts as of %v", numContacts, time.Now() )
	rj.DisplayText = rj.Speech

	return err
}

func deleteContact( ar *APIAIRequest, rj *APIAIMessage ) error {
	var err error
	numContacts := 0

	// Do something to get this value

	rj.Speech = fmt.Sprintf( "You have %d contacts as of %v", numContacts, time.Now() )
	rj.DisplayText = rj.Speech

	return err
}

func unhandledIntent( ar *APIAIRequest, rj *APIAIMessage ) error {
	var err error
	numContacts := 0

	// Do something to get this value

	rj.Speech = fmt.Sprintf( "You have %d contacts as of %v", numContacts, time.Now() )
	rj.DisplayText = rj.Speech

	return err
}

func extractContactFromAPIAIRequest( ar *APIAIRequest ) *APIAIContact {
	if nil == ar {
		// HACK: this a) should never happen, and b) should probably throw and exception if it does
		appErrorf( nil, "Missing APIAIRequest", ar )
		return nil
	}

	return &APIAIContact{
		GivenName: ar.Result.Parameters["given-name"],
		LastName: ar.Result.Parameters["last-name"],
		Address: ar.Result.Parameters["address"],
		Email: ar.Result.Parameters["email"],
		Phone: ar.Result.Parameters["phone-number"],
	}
}
