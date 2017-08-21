// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rjj-work/yum-contacts"
	"github.com/GoogleCloudPlatform/golang-samples/internal/testutil"
	"github.com/GoogleCloudPlatform/golang-samples/internal/webtest"
)

var wt *webtest.W

func TestMain(m *testing.M) {
	serv := httptest.NewServer(nil)
	wt = webtest.New(nil, serv.Listener.Addr().String())
	registerHandlers()

	os.Exit(m.Run())
}

func TestMainFunc(t *testing.T) {
	wt := webtest.New(t, "localhost:8080")
	m := testutil.BuildMain(t)
	defer m.Cleanup()
	m.Run(nil, func() {
		wt.WaitForNet()
		bodyContains(t, wt, "/", "No contacts found")
	})
}

func TestNoContacts(t *testing.T) {
	bodyContains(t, wt, "/", "No contacts found")
}

func TestContactDetail(t *testing.T) {
	const fn = "fn_contact"
	const ln = "ln_contact"
	id, err := contacts.DB.AddContact(&contacts.Contact{
		FirstName: fn,
		LastName: ln,
	})
	if err != nil {
		t.Fatal(err)
	}

	bodyContains(t, wt, "/", fn)

	contactPath := fmt.Sprintf("/contacts/%d", id)
	bodyContains(t, wt, contactPath, fn)

	if err := contacts.DB.DeleteContact(id); err != nil {
		t.Fatal(err)
	}

	bodyContains(t, wt, "/", "No contacts found")
}

func TestEditContact(t *testing.T) {
	const fn = "fn_contact"
	const ln = "ln_contact"
	id, err := contacts.DB.AddContact(&contacts.Contact{
		FirstName: fn,
		LastName: ln,
	})
	if err != nil {
		t.Fatal(err)
	}

	contactPath := fmt.Sprintf("/contacts/%d", id)
	editPath := contactPath + "/edit"
	bodyContains(t, wt, editPath, "Edit contact")
	bodyContains(t, wt, editPath, fn)

	var body bytes.Buffer
	m := multipart.NewWriter(&body)
	m.WriteField("firstname", "homer-edited")
	m.WriteField("lastname", "simpson-edited")
	m.WriteField("address", "742 Evergreen Terrace, Springfield-edited, HS")
	m.WriteField( "email", "homer.simpson-edited@simpsons.guru" )
	m.WriteField( "phone", "555-765-4321" )
	m.CreateFormFile("image", "")
	m.Close()

	resp, err := wt.Post(contactPath, "multipart/form-data; boundary="+m.Boundary(), &body)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := resp.Request.URL.Path, contactPath; got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	bodyContains(t, wt, contactPath, "simpson")
	bodyContains(t, wt, contactPath, "homer")

	if err := contacts.DB.DeleteContact(id); err != nil {
		t.Fatalf("got err %v, want nil", err)
	}
}

func TestAddAndDelete(t *testing.T) {
	bodyContains(t, wt, "/contacts/add", "Add contact")

	contactPath := fmt.Sprintf("/contacts")

	var body bytes.Buffer
	m := multipart.NewWriter(&body)
	m.WriteField("firstname", "homer")
	m.WriteField("lastname", "simpson")
	m.WriteField("address", "742 Evergreen Terrace, Springfield, HS")
	m.WriteField( "email", "homer.simpson@simpsons.guru" )
	m.WriteField( "phone", "555-123-4567" )
	m.CreateFormFile("image", "")
	m.Close()

	resp, err := wt.Post(contactPath, "multipart/form-data; boundary="+m.Boundary(), &body)
	if err != nil {
		t.Fatal(err)
	}

	gotPath := resp.Request.URL.Path
	if wantPrefix := "/contacts/"; !strings.HasPrefix(gotPath, wantPrefix) {
		t.Fatalf("redirect: got %q, want prefix %q", gotPath, wantPrefix)
	}

	bodyContains(t, wt, gotPath, "simpson")
	bodyContains(t, wt, gotPath, "homer")

	_, err = wt.Post(gotPath+":delete", "", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func bodyContains(t *testing.T, wt *webtest.W, path, contains string) (ok bool) {
	body, _, err := wt.GetBody(path)
	if err != nil {
		t.Error(err)
		return false
	}
	if !strings.Contains(body, contains) {
		t.Errorf("want %s to contain %s", body, contains)
		return false
	}
	return true
}
