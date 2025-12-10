package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Hyphen/go-sdk"
)

func main() {
	// Create a NetInfo client using functional options
	netInfo, err := hyphen.NewNetInfo(
		hyphen.WithAPIKey("your_api_key"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get info for a single IP
	fmt.Println("Getting info for single IP address...")
	ipInfo, err := netInfo.GetIPInfo(ctx, "8.8.8.8")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("IP: %s\n", ipInfo.IP)
	fmt.Printf("Type: %s\n", ipInfo.Type)
	fmt.Printf("Country: %s\n", ipInfo.Location.Country)
	fmt.Printf("City: %s\n", ipInfo.Location.City)
	fmt.Printf("Coordinates: %.4f, %.4f\n", ipInfo.Location.Lat, ipInfo.Location.Lng)

	// Get info for multiple IPs
	fmt.Println("\nGetting info for multiple IP addresses...")
	ips := []string{"8.8.8.8", "1.1.1.1"}
	ipInfos, err := netInfo.GetIPInfos(ctx, ips)
	if err != nil {
		log.Fatal(err)
	}

	for i, info := range ipInfos {
		fmt.Printf("\nIP %d: %+v\n", i+1, info)
	}
}
