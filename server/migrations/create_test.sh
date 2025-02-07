cd migrations &&
echo "Compile .SQL files..." && 
go run make_sql.go ../sql/personnes/ ../sql/dossiers/ ../sql/camps/ ../sql/files ../sql/inscriptions  ../sql/dons && 
echo "Reset database <test>..." &&
dropdb test && 
createdb test && 
echo "Setup tables..." && 
psql -v ON_ERROR_STOP=1 test < create_1_tables.sql &&
psql -v ON_ERROR_STOP=1 test < create_2_json_funcs.sql &&
psql -v ON_ERROR_STOP=1 test < create_3_constraints.sql &&
echo "Add predeclared items..." &&
psql -v ON_ERROR_STOP=1 test < init.sql &&
echo "Done."