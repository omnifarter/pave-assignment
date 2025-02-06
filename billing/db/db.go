package db

import "encore.dev/storage/sqldb"


var BillDb = sqldb.NewDatabase("bill", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})