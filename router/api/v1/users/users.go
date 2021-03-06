package users

import (
	"github.com/TechMinerApps/upmaster/models"
	"github.com/gin-gonic/gin"
)

/*
	usersAPI.GET("", users.Index)
	usersAPI.PUT("/:username", users.Update)
	usersAPI.DELETE("/:username", users.Destroy)
*/

// swagger:response IndexUserResponse
type IndexUserResponse []models.User

// swagger:parameters DestroyUserRequest
type DestroyUserRequest struct {
	// in:body
	Password string `json:"password"`
}

// Index list all users
func Index(c *gin.Context) {
	// swagger:route GET /users User IndexUser
	//		list all users, FOR ADMIN ONLY
	// responses:
	//		200: IndexUserResponse
}

// Update update user info
func Update(c *gin.Context) {
	// swagger:route PUT /users/{username} User UpdateUser
	//		update user info
	// responses:
	//		200: UserResponse
}

// Destroy delete user
func Destroy(c *gin.Context) {
	// swagger:route DELETE /users/{username} User DestroyUser
	//		destroy user by admin or user himself
	// responses:
	//		204: DestroyUserRequest
}
