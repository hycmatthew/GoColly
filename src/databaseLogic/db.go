package databaseLogic

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func CreateDBLogic() {
	createDB()
}

func createDB() {
	db, err := sql.Open("sqlite3", "../pcdata.db")
	checkErr(err)

	//插入資料
	stmt, err := db.Prepare("INSERT INTO motherboardSpec(code, name) values(?,?)")
	checkErr(err)

	res, err := stmt.Exec("astaxie", "研發部門")
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)

	fmt.Println(affect)

	//查詢資料
	rows, err := db.Query("SELECT * FROM motherboardSpec")
	checkErr(err)

	for rows.Next() {
		var id int
		var code string
		var name string
		err = rows.Scan(&id, &code, &name)
		checkErr(err)
		fmt.Println(id)
		fmt.Println(code)
		fmt.Println(name)
	}

	db.Close()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
