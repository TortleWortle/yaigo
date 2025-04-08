package db

type User struct {
	Username string `json:"username"`
	Age      int    `json:"age"`
}

type Group struct {
	Name string `json:"name"`
}
