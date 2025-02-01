package testutils

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"registro/config"
)

func getUserName() string {
	var buf bytes.Buffer
	cmd := exec.Command("whoami")
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}

func runCmd(cmd *exec.Cmd) {
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		fmt.Println(stdOut.String())
		fmt.Println(stdErr.String())
		panic(err)
	}
}

type TestDB struct {
	*sql.DB
	name string // unique randomly generated
}

var (
	dbCount      int
	dbCountMutex sync.Mutex
)

// NewTestDB creates a new database and add all the tables
// as defined in the `generateSQLFile` files.
func NewTestDB(t *testing.T, generateSQLFile ...string) TestDB {
	const userPassword = "dummy"

	dbCountMutex.Lock()
	name := fmt.Sprintf("tmp_dev_%d_%d", time.Now().UnixNano(), dbCount)
	dbCount++
	dbCountMutex.Unlock()

	runCmd(exec.Command("createdb", name))

	for _, file := range generateSQLFile {
		fileA, err := filepath.Abs(file)
		if err != nil {
			t.Fatal(err)
		}
		_, err = os.Stat(fileA)
		if err != nil {
			t.Fatal(err)
		}
		runCmd(exec.Command("bash", "-c", fmt.Sprintf("psql %s < %s", name, fileA)))
	}

	creds := config.DB{
		Name:     name,
		Host:     "localhost",
		User:     getUserName(),
		Password: userPassword,
		Port:     5432,
	}
	db, err := creds.ConnectPostgres()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	t.Logf("Successfully created dev DB %s", name)

	return TestDB{DB: db, name: name}
}

// Remove closes the connection and remove the DB.
func (db TestDB) Remove() {
	err := db.DB.Close()
	if err != nil {
		panic(err)
	}

	runCmd(exec.Command("dropdb", "--if-exists", "--force", "--username="+getUserName(), db.name))
}

// SampleDBACVE is a test DB manually setup
var SampleDBACVE = config.DB{
	Host:     "localhost",
	User:     "benoit",
	Password: "dummy",
	Name:     "tmp_acve",
	Port:     5432,
}
