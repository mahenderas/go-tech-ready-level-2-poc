package consul

import (
	"fmt"
	"log"
	"os"
	consulapi "github.com/hashicorp/consul/api"
)

func RegisterWithConsul(serviceName string, servicePort int) {
	consulAddr := os.Getenv("SERVICE_DISCOVERY")
	if consulAddr == "" {
		consulAddr = "localhost:8500"
	}
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Printf("Consul client error: %v", err)
		return
	}
	var registration *consulapi.AgentServiceRegistration
	if os.Getenv("DEPLOY_ENV") == "gcp" {
		registration = &consulapi.AgentServiceRegistration{
			Name:    serviceName,
			Address: serviceName,
			Port:    servicePort,
		}
	} else {
		registration = &consulapi.AgentServiceRegistration{
			Name:    serviceName,
			Address: serviceName,
			Port:    servicePort,
			Check: &consulapi.AgentServiceCheck{
				HTTP:     fmt.Sprintf("http://%s:%d/health", serviceName, servicePort),
				Interval: "10s",
				Timeout:  "1s",
			},
		}
	}
	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Printf("Consul registration failed: %v", err)
	} else {
		log.Printf("Registered with Consul: %s", serviceName)
	}
}
