# assume build_db.sh has already been run
for arg in "$@"; do
  case "$arg" in
  -commit | --commit)
    echo "Reset database <test>..." &&
      dropdb --if-exists test &&
      createdb test &&
      echo "Setup tables and add predeclared items..." &&
      psql -v ON_ERROR_STOP=1 test <../create_all.sql ||
      exit 1
    break
    ;;
  esac
done

echo "Launching migration..." &&
  source env.sh &&
  go run . "$@"
