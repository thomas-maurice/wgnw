#!/bin/bash

./bin/wgnw-server -sql-driver postgres -sql-string "host=127.0.0.1 user=root port=26257 dbname=postgres sslmode=disable" -debug
