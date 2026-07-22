package tools

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "gitee.com/opengauss/openGauss-connector-go-pq"
	"github.com/spf13/cobra"
)

// DefaultGaussDBDatabase is the default database used to connect to GaussDB
// (openGauss) before creating the target database. The openGauss default
// installation usually has a "postgres" database, and props.gaussdb uses it
// as the default connection database.
var DefaultGaussDBDatabase = "postgres"

type GaussDBClient struct {
	Host     string
	Port     int
	Username string
	Password string

	db *sql.DB
}

func NewGaussdbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gaussdb",
		Short: "GaussDB (openGauss) tools",
	}

	cmd.AddCommand(newCreateGaussdbDatabaseCmd())
	cmd.AddCommand(newDropGaussdbDatabaseCmd())
	cmd.AddCommand(newPingGaussdbDatabaseCmd())

	return cmd
}

func newCreateGaussdbDatabaseCmd() *cobra.Command {
	client := &GaussDBClient{}

	cmd := &cobra.Command{
		Use:   "create [database name]",
		Short: "Create a new GaussDB database",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				log.Fatalf("Failed to init client: %v", err)
			}
			defer client.Close()

			for _, name := range args {
				if err := client.CreateDatabase(name); err != nil {
					log.Fatalf("Failed to create database: %v", err)
				}
				fmt.Printf("Database %s created\n", name)
			}
		},
	}

	addGaussdbFlags(cmd, client)

	return cmd
}

func newDropGaussdbDatabaseCmd() *cobra.Command {
	client := &GaussDBClient{}

	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Drop GaussDB database",
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

	addGaussdbFlags(cmd, client)

	return cmd
}

func newPingGaussdbDatabaseCmd() *cobra.Command {
	client := &GaussDBClient{}

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping a GaussDB database",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				log.Fatalf("Failed to init client: %v", err)
			}
			defer client.Close()

			if err := client.CheckConnection(); err != nil {
				log.Fatalf("Failed to ping gaussdb server: %v", err)
			}

			fmt.Println("Ping database success")
		},
	}

	addGaussdbFlags(cmd, client)

	return cmd
}

func (c *GaussDBClient) InitClient() error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, DefaultGaussDBDatabase)

	db, err := sql.Open("opengauss", connStr)
	if err != nil {
		return err
	}

	c.db = db
	return nil
}

func (c *GaussDBClient) Close() error {
	if c.db == nil {
		return nil
	}
	return c.db.Close()
}

func (c *GaussDBClient) CreateDatabase(name string) error {
	query := fmt.Sprintf("CREATE DATABASE %s ", name)
	if err := c.Exec(query); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Printf("Database %s already exists\n", name)
			return nil
		}
		return err
	}

	return nil
}

func (c *GaussDBClient) DropDatabase(name string) error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)
	return c.Exec(query)
}

func (c *GaussDBClient) Exec(query string) error {
	if c.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	_, err := c.db.Exec(query)
	return err
}

func (c *GaussDBClient) CheckConnection() error {
	if c.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	return c.db.Ping()
}

func addGaussdbFlags(cmd *cobra.Command, client *GaussDBClient) {
	cmd.Flags().StringVar(&client.Host, "host", "localhost", "GaussDB host")
	cmd.Flags().IntVar(&client.Port, "port", 5432, "GaussDB port")
	cmd.Flags().StringVar(&client.Username, "user", "postgres", "GaussDB username")
	cmd.Flags().StringVar(&client.Password, "password", "", "GaussDB password")
}
