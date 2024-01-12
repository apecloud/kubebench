package tools

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

var DefaultPGDatabase = "postgres"

type PostgreSQLClient struct {
	Host     string
	Port     int
	Username string
	Password string

	db *sql.DB
}

func NewPostgreSQLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "postgresql",
		Short: "PostgreSQL tools",
	}

	cmd.AddCommand(newCreatePgDatabaseCmd())
	cmd.AddCommand(newDropPgDatabaseCmd())
	cmd.AddCommand(newPingPgDatabaseCmd())

	return cmd
}

func newCreatePgDatabaseCmd() *cobra.Command {
	client := &PostgreSQLClient{}

	cmd := &cobra.Command{
		Use:   "create [database name]",
		Short: "Create a new database",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				fmt.Println(err)
				return
			}
			defer client.Close()

			for _, name := range args {
				if err := client.CreateDatabase(name); err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("Database %s created\n", name)
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
			client := &PostgreSQLClient{}

			if err := client.InitClient(); err != nil {
				fmt.Println(err)
				return
			}
			defer client.Close()

			for _, name := range args {
				if err := client.DropDatabase(name); err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("Database %s dropped\n", name)
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
				fmt.Println(err)
				return
			}
			defer client.Close()

			fmt.Println("Ping database success")
		},
	}

	addPostgreSQLFlags(cmd, client)

	return cmd
}

func (c *PostgreSQLClient) InitClient() error {
	// create connection string with default pg database
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, DefaultPGDatabase)

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
	// create database if not exists for postgresql
	query := fmt.Sprintf("CREATE DATABASE %s IF NOT EXISTS", name)
	return c.Exec(query)
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

func addPostgreSQLFlags(cmd *cobra.Command, client *PostgreSQLClient) {
	cmd.Flags().StringVar(&client.Host, "host", "localhost", "PostgreSQL host")
	cmd.Flags().IntVar(&client.Port, "port", 5432, "PostgreSQL port")
	cmd.Flags().StringVar(&client.Username, "user", "postgres", "PostgreSQL username")
	cmd.Flags().StringVar(&client.Password, "password", "", "PostgreSQL password")
}
