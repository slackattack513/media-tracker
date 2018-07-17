package databaseDriver

import (
	"database/sql"
	"fmt"
	mysqlerr "mysqlerror"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// Contains *sql.DB object and implements getDBConnection() to return this object
type DataBaseConnectionObject struct {
	DB *sql.DB
}

func NewDataBaseConnectionObject(db *sql.DB) DataBaseConnectionObject {
	return DataBaseConnectionObject{db}
}

func (dbco *DataBaseConnectionObject) getDBConnection() *sql.DB {
	return dbco.DB
}

// Interface to establish behavior to make a database query and parse the response
// type DatabaseQuery interface {
// 	setQueryStatement(string)
// 	getQueryStatement() string
// }

// Struct that implements DatabaseQuery interface and embeds DataBaseConnectionObject to allow for db queries
// type DataBaseQueryObject struct {
// 	DataBaseConnectionObject
// }

// General method to create a query statement. Specific types of objects should implement their own version of this method.
// func (DBQO *DataBaseTableRequestObject) getQueryStatement() string {
// 	return ""
// }

// General method to execute a database query. This is boilerplate code that should work for any query assuming the object has properly implemented the getQueryStatement() method
func (DBQO *DataBaseTableRequestObject) RunDBQuery() (*sql.Rows, error) {
	return DBQO.getDBConnection().Query(DBQO.GetRequestStatement())
}

// Parses the *sql.Rows response of a query into a slice of maps ( {key:val} pairs)
func (DBQO *DataBaseTableRequestObject) ParseQueryRowsResponse(rows *sql.Rows) []map[string]interface{} {
	var retrievedDataSlice []map[string]interface{}

	cols, _ := rows.Columns()

	// Modified from https://kylewbanks.com/blog/query-result-to-map-in-golang
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		retrievedDataSlice = append(retrievedDataSlice, m)
	}
	return retrievedDataSlice
}

type DataBaseTableRequestObjectI interface {
	Exists(string) bool
	TableExists() bool
	ColumnExists(string) bool
	CellExists(string, interface{}) bool
	PrepareRequestStatement()
	SetRequestStatement(string)
	GetRequestStatement() string
	RunDBExecution() (sql.Result, error)
	RunDBQuery() (*sql.Rows, error)
	ParseQueryRowsResponse(*sql.Rows) []map[string]interface{}
	GetCellValue(string, map[string]string) interface{}
}

type DataBaseTableRequestObject struct {
	DataBaseConnectionObject
	DataBaseTableObject
	RequestStatement string
}

func NewDataBaseTableRequestObject(DBCO DataBaseConnectionObject, DBTO DataBaseTableObject, requestStatement string) DataBaseTableRequestObject {
	return DataBaseTableRequestObject{DBCO, DBTO, requestStatement}
}

func (DBTRO *DataBaseTableRequestObject) GetCellValue(desCol, constraintMap string) interface{} {
	// TODO
}

func (DBTRO *DataBaseTableRequestObject) Exists(statement string) bool {
	DBTRO.SetRequestStatement(statement)
	rows, err := DBTRO.RunDBQuery()

	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == mysqlerr.ER_BAD_FIELD_ERROR {

				return false
			}
		}
		fmt.Printf("Error checking existence of:\n%s\n", statement)
		fmt.Println(err)
		os.Exit(1)
	}

	if rows.Next() {
		rows.Close()
		return true
	}

	rows.Close()
	return false
}

func (DBTRO *DataBaseTableRequestObject) TableExists() bool {
	existenceCommand := "SELECT * FROM information_schema.tables WHERE table_schema = '" + DBTRO.GetDatabaseName() + "'AND table_name = '" + DBTRO.GetTableName() + "'"
	return DBTRO.Exists(existenceCommand)
}

