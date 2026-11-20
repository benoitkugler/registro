#!/usr/bin/env bash
# Required environment variables:
#   DB_HOST, DB_USER, DB_PASSWORD, DB_NAME (DB_PORT is optional)
#   COMPTA_GOOGLE_DRIVE_KEY_FILE, COMPTA_GOOGLE_DRIVE_FOLDER_ID
#   COMPTA_FILES_DIR, COMPTA_ADMIN_PASSWORD, COMPTA_ADMIN_MAIL
source ../../env.sh &&
go run main.go "$@"
