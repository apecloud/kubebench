package tools

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"
	"github.com/spf13/cobra"
)

var DefaultPGDatabase = "postgres"

type PostgreSQLClient struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string

	db *sql.DB
}

func NewPostgreSQLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "postgresql",
		Short: "PostgreSQL tools",
	}

	cmd.AddCommand(newCreatePgDatabaseCmd())
	cmd.AddCommand(newDropPgDatabaseCmd())
	cmd.AddCommand(newDropPgTableCmd())
	cmd.AddCommand(newPingPgDatabaseCmd())

	return cmd
}

func newCreatePgDatabaseCmd() *cobra.Command {
	client := &PostgreSQLClient{}

	cmd := &cobra.Command{
		Use:   "create [database name]",
		Short: "Create a new database",
		Run: func(cmd *cobra.Command, args []string) {
			explicit := cmd.Flags().Changed("database")
			for _, name := range args {
				if err := client.EnsureDatabase(name, explicit); err != nil {
					log.Fatalf("Failed to create database: %v", err)
				}
			}
		},
	}

	addPostgreSQLFlags(cmd, client)

	return cmd
}

func newDropPgDatabaseCmd() *cobra.Command {
	client := &PostgreSQLClient{}

	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Drop PostgreSQL Server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				log.Fatalf("Failed to init client: %v", err)
			}
			defer client.Close()

			for _, name := range args {
				if err := client.DropDatabase(name); err != nil {
					log.Fatalf("Failed to drop database: %v", err)
				}
				fmt.Printf("Database %s dropped\n", name)
			}
		},
	}

	addPostgreSQLFlags(cmd, client)

	return cmd
}

func newDropPgTableCmd() *cobra.Command {
	client := &PostgreSQLClient{}

	cmd := &cobra.Command{
		Use:   "drop-table [table name]",
		Short: "Drop a table in the target database",
		Run: func(cmd *cobra.Command, args []string) {
			if !cmd.Flags().Changed("database") {
				log.Fatal("--database is required")
			}
			if err := client.InitClient(); err != nil {
				log.Fatalf("Failed to init client: %v", err)
			}
			defer client.Close()

			if err := client.CheckConnection(); err != nil {
				if isMissingDatabase(err) {
					fmt.Printf("Database %s does not exist, skip cleanup\n", client.Database)
					return
				}
				log.Fatalf("Failed to connect database: %v", err)
			}

			for _, name := range args {
				if err := client.DropTable(name); err != nil {
					log.Fatalf("Failed to drop table: %v", err)
				}
				fmt.Printf("Table %s dropped\n", name)
			}
		},
	}

	addPostgreSQLFlags(cmd, client)

	return cmd
}

func newPingPgDatabaseCmd() *cobra.Command {
	client := &PostgreSQLClient{}

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping a database",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				log.Fatalf("Failed to init client: %v", err)
			}
			defer client.Close()

			if err := client.CheckConnection(); err != nil {
				log.Fatalf("Failed to ping postgresql server: %v", err)
			}

			fmt.Println("Ping database success")
		},
	}

	addPostgreSQLFlags(cmd, client)

	return cmd
}

func (c *PostgreSQLClient) InitClient() error {
	database := c.Database
	if database == "" {
		database = DefaultPGDatabase
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, database)

	// open connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	c.db = db
	return nil
}

func (c *PostgreSQLClient) Close() error {
	return c.db.Close()
}

func (c *PostgreSQLClient) CreateDatabase(name string) error {
	// create database
	query := fmt.Sprintf("CREATE DATABASE %s ", name)
	if err := c.Exec(query); err != nil {
		// if database already exists, ignore error
		if strings.Contains(err.Error(), "already exists") {
			fmt.Printf("Database %s already exists\n", name)
			return nil
		}
		return err
	}

	return nil
}

func (c *PostgreSQLClient) DropDatabase(name string) error {
	// drop database if exists for postgresql
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)
	return c.Exec(query)
}

func (c *PostgreSQLClient) Exec(query string) error {
	if c.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	_, err := c.db.Exec(query)
	return err
}

func (c *PostgreSQLClient) CheckConnection() error {
	if c.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	return c.db.Ping()
}

func addPostgreSQLFlags(cmd *cobra.Command, client *PostgreSQLClient) {
	cmd.Flags().StringVar(&client.Host, "host", "localhost", "PostgreSQL host")
	cmd.Flags().IntVar(&client.Port, "port", 5432, "PostgreSQL port")
	cmd.Flags().StringVar(&client.Username, "user", "postgres", "PostgreSQL username")
	cmd.Flags().StringVar(&client.Password, "password", "", "PostgreSQL password")
	cmd.Flags().StringVar(&client.Database, "database", DefaultPGDatabase, "PostgreSQL database")
}

// fixed maintenance list. template1 is the CREATE DATABASE source, so never connect to it.
var pgMaintenanceDatabases = []string{"postgres", "test", "vastbase", "security"}

func maintenanceCandidates(target, explicit string, explicitSet bool) []string {
	if explicitSet && explicit != "" && explicit != target && explicit != "template1" {
		return []string{explicit}
	}
	out := make([]string, 0, len(pgMaintenanceDatabases))
	for _, name := range pgMaintenanceDatabases {
		if name == target {
			continue
		}
		out = append(out, name)
	}
	return out
}

func isMissingDatabase(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "3D000"
}

func safeIdent(name string) error {
	if name == "" {
		return fmt.Errorf("empty identifier")
	}
	for i, r := range name {
		ok := r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (i > 0 && r >= '0' && r <= '9')
		if !ok {
			return fmt.Errorf("invalid identifier %q", name)
		}
	}
	return nil
}

func (c *PostgreSQLClient) connStr(database string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, database)
}

func (c *PostgreSQLClient) open(database string) (*sql.DB, error) {
	db, err := sql.Open("postgres", c.connStr(database))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func (c *PostgreSQLClient) openMaintenance(target string, explicitSet bool) (*sql.DB, error) {
	var last error
	for _, name := range maintenanceCandidates(target, c.Database, explicitSet) {
		db, err := c.open(name)
		if err == nil {
			return db, nil
		}
		last = err
	}
	if last == nil {
		last = fmt.Errorf("no maintenance database available")
	}
	return nil, fmt.Errorf("connect maintenance database to create %s: %w", target, last)
}

func (c *PostgreSQLClient) EnsureDatabase(name string, explicitSet bool) error {
	db, err := c.open(name)
	if err == nil {
		_ = db.Close()
		fmt.Printf("Database %s already exists\n", name)
		return nil
	}
	if !isMissingDatabase(err) {
		return err
	}

	maint, err := c.openMaintenance(name, explicitSet)
	if err != nil {
		return err
	}
	defer maint.Close()

	c.db = maint
	if err := c.CreateDatabase(name); err != nil {
		return err
	}
	fmt.Printf("Database %s created\n", name)
	return nil
}

func (c *PostgreSQLClient) DropTable(name string) error {
	if err := safeIdent(name); err != nil {
		return err
	}
	return c.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", name))
}