func (DBTRO *DataBaseTableRequestObject) CellExists(columnName string, cellValue interface{}) bool {
	cellValueString := ""

	switch cellValue.(type) {
	case string:
		cellValueString = "'" + cellValue.(string) + "'"
	case int:
		cellValueString = strconv.Itoa(cellValue.(int))
	case bool:
		if cellValue.(bool) {
			cellValueString = "1"
		} else {
			cellValueString = "0"
		}
	default:
		// Todo
		return true
	}
	existenceCommand := "SELECT " + columnName + " FROM " + DBTRO.GetTableName() + " WHERE " + columnName + " = " + cellValueString + ";"
	return DBTRO.Exists(existenceCommand)
}

func (DBTRO *DataBaseTableRequestObject) ColumnExists(columnName string) bool {
	existenceCommand := "SELECT " + columnName + " FROM " + DBTRO.GetTableName() + ";"
	return DBTRO.Exists(existenceCommand)
}

// DatabaseExecution interface{} provides general interface to execute call to database with some execution statement
// type DatabaseExecution interface {

// }

// DataBaseExecutionObject struct implements DatabaseExecution and embeds dataBaseConnectionObject
// Embedding allows for access to the database connection object (in mysql this is *sql.DB)
// Structs embedding 'this' should rewrite their own getExecutionStatement() functions to ensure proper calls
// to the database
// type DataBaseExecutionObject struct {
// 	DataBaseTableRequestObject
// }

func (DBR *DataBaseTableRequestObject) PrepareRequestStatement() {
	DBR.SetRequestStatement("")
}

func (DBR *DataBaseTableRequestObject) RunDBExecution() (sql.Result, error) {
	return DBR.getDBConnection().Exec(DBR.GetRequestStatement())
}

func (DBR *DataBaseTableRequestObject) GetRequestStatement() string {
	return DBR.RequestStatement
}

func (DBR *DataBaseTableRequestObject) SetRequestStatement(newStatement string) {
	DBR.RequestStatement = newStatement
}

// type dataRequest struct {
// 	dataBaseTableObject
// 	desiredColumns    []string
// 	columnConstraints map[string]interface{}
// }

// type dataSelectObjectI interface {
// 	getDataMap() map[string]interface{}
// }

// type DataSelectObject struct {
// 	DataBaseExecutionObject
// 	dataBaseTableObject
// }

type DataInsertObjectI interface {
	CreateDataMap()
	SetDataMap(map[string]interface{})
	GetDataMap() map[string]interface{}
}

type DataInsertObject struct {
	DataBaseTableRequestObject
	dataMap map[string]interface{}
}

func NewDataInsertObject(DBTRO DataBaseTableRequestObject, dm map[string]interface{}) DataInsertObject {
	return DataInsertObject{DBTRO, dm}
}

func (DIO *DataInsertObject) SetDataMap(newMap map[string]interface{}) {
	DIO.dataMap = newMap
}

func (DIO *DataInsertObject) CreateDataMap() {
	DIO.dataMap = make(map[string]interface{})
}

func (DIO *DataInsertObject) PrepareRequestStatement() {
	db := DIO.getDBConnection()
	databaseName := DIO.GetDatabaseName()
	tableName := DIO.GetTableName()
	newData := DIO.GetDataMap()

	changeDatabase(db, databaseName)

	//insertData
	var columns string
	var values string

	firstTime := true
	for key, val := range newData {
		if !firstTime {
			columns = columns + ","
			values = values + ","
		} else {
			firstTime = false
		}
		if !DIO.ColumnExists(key) {
			createColumn(db, tableName, key, val)
		}
		columns = columns + key

		switch val.(type) {
		case string:
			values = values + "'" + val.(string) + "'"
		case int, int64:
			values = values + strconv.Itoa(val.(int))
		case uint64:
			values = values + strconv.FormatUint(val.(uint64), 10)
		case bool:
			if val.(bool) {
				values = values + "1"
			} else {
				values = values + "0"
			}
		}

	}

	if columns == "" && values == "" {
		DIO.SetRequestStatement("")
	} else {
		DIO.SetRequestStatement("INSERT INTO " + tableName + "( " + columns + " ) VALUES (" + values + ");")
	}

}

