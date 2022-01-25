package assets

import "embed"

//go:embed config/*
var configFS embed.FS

func GetConfigurationAssets() *embed.FS {
	return &configFS
}

func GetConfigAsset(name string) ([]byte, error) {
	return configFS.ReadFile("config/" + name)
}
