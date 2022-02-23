package easy_db

import (
	"testing"

	_ "github.com/lib/pq"
)

type UserRecord struct {
	ID       int64
	Username string
	Email    string
}

func TestPgSqlQuery(t *testing.T) {
	// driver, err := Open("postgres", "host=localhost port=5432 user=cherrydb dbname=cherrydb password=Cherrydb_123! sslmode=disable")
	// if err != nil {
	// 	panic(err)
	// }
	// context := context.Background()
	// result, err := driver.QueryContext(context, "select id, username, email from users")
	// if err != nil {

	// 	panic(err)
	// }
	// defer result.Close()
	// for result.Next() {
	// 	var record UserRecord
	// 	if err = result.Scan(&record.ID, &record.Username, &record.Email); err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(record.ID, record.Username, record.Email)
	// }

}
