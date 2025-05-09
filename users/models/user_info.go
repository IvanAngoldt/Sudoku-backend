package models

type UserInfo struct {
	ID       int    `json:"id"`
	UserID   string `json:"user_id"`
	FullName string `json:"full_name"`
	Age      int    `json:"age"`
	City     string `json:"city"`
}

type UserInfoResponse struct {
	UserID   string `json:"user_id"`
	FullName string `json:"full_name"`
	Age      int    `json:"age"`
	City     string `json:"city"`
}

type UpdateUserInfoInput struct {
	FullName *string `json:"full_name"`
	Age      *int    `json:"age"`
	City     *string `json:"city"`
}
