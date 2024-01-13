package tools

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

type MySQLClient struct {
	Host     string
	Port     int
	Username string
	Password string

	db *sql.DB
}

func NewMySQLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mysql",
		Short: "MySQL tools",
		Long:  "MySQL tools for creating and dropping databases",
	}

	cmd.AddCommand(newCreateMysqlDatabaseCmd())
	cmd.AddCommand(newDropMysqlDatabaseCmd())
	cmd.AddCommand(newPingMysqlCmd())

	return cmd
}

func newCreateMysqlDatabaseCmd() *cobra.Command {
	client := &MySQLClient{}

	cmd := &cobra.Command{
		Use:   "create [database name]",
		Short: "Create new databases",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				log.Fatalf("failed to connect to MySQL server: %v", err)
			}
			defer client.Close()

			for _, name := range args {
				if err := client.CreateDatabase(name); err != nil {
					log.Fatalf("failed to create database %s: %v", name, err)
				}
				fmt.Printf("Database %s created\n", name)
			}
		},
	}

	addMysqlFlags(cmd, client)

	return cmd
}

func newDropMysqlDatabaseCmd() *cobra.Command {
	client := &MySQLClient{}

	cmd := &cobra.Command{
		Use:   "drop [database name]",
		Short: "Drop databases",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				log.Fatalf("failed to connect to MySQL server: %v", err)
			}
			defer client.Close()

			for _, name := range args {
				if err := client.DropDatabase(name); err != nil {
					log.Fatalf("failed to drop database %s: %v", name, err)
				}
				fmt.Printf("Database %s dropped\n", name)
			}
		},
	}

	addMysqlFlags(cmd, client)

	return cmd
}

func newPingMysqlCmd() *cobra.Command {
	client := &MySQLClient{}

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping MySQL server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				log.Fatalf("failed to connect to MySQL server: %v", err)
			}
			defer client.Close()

			if err := client.CheckConnection(); err != nil {
				log.Fatalf("failed to ping MySQL server: %v", err)
			}
			fmt.Println("Pong")
		},
	}

	addMysqlFlags(cmd, client)

	return cmd
}

func (c *MySQLClient) InitClient() error {
	var err error
	c.db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", c.Username, c.Password, c.Host, c.Port))
	if err != nil {
		return err
	}

	return nil
}

func (c *MySQLClient) Close() error {
	return c.db.Close()
}

func (c *MySQLClient) Exec(query string) error {
	_, err := c.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (c *MySQLClient) CheckConnection() error {
	return c.db.Ping()
}

func (c *MySQLClient) CreateDatabase(name string) error {
	// create database if not exists
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", name)
	return c.Exec(query)
}

func (c *MySQLClient) DropDatabase(name string) error {
	// drop database if exists
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)
	return c.Exec(query)
}

func addMysqlFlags(cmd *cobra.Command, client *MySQLClient) {
	cmd.Flags().StringVar(&client.Host, "host", "localhost", "MySQL host")
	cmd.Flags().IntVar(&client.Port, "port", 3306, "MySQL port")
	cmd.Flags().StringVar(&client.Username, "user", "root", "MySQL username")
	cmd.Flags().StringVar(&client.Password, "password", "", "MySQL password")
}