func (DIO *DataInsertObject) GetDataMap() map[string]interface{} {
	return DIO.dataMap
}

type DataBaseTableObjectI interface {
	GetDatabaseName() string
	GetTableName() string
}

type DataBaseTableObject struct {
	DatabaseName string
	TableName    string
}

func NewDataBaseTableObject(databaseName, tableName string) DataBaseTableObject {
	return DataBaseTableObject{databaseName, tableName}
}

func (obj *DataBaseTableObject) GetDatabaseName() string {
	return obj.DatabaseName
}

func (obj *DataBaseTableObject) GetTableName() string {
	return obj.TableName
}

func OpenDBConnection(database, username, password string) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", username, password, database)) //Onyixj@bs$@/youtubedata")
	return db, err
}

func OpenAndPingDBConnection(database, username, password string) (*sql.DB, error) {
	db, err := OpenDBConnection(database, username, password)
	if err != nil {
		fmt.Println("Error opening DB connection")
	} else {
		// Test DB connection with ping
		err = db.Ping()
		if err != nil {
			fmt.Println("Error pinging DB")
		}
	}

	return db, err

}

func changeDatabase(db *sql.DB, databasename string) {
	_, err := db.Exec("USE " + databasename)
	// fmt.Println(res)
	if err != nil {
		fmt.Printf("Error using database %s\n", databasename)
		os.Exit(1)
	}
}

func columnExists(db *sql.DB, tableName string, columnName string) bool {
	existenceCommand := "SELECT " + columnName + " FROM " + tableName + ";"
	rows, err := db.Query(existenceCommand)

	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == mysqlerr.ER_BAD_FIELD_ERROR {

				return false
			}
		}
		fmt.Println("Error checking column existence")
		fmt.Println(err)
		os.Exit(1)
	}

	if rows.Next() {
		rows.Close()
		return true
	}
	rows.Close()
	return false
}

func tableExists(db *sql.DB, databasename string, tableName string) bool {
	existenceCommand := "SELECT * FROM information_schema.tables WHERE table_schema = '" + databasename + "'AND table_name = '" + tableName + "'"
	rows, err := db.Query(existenceCommand)

	if err != nil {
		fmt.Println("Error checking table existence")
		os.Exit(1)
	}

	defer rows.Close()

	if rows.Next() {
		return true
	}

	return false

}

func cellExists() {

}

// func getDataFromTable(tableName string) {
// 	var (
// 		id   int
// 		name string
// 	)
// 	rows, err := db.Query("select id, name from users where id = ?", 1)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		err := rows.Scan(&id, &name)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Println(id, name)
// 	}
// 	err = rows.Err()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// For now, this assumes that constraints on the retrieved data are strings
func GetDataFromTable(db *sql.DB, databaseName, tableName string, columnNames []string, dataValues map[string]string) []map[string]interface{} {
	changeDatabase(db, databaseName)

	var columnsOrAll string
	if len(columnNames) == 0 {
		columnsOrAll = " * "
	} else {
		columnsOrAll = strings.Join(columnNames, ", ")
	}

	prepareStatement := "SELECT " + columnsOrAll + " FROM " + tableName

	if len(dataValues) > 0 {
		prepareStatement = prepareStatement + " WHERE "

		counter := 1
		for key, val := range dataValues {
			prepareStatement = prepareStatement + key + " IN " + "('" + val + "')"

			if counter != len(dataValues) {
				prepareStatement = prepareStatement + " OR "
			}
			counter++

		}
	}

	prepareStatement = prepareStatement + ";"

	rows, err := db.Query(prepareStatement)

	if err != nil {
		fmt.Println("Error querying table")
		fmt.Print(err)
		os.Exit(1)
	}
	cols, _ := rows.Columns()

	defer rows.Close()

	var retrievedDataSlice []map[string]interface{}

	// Modified from https://kylewbanks.com/blog/query-result-to-map-in-golang
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		retrievedDataSlice = append(retrievedDataSlice, m)
	}
	return retrievedDataSlice
}

