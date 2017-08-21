// Adapted from Contacts
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package contacts

import (
	_ "errors"
	"log"
	"os"

	"gopkg.in/mgo.v2"

	"github.com/gorilla/sessions"

	_ "golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	DB          ContactDatabase
	OAuthConfig *oauth2.Config

	//StorageBucket     *storage.BucketHandle
	//StorageBucketName string

	SessionStore sessions.Store

	//PubsubClient *pubsub.Client

	// Force import of mgo yum_contacts.
	_ mgo.Session
)

const PubsubTopicID = "fill-contact-details"

func init() {
	var err error

	// [START cloudsql]
	// To use Cloud SQL, uncomment the following lines, and update the username,
	// password and instance connection string. When running locally,
	// localhost:3306 is used, and the instance name is ignored.
	DB, err = configureCloudSQL(cloudSQLConfig{
		Username: "root",
		Password: "<YOUR-Cloud-SQL-root-password>",
		// The connection name of the Cloud SQL v2 instance, i.e.,
		// "project:region:instance-id"
		// Cloud SQL v1 instances are not supported.
		Instance: "rjj-work-testing:us-east1:rjj-work-mysql-01",
		Port: 13306,
	})
	// [END cloudsql]


	if err != nil {
		log.Fatal(err)
	}

	// [START auth]
	// To enable user sign-in, uncomment the following lines and update the
	// Client ID and Client Secret.
	// You will also need to update OAUTH2_CALLBACK in app.yaml when pushing to
	// production.
	//
	// OAuthConfig = configureOAuthClient("clientid", "clientsecret")
	// [END auth]

	// [START sessions]
	// Configure storage method for session-wide information.
	// Update "something-very-secret" with a hard to guess string or byte sequence.
	cookieStore := sessions.NewCookieStore([]byte("something-very-secret"))
	cookieStore.Options = &sessions.Options{
		HttpOnly: true,
	}
	SessionStore = cookieStore
	// [END sessions]

	if err != nil {
		log.Fatal(err)
	}
}


func configureOAuthClient(clientID, clientSecret string) *oauth2.Config {
	redirectURL := os.Getenv("OAUTH2_CALLBACK")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/oauth2callback"
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

type cloudSQLConfig struct {
	Username, Password, Instance string
	Port int
}

func configureCloudSQL(config cloudSQLConfig) (ContactDatabase, error) {
	if os.Getenv("GAE_INSTANCE") != "" {
		// Running in production.
		return newMySQLDB(MySQLConfig{
			Username:   config.Username,
			Password:   config.Password,
			UnixSocket: "/cloudsql/" + config.Instance,
		})
	}

	// Running locally.
	return newMySQLDB(MySQLConfig{
		Username: config.Username,
		Password: config.Password,
		Host:     "localhost",
		// 3306 conflicts with local MySQL instance, so use different port for proxy
		// Port:     3306,
		Port:     config.Port,
	})
}
