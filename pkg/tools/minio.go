package tools

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

type MinioClient struct {
	Host      string
	Port      int
	AccessKey string
	SecretKey string
	UseSSL    bool
}

func NewMinioCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "minio",
		Short: "MinIO tools",
	}

	cmd.AddCommand(newPingMinioCmd())

	return cmd
}

func newPingMinioCmd() *cobra.Command {
	client := &MinioClient{}

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping minio server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.CheckConnection(); err != nil {
				log.Fatalf("Failed to ping minio server: %v", err)
			}
			fmt.Printf("Ping minio server success\n")
		},
	}

	addMinioFlags(cmd, client)

	return cmd
}

func addMinioFlags(cmd *cobra.Command, client *MinioClient) {
	cmd.Flags().StringVar(&client.Host, "host", "localhost", "MinIO server host")
	cmd.Flags().IntVar(&client.Port, "port", 9000, "MinIO server port")
	cmd.Flags().StringVar(&client.AccessKey, "access-key", "", "MinIO access key")
	cmd.Flags().StringVar(&client.SecretKey, "secret-key", "", "MinIO secret key")
	cmd.Flags().BoolVar(&client.UseSSL, "use-ssl", false, "Use SSL for MinIO connection")
}

// CheckConnection probes the MinIO health/live endpoint.
// This does not require credentials and works with any recent MinIO version.
func (c *MinioClient) CheckConnection() error {
	scheme := "http"
	if c.UseSSL {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s:%d/minio/health/live", scheme, c.Host, c.Port)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
