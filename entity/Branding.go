package entity

type VideoBranding struct {
	Titles []Title `json:"titles,omitempty"`
}

type Title struct {
	Title    string `json:"title"`
	Original bool   `json:"original"`
	Votes    int    `json:"votes"`
	Locked   bool   `json:"locked"`
	UUID     string `json:"uuid"`
}
