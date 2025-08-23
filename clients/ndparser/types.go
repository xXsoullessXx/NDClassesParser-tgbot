package ndparser

// Class represents information about a class
type Class struct {
	CRN   string `json:"crn"`
	Title string `json:"title"`
	Seats int    `json:"seats"`
}
