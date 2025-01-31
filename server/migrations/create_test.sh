go run make_sql.go ../sql/personnes/ ../sql/dossiers/ ../sql/camps/ && 
dropdb test && 
createdb test && 
psql test < create_1_tables.sql &&
psql test < create_2_json_funcs.sql &&
psql test < create_3_constraints.sql &&
echo "Done."