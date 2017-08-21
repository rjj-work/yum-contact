# Introduction
This work was started as a response to a Golang challenge test for a possible employment opportunity.

 2017.08.19 rjj: Shamely use of https://cloud.google.com/go/getting-started/tutorial-app
The bookshelf app illustrated many of the necessary points to deal with GCP features, e.g. Cloud SQL
so I've borrowed from that example.

This solution should support a simple web interface, similar to the bookshelf app.
But also include support for POST messages from Google Actions

## Setup
* Cloud SQL instance:
	* My testing is using: rjj-work-testing:us-east1:rjj-work-mysql-01
	* You would probably want to connect to a different Cloud SQL instance
	* Modify app/app.yaml to point to *your* Cloud SQL instance
	* Modify config.go to supply the SQL root password for *your* Cloud SQL instance


## Local Testing
Note that I had a local MySQL DB instance running on the development machine.
As such when running the SQL proxy, I had a port conflict for 3306 (default MySQL port)
So made these changes:
* SQL proxy port 13306
	* CMD: ./cloud_sql_proxy -instances=rjj-work-testing:us-east1:rjj-work-mysql-01=tcp:13306
* config.go changes
	* added Port to type: cloudSQLConfig
	* func init() now includes hard-coded 'Port: 13306' argument
	* func configureCloudSQL(), for local connection now references config.Port


## Webhook for Filfullment

This URL provided an initial clue for how to handle the Webhook
	https://github.com/rominirani/api-ai-workshop/blob/master/webhook%20implementations/golang/server.go

### Webhook implementation
* contact.go
	* ContactDatabase.TallyContacts() added for intent "number_of_contacts"
* db_mysql.go
	* type mysqlDB struct: field "tally" added
	* newMySQLDB: tallStatement added as prepared SQL stmt
	* const tallyStatement: added to do query - note that per user contacts not supported (yet)
	* func (db *mysqlDB) TallyContacts() (int64, error): added to implement new tally behavior.
* app/webhook.go: new file for implementation of webhook handler


## Manual Testing via curl
* Start "cloud_sql_proxy" as noted above
* Start the app locally
```go
cd app
go run app.go auth.go template.go webhook.go
```
	* Once the app is started the local web interface can be used to both inspect and modify data
	http://localhost:8080/contacts
* Execute a curl statement
	* For INTENT number_of_contacts
```bash
cd ../manual-testing
./curl-webhook-number_of_contacts.sh
```
	* For INTENT find_contact
```bash
cd ../manual-testing
./curl-webhook-find_contact.sh
```
