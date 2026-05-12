package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type ElasticsearchClient struct {
	Host     string
	Port     int
	Username string
	Password string
	Scheme   string
	Path     string

	client *http.Client
}

func NewElasticsearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "elasticsearch",
		Short: "Elasticsearch tools",
	}

	cmd.AddCommand(newPingElasticsearchCmd())

	return cmd
}

func newPingElasticsearchCmd() *cobra.Command {
	client := &ElasticsearchClient{}

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping an Elasticsearch cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.InitClient(); err != nil {
				log.Fatalf("Failed to init client: %v", err)
			}

			if err := client.CheckConnection(); err != nil {
				log.Fatalf("Failed to ping elasticsearch server: %v", err)
			}
			fmt.Println("Ping elasticsearch server success")
		},
	}

	addElasticsearchFlags(cmd, client)

	return cmd
}

func addElasticsearchFlags(cmd *cobra.Command, client *ElasticsearchClient) {
	cmd.Flags().StringVar(&client.Host, "host", "localhost", "Elasticsearch server host")
	cmd.Flags().IntVar(&client.Port, "port", 9200, "Elasticsearch server port")
	cmd.Flags().StringVar(&client.Username, "user", "", "Elasticsearch username")
	cmd.Flags().StringVar(&client.Password, "password", "", "Elasticsearch password")
	cmd.Flags().StringVar(&client.Scheme, "scheme", "http", "Elasticsearch HTTP scheme")
	cmd.Flags().StringVar(&client.Path, "path", "/_cluster/health", "Elasticsearch health check path")
}

func (c *ElasticsearchClient) InitClient() error {
	c.Scheme = strings.TrimSuffix(c.Scheme, "://")
	if c.Scheme == "" {
		c.Scheme = "http"
	}
	if c.Path == "" {
		c.Path = "/_cluster/health"
	}
	if !strings.HasPrefix(c.Path, "/") {
		c.Path = "/" + c.Path
	}

	c.client = &http.Client{Timeout: 10 * time.Second}
	return nil
}

func (c *ElasticsearchClient) CheckConnection() error {
	if c.client == nil {
		return fmt.Errorf("http client is not initialized")
	}

	url := fmt.Sprintf("%s://%s:%d%s", c.Scheme, c.Host, c.Port, c.Path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if c.Username != "" || c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("unexpected status %d from elasticsearch: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
		if status, ok := payload["status"].(string); ok {
			fmt.Printf("Elasticsearch cluster health status: %s\n", status)
		}
	}

	return nil
}
