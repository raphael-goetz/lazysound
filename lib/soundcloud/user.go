package api

type User struct {
	// User identifier
	ID int `json:"id"`
	// User identifier
	URN string `json:"urn"`
	// Username title
	Username string `json:"username"`
}
