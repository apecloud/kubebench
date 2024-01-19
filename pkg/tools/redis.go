package tools

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

type RedisClient struct {
	Host     string
	Port     int
	Username string
	Password string

	client *redis.Client
}

func NewRedisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redis",
		Short: "Redis tools",
	}

	cmd.AddCommand(newPingRedisCmd())

	return cmd
}

func newPingRedisCmd() *cobra.Command {
	client := &RedisClient{}

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping redis server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				log.Fatalf("Failed to init client: %v", err)
			}
			defer client.Close()

			if err := client.CheckConnection(); err != nil {
				log.Fatalf("Failed to ping redis server: %v", err)
			}
			fmt.Printf("Ping redis server success\n")
		},
	}

	addRedisFlags(cmd, client)

	return cmd
}

func addRedisFlags(cmd *cobra.Command, client *RedisClient) {
	cmd.Flags().StringVar(&client.Host, "host", "localhost", "Redis server host")
	cmd.Flags().IntVar(&client.Port, "port", 6379, "Redis server port")
	cmd.Flags().StringVar(&client.Username, "user", "", "Redis server username")
	cmd.Flags().StringVar(&client.Password, "password", "", "Redis server password")
}

func (c *RedisClient) InitClient() error {

	c.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Username: c.Username,
		Password: c.Password,
		DB:       0,
	})

	return nil
}

func (c *RedisClient) Close() error {
	return c.client.Close()
}

func (c *RedisClient) CheckConnection() error {
	_, err := c.client.Ping(context.TODO()).Result()
	if err != nil {
		return err
	}

	return nil
}
