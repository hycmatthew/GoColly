package databaseLogic

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DBRecord struct {
	Brand    string
	Name     string
	PriceCN  string
	LinkSpec string
	LinkCN   string
	LinkUS   string
	LinkHK   string
}

func CreateDBLogic() {
	const (
		cpuRecord         = "cpuRecord"
		gpuRecord         = "gpuRecord"
		motherboardRecord = "motherboardRecord"
		ramRecord         = "ramRecord"
		ssdRecord         = "ssdRecord"
		powerRecord       = "powerRecord"
		coolerRecord      = "coolerRecord"
		pcCaseRecord      = "caseRecord"
	)

	const (
		cpudb         = "cpuData"
		gpudb         = "gpuData"
		motherboarddb = "motherboardData"
		ramdb         = "ramData"
		ssddb         = "ssdData"
		powerdb       = "powerData"
		coolerdb      = "coolerData"
		pcCasedb      = "caseData"
	)

	createDB()
}

func openDataBase() *sql.DB {
	db, err := sql.Open("sqlite3", "../pcdata.db")
	checkErr(err)
	return db
}

func InsertRecord(part string, data DBRecord) {
	db := openDataBase()
	// row, err := db.Query("SELECT * as count from motherboardRecord WHERE rollno = ? AND isReward = ?", 1, 1)
	sql := `REPLACE INTO motherboardRecord (brand, name, price_cn, spec, link_cn, link_us, link_hk) 
            VALUES (?, ?, ?, ?, ?, ?, ?);`
	fmt.Println(data.Brand, " : ", data.Name)

	row, err := db.Exec(sql, data.Brand, data.Name, data.PriceCN, data.LinkSpec, data.LinkCN, data.LinkUS, data.LinkHK)
	checkErr(err)

	affect, err := row.RowsAffected()
	checkErr(err)
	fmt.Println(affect)
}

func UpdateData(part string, data DBRecord) {
	db := openDataBase()
	// row, err := db.Query("SELECT * as count from motherboardRecord WHERE rollno = ? AND isReward = ?", 1, 1)
	sql := `REPLACE INTO motherboardData (brand, name, price_cn, spec, link_cn, link_us, link_hk) 
            VALUES (?, ?, ?, ?, ?, ?, ?);`
	fmt.Println(data.Brand, " : ", data.Name)

	row, err := db.Exec(sql, data.Brand, data.Name, data.PriceCN, data.LinkSpec, data.LinkCN, data.LinkUS, data.LinkHK)
	checkErr(err)

	affect, err := row.RowsAffected()
	checkErr(err)
	fmt.Println(affect)
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