func createColumn(db *sql.DB, tableName, columnName string, dataType interface{}) {

	// if err != nil {
	// 	fmt.Println("Error creating column prepare statement")
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	var sql_datatype string
	switch dataType.(type) {
	case string:
		sql_datatype = "VARCHAR(300)"
	case int:
		sql_datatype = "INT"
	case bool:
		sql_datatype = "BIT"
	default:
		sql_datatype = "VARCHAR(300)"
	}

	prepareStatement := "ALTER TABLE " + tableName + " ADD " + columnName + " " + sql_datatype + ";"
	_, err := db.Exec(prepareStatement)

	if err != nil {
		fmt.Println("Error creating column")
		fmt.Println(err)
		os.Exit(1)
	}

}

func AddDataToTable(db *sql.DB, databaseName, tableName string, newData map[string]interface{}) {
	changeDatabase(db, databaseName)
	// fmt.Println(tableName)
	// prepState := "INSERT INTO " + tableName + "( ? ) VALUES (?);"
	// fmt.Println(prepState)
	// preparedInsert, err := db.Prepare(prepState) // ? = placeholder
	// if err != nil {
	// 	fmt.Println("Error creating 'prepare' statement in AddDataToTable")
	// 	fmt.Println(err)
	// 	debug.PrintStack()
	// 	os.Exit(1)
	// }
	// defer preparedInsert.Close() // Close the statement when we leave main() / the program terminates

	//insertData
	var columns string
	var values string

	firstTime := true
	for key, val := range newData {
		// fmt.Println(key)
		// fmt.Println(reflect.TypeOf(val))
		// fmt.Println()

		// fmt.Println(columns)
		// fmt.Println(values)
		// fmt.Println()

		if !firstTime {
			columns = columns + ","
			values = values + ","
		} else {
			firstTime = false
		}
		if !columnExists(db, tableName, key) {
			createColumn(db, tableName, key, val)
		}
		columns = columns + key

		switch val.(type) {
		case string:
			values = values + "'" + val.(string) + "'"
		case int, int64:
			values = values + strconv.Itoa(val.(int))
		case uint64:
			values = values + strconv.FormatUint(val.(uint64), 10)
		case bool:
			if val.(bool) {
				values = values + "1"
			} else {
				values = values + "0"
			}
		}

	}

	// fmt.Println("columns: ")
	// prettyPrinting.PrintJSON(columns)
	// fmt.Println("values: ")
	// prettyPrinting.PrintJSON(values)
	prepState := "INSERT INTO " + tableName + "( " + columns + " ) VALUES (" + values + ");"

	// prettyPrinting.PrintJSON(prepState)

	_, err := db.Exec(prepState)

	if err != nil {
		fmt.Println("Error executing insertion statement in AddDataToTable")
		fmt.Println(err)
		debug.PrintStack()
		os.Exit(1)
	}
	// defer preparedInsert.Close() // Close the statement when we leave main() / the program terminates

}

func MakeYoutubePlaylistTable(db *sql.DB, databaseName, tableName string) bool {

	changeDatabase(db, databaseName)
	if !tableExists(db, databaseName, tableName) {
		_, err := db.Exec("CREATE TABLE " + tableName + " (videoID VARCHAR(500) NOT NULL, videoTableID INT NOT NULL, " + tableName + "TableID INT NOT NULL AUTO_INCREMENT PRIMARY KEY);")
		if err != nil {
			fmt.Printf("Error making table", tableName)
			return false
			os.Exit(1)
		}
		fmt.Printf("Created table")
		return true
	} else {
		fmt.Printf("Table already exists")
		return false
	}
}
