package dialectquery

import (
	"fmt"
	"strings"
)

type Clickhouse struct {
	ClusterName string
}

var _ Querier = (*Clickhouse)(nil)

func (c *Clickhouse) CreateTable(tableName string) string {
	if c.ClusterName != "" {
		var dbName string
		split := strings.SplitN(tableName, ".", 2)
		if len(split) != 2 {
			dbName = "default"
		} else {
			dbName = split[0]
			tableName = split[1]
		}

		fullTableName := fmt.Sprintf("%s.%s", dbName, tableName)
		const localPostfix = "_local_v1"

		return `CREATE TABLE IF NOT EXISTS ` + fullTableName + localPostfix + ` ON CLUSTER '` + c.ClusterName + `' (
version_id Int64,
is_applied UInt8,
date Date default now(),
tstamp DateTime default now()
)
ENGINE = ReplicatedMergeTree('/clickhouse/{installation}/{cluster}/tables/{shard}/` + dbName + `/` + tableName + localPostfix + `', '{replica}')
ORDER BY (date);

CREATE TABLE IF NOT EXISTS ` + fullTableName + ` ON CLUSTER '` + c.ClusterName + `' AS ` + fullTableName + localPostfix + `
ENGINE = Distributed('` + c.ClusterName + `', ` + dbName + `, '` + tableName + localPostfix + `', rand());
ORDER BY (date);`
	}

	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			version_id Int64,
			is_applied UInt8,
			date Date default now(),
			tstamp DateTime default now()
			)
	  ENGINE = MergeTree()
		ORDER BY (date)`, tableName)
}

func (c *Clickhouse) InsertVersion(tableName string) string {
	q := `INSERT INTO %s (version_id, is_applied) VALUES ($1, $2)`
	return fmt.Sprintf(q, tableName)
}

func (c *Clickhouse) DeleteVersion(tableName string) string {
	q := `ALTER TABLE %s DELETE WHERE version_id = $1 SETTINGS mutations_sync = 2`
	return fmt.Sprintf(q, tableName)
}

func (c *Clickhouse) GetMigrationByVersion(tableName string) string {
	q := `SELECT tstamp, is_applied FROM %s WHERE version_id = $1 ORDER BY tstamp DESC LIMIT 1`
	return fmt.Sprintf(q, tableName)
}

func (c *Clickhouse) ListMigrations(tableName string) string {
	q := `SELECT version_id, is_applied FROM %s ORDER BY version_id DESC`
	return fmt.Sprintf(q, tableName)
}

func (c *Clickhouse) GetLatestVersion(tableName string) string {
	q := `SELECT max(version_id) FROM %s`
	return fmt.Sprintf(q, tableName)
}
