echo "Compile .SQL files..." && 
go run make_sql.go ../sql/personnes/ ../sql/dossiers/ ../sql/camps/ ../sql/files ../sql/dons && 
echo "Reset database <test>..." &&
dropdb test && 
createdb test && 
echo "Setup tables..." && 
psql test < create_1_tables.sql &&
psql test < create_2_json_funcs.sql &&
psql test < create_3_constraints.sql &&
echo "Add predeclared items..." &&
psql test < init.sql &&
echo "Done."