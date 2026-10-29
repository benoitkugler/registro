package compta

// func TestGetCamps(t *testing.T) {
// 	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
// 		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
// 		"../../migrations/init.sql")
// 	defer db.Remove()

// 	_, _ = createCamps(t, db)

// 	ct := Controller{db: db.DB}
// 	l, err := ct.getCamps("localhost")
// 	tu.AssertNoErr(t, err)
// 	tu.Assert(t, len(l) == 2)

// 	_, err = DirecteurURL(ct.key, "localhost", l[0].IdCamp)
// 	tu.AssertNoErr(t, err)
// }
