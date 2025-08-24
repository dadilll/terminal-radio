package ui

import (
	"fmt"
	"radio/internal/client"
	"strings"
)

type StationItem struct {
	Station  client.Station
	Playing  bool
	Favorite bool
	TitleStr string
}

func (i StationItem) Title() string {
	title := i.Station.Name
	if i.Playing {
		title = "🎵 " + title
	}
	if i.Favorite {
		title = "★ " + title
	}
	return truncate(title, 30)
}

func (i StationItem) Description() string {
	tags := colorTags(strings.Split(i.Station.Tags, ","))
	return fmt.Sprintf("%s • %dkbps • %s", i.Station.Country, i.Station.Bitrate, tags)
}

func (i StationItem) FilterValue() string {
	return i.Station.Name + " " + i.Station.Country + " " + i.Station.Tags + " " + i.Station.Language
}

type SortMode int

const (
	SortByName SortMode = iota
	SortByBitrate
	SortByCountry
)
