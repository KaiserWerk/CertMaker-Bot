package assets

import "embed"

//go:embed config/*
var configFS embed.FS

func GetConfigurationAssets() *embed.FS {
	return &configFS
}
