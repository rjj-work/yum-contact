#!/bin/bash

curl -v \
  -H 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8' \
  -H 'Content-Type: application/json' \
  -d@curl-webhook-number_of_contacts.json \
  http://localhost:8080/contactsWebhook


