cd migrations &&
echo "Compile .SQL files..." && 
go run make_sql.go ../sql/personnes/ ../sql/dossiers/ ../sql/camps/ ../sql/files ../sql/inscriptions  ../sql/events ../sql/dons && 
echo "Packing..." &&
cat create_1_tables.sql > create_all.sql &&
cat create_2_json_funcs.sql >> create_all.sql && 
cat create_3_constraints.sql >> create_all.sql && 
cat init.sql >> create_all.sql && 
echo "Reset database <test>..." &&
dropdb test && 
createdb test && 
echo "Setup tables and add predeclared items..." && 
psql -v ON_ERROR_STOP=1 test < create_all.sql &&
echo "Done (created DB 'test')."