package main

import "time"

type DeviceSoftwareDefinition struct {
	Manifest Manifest `json:"devicesoftwaredefinition"`
}

type Manifest struct {
	Description string              `json:"description"`
	Packages    map[string]Packages `json:"packages"`
}

type Packages struct {
	ImageName string `json:"imageName"`
	Version   string `json:"version"`
}

type SecurePackageDownloadTokenRequest struct {
	PackageName string `json:"packageName"`
	DeviceID    string `json:"deviceId"`
}

type SecurePackageDownloadTokenResponse struct {
	Url         string `json:"url"`
	PackageName string `json:"packageName"`
	DeviceID    string `json:"deviceId"`
	DlToken     string `json:"dlToken"`
}

//https://github.com/cloudevents/spec/blob/v1.0/spec.md#required-attributes
type CloudEvent struct {
	ID          string      `json:"id"`
	Source      string      `json:"source"`
	SpecVersion string      `json:"specversion"`
	Type        string      `json:"type"`
	Time        time.Time   `json:"time"`
	Data        interface{} `json:"data"`
}
