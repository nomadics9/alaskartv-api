package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

func getVersion(Name string) string {
	var repoURL string

	// Determine the correct API endpoint
	switch Name {
	case "alaskarfin":
		repoURL = "https://api.github.com/repos/jellyfin/jellyfin/releases/latest"
	case "alaskarseer":
		repoURL = "https://api.github.com/repos/Fallenbagel/jellyseerr/releases/latest"
	case "alaskartv":
		repoURL = "https://api.github.com/repos/jellyfin/jellyfin-androidtv/releases/latest"
	default:
		fmt.Println("Invalid repository name")
		return "Error"
	}

	var result map[string]interface{}
	_, err := resty.New().R().SetResult(&result).Get(repoURL)
	if err != nil {
		fmt.Println("Error fetching version:", err)
		return "Error"
	}

	version, ok := result["tag_name"].(string)
	if !ok {
		fmt.Println("Error parsing version")
		return "Error"
	}

	return version
}
