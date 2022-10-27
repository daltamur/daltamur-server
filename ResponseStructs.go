package main

type SongRange struct {
	AllSongs map[string]DaySongs `json:"AllSongs"`
}

type DaySongs struct {
	Songs []SongData `json:"songs"`
}

type Status struct {
	Table       string `json:"table"`
	RecordCount int64  `json:"recordCount"`
	Time        string `json:"time"`
}

type Image struct {
	Size string `json:"size"`
	Text string `json:"#text"`
}

type Songs struct {
	Error   string     `json:"error"`
	Songs   []SongData `json:"songs"`
	CurPage string     `json:"curPage"`
}

type SongData struct {
	Artist       string  `json:"artist"`
	Album        string  `json:"album"`
	ArtistImages []Image `json:"artistImages"`
	AlbumImages  []Image `json:"albumImages"`
	Date         string  `json:"date"`
	Name         string  `json:"name"`
	Time         string  `json:"time"`
	UTS          int     `json:"UTS"`
}
