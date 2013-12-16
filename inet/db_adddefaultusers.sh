#!/bin/bash

db="db.sqlite3"

# Add users

sqlite3 $db 'INSERT INTO users VALUES (1, "0001", 0000, 1000);'
sqlite3 $db 'INSERT INTO users VALUES (2, "1234432143211234", 4321, 1000);'
sqlite3 $db 'INSERT INTO users VALUES (3, "1000200030004000", 1234, 1000);'

# Add codes (odd numbers 1..99)

for ((user = 1; user <= 3; user++)); do
	for ((i = 1; i <= 99; i=i+2)); do
	    sqlite3 $db 'INSERT INTO codes VALUES (NULL, '$i', '$user');'
	done
done