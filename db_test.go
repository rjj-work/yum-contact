// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package contacts

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"cloud.google.com/go/datastore"

	"golang.org/x/net/context"

	"github.com/GoogleCloudPlatform/golang-samples/internal/testutil"
)

func testDB(t *testing.T, db ContactDatabase) {
	defer db.Close()

	b := &Contact{
		Address:      "testy mc testface",
		FirstName LastName:       fmt.Sprintf("t-%d", time.Now().Unix()),
		Phone: "desc",
	}

	id, err := db.AddContact(b)
	if err != nil {
		t.Fatal(err)
	}

	b.ID = id
	b.Phone = "newdesc"
	if err := db.UpdateContact(b); err != nil {
		t.Error(err)
	}

	gotContact, err := db.GetContact(id)
	if err != nil {
		t.Error(err)
	}
	if got, want := gotContact.Phone, b.Phone; got != want {
		t.Errorf("Update phone: got %q, want %q", got, want)
	}

	if err := db.DeleteContact(id); err != nil {
		t.Error(err)
	}

	if _, err := db.GetContact(id); err == nil {
		t.Error("want non-nil err")
	}
}

func TestMemoryDB(t *testing.T) {
	testDB(t, newMemoryDB())
}

func TestDatastoreDB(t *testing.T) {
	tc := testutil.SystemTest(t)
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, tc.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	db, err := newDatastoreDB(client)
	if err != nil {
		t.Fatal(err)
	}
	testDB(t, db)
}

func TestMySQLDB(t *testing.T) {
	t.Parallel()

	host := os.Getenv("GOLANG_SAMPLES_MYSQL_HOST")
	port := os.Getenv("GOLANG_SAMPLES_MYSQL_PORT")

	if host == "" {
		t.Skip("GOLANG_SAMPLES_MYSQL_HOST not set.")
	}
	if port == "" {
		port = "13306"
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("Could not parse port: %v", err)
	}

	db, err := newMySQLDB(MySQLConfig{
		Username: "root",
		Host:     host,
		Port:     p,
	})
	if err != nil {
		t.Fatal(err)
	}
	testDB(t, db)
}
