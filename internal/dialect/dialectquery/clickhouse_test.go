package dialectquery

import (
	"fmt"
	"testing"
)

func TestClickhouse_CreateTable(t *testing.T) {
	c := &Clickhouse{
		ClusterName: "prod01",
	}
	tableName := "traffic.goose_db_version"

	s := c.CreateTable(tableName)
	fmt.Println(s)
}
