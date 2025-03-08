package entity

type SkipSegment struct {
	VideoID  string    `json:"video_id"`
	Segments []Segment `json:"segments"`
}

type Segment struct {
	Category      string    `json:"category"`
	ActionType    string    `json:"actionType"`
	Segment       []float32 `json:"segment"`
	UUID          string    `json:"UUID"`
	VideoDuration string    `json:"videoDuration"`
	Locked        int       `json:"locked"`
	Votes         int       `json:"votes"`
	Description   string    `json:"description"`
}
