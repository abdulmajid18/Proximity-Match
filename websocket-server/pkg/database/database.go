package database

import (
	"log"
	"time"

	"github.com/gocql/gocql"
)

var session *gocql.Session

func Init(keyspace string) {
	cluster := gocql.NewCluster("localhost")
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = time.Second * 10
	cluster.ConnectTimeout = time.Second * 10

	// Set the keyspace in the cluster configuration

	var err error
	session, err = cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to create Cassandra session: %v", err)
	}

	err = createKeyspace(keyspace)
	if err != nil {
		log.Fatalf("Failed to create keyspace: %v", err)
	}

	err = createTable(keyspace)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func GetSession() *gocql.Session {
	return session
}

func createKeyspace(keyspace string) error {
	query := `
		CREATE KEYSPACE IF NOT EXISTS ` + keyspace + `
		WITH REPLICATION = {
			'class' : 'SimpleStrategy',
			'replication_factor' : 1
		};
	`
	return session.Query(query).Exec()
}

func createTable(keyspace string) error {
	query := `
		CREATE TABLE IF NOT EXISTS ` + keyspace + `.locations (
			user_id UUID PRIMARY KEY,
			current_latitude DOUBLE,
			current_longitude DOUBLE,
			destination_latitude DOUBLE,
			destination_longitude DOUBLE,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		);
	`
	return session.Query(query).Exec()
}
