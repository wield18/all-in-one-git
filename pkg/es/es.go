package es

import (
	"context"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v9"
)

type ESConfigSingle struct {
	Address string `yaml:"Address"`
}

func NewESClient(esConfig ESConfigSingle) *elasticsearch.TypedClient {
	// 创建客户端
	client, err := elasticsearch.NewTyped(elasticsearch.WithAddresses(esConfig.Address))
	if err != nil {
		log.Fatalf("创建客户端失败: %s", err)
	}

	getInfo(client)
	// getHealthInfo(client)
	return client
}

func getInfo(client *elasticsearch.TypedClient) {
	// 获取集群信息
	info := client.Info()
	res, err := info.Do(context.Background())
	if err != nil {
		log.Fatalf("获取集群信息失败: %s", err)
	}

	fmt.Println("📊 es 集群信息:")
	fmt.Println(res.ClusterName)
	fmt.Println(res.Version)
}

func getHealthInfo(client *elasticsearch.Client) {
	// 获取集群健康状态
	res, err := client.Cluster.Health(
		client.Cluster.Health.WithLevel("indices"),
	)
	if err != nil {
		log.Fatalf("获取健康状态失败: %s", err)
	}
	defer res.Body.Close()

	fmt.Println("❤️ es 集群健康状态:")
	fmt.Println(res.String())
}
