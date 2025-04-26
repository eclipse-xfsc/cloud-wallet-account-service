package credentials

import (
	"github.com/gin-gonic/gin"
)

type UserService struct {
}

func (t *UserService) AddRoutes(group *gin.RouterGroup) error {

	userGroup := group.Group("/devices")

	userGroup.GET("/list", func(ctx *gin.Context) {
		GetUser(ctx)
	})

	userGroup.DELETE("/:id", func(ctx *gin.Context) {
		DeleteUser(ctx)
	})

	userGroup.PUT("/:id", func(ctx *gin.Context) {
		CreateUser(ctx)
	})

	return nil
}

func GetUser(ctx *gin.Context) {

}

func DeleteUser(ctx *gin.Context) {

}

func CreateUser(ctx *gin.Context) {

}
