package model

type Link struct {
	Short    string `json:"short"`
	Long     string `json:"long"`
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
}
