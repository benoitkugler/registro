# assume build_db.sh has already been run
echo "Reset database <test>..." &&
dropdb --if-exists test && 
createdb test && 
echo "Setup tables and add predeclared items..." && 
psql -v ON_ERROR_STOP=1 test < ../create_all.sql &&
echo "Launching migration script..." &&
source env.sh &&
go run *.go