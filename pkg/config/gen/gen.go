package main

import (
	cfg "github.com/conductorone/baton-googleads/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("googleads", cfg.Config)
}
