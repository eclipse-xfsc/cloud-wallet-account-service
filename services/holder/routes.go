package holder

import "github.com/gin-gonic/gin"

type HolderService struct {
}

func (t *HolderService) AddRoutes(group *gin.RouterGroup) error {

	historyGroup := group.Group("/history")

	historyGroup.POST("/sign", func(ctx *gin.Context) {
		SignPresentation(ctx)
	})

	return nil
}

func SignPresentation(ctx *gin.Context) {
	mock := make(map[string]string)

	ctx.JSON(200, mock)
}
