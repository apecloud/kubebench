package tools

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBClient struct {
	Host     string
	Port     int
	Username string
	Password string

	client *mongo.Client
}

func NewMongoDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mongodb",
		Short: "MongoDB tools",
		Long:  "MongoDB tools for creating and dropping databases",
	}

	cmd.AddCommand(newCreateMongoDBCmd())
	cmd.AddCommand(newDropMongoDBCmd())
	cmd.AddCommand(newPingMongoDBCmd())

	return cmd
}

func newCreateMongoDBCmd() *cobra.Command {
	client := &MongoDBClient{}

	cmd := &cobra.Command{
		Use:   "create [database name]",
		Short: "Create new databases",
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

	addMongoDBFlags(cmd, client)

	return cmd
}

func newDropMongoDBCmd() *cobra.Command {
	client := &MongoDBClient{}

	cmd := &cobra.Command{
		Use:   "drop [database name]",
		Short: "Drop databases",
		Run: func(cmd *cobra.Command, args []string) {
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

	addMongoDBFlags(cmd, client)

	return cmd
}

func newPingMongoDBCmd() *cobra.Command {
	client := &MongoDBClient{}

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping a MongoDB server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				fmt.Println(err)
				return
			}
			defer client.Close()

			if err := client.CheckConnection(); err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("MongoDB server is up and running")
		},
	}

	addMongoDBFlags(cmd, client)

	return cmd
}

func (c *MongoDBClient) InitClient() error {
	mongodbURI := fmt.Sprintf("mongodb://%s:%s@%s:%d", c.Username, c.Password, c.Host, c.Port)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongodbURI))
	if err != nil {
		return err
	}
	c.client = client

	return nil
}

func (c *MongoDBClient) Close() error {
	return c.client.Disconnect(context.Background())
}

func (c *MongoDBClient) CreateDatabase(name string) error {
	// if not exists, create database
	return c.client.Database(name).CreateCollection(context.Background(), "test", nil)
}

func (c *MongoDBClient) DropDatabase(name string) error {
	return c.client.Database(name).Drop(context.Background())
}

func (c *MongoDBClient) CheckConnection() error {
	return c.client.Ping(context.Background(), nil)
}

func addMongoDBFlags(cmd *cobra.Command, client *MongoDBClient) {
	cmd.Flags().StringVarP(&client.Host, "host", "", "localhost", "MongoDB host")
	cmd.Flags().IntVarP(&client.Port, "port", "", 27017, "MongoDB port")
	cmd.Flags().StringVarP(&client.Username, "user", "", "", "MongoDB username")
	cmd.Flags().StringVarP(&client.Password, "password", "", "", "MongoDB password")
}
