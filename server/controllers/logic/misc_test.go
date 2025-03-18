package logic

import (
	"database/sql"
	"testing"
	"time"

	pr "registro/sql/personnes"
	"registro/utils"
	tu "registro/utils/testutils"
)

func TestController_searchSimilaires(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	err := utils.InTx(db.DB, func(tx *sql.Tx) error {
		for range [2000]int{} {
			_, err := pr.Personne{}.Insert(tx)
			if err != nil {
				return err
			}
		}
		return nil
	})
	tu.AssertNoErr(t, err)

	ti := time.Now()
	_, err = SearchSimilaires(db, 1)
	tu.AssertNoErr(t, err)
	tu.Assert(t, time.Since(ti) < 50*time.Millisecond)
}
