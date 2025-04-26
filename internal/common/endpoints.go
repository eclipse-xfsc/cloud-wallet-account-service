package common

import (
	"context"
	"errors"
	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"net/http"
	"strings"
)

func ConstructResponse(f EndpointHandler, e Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := f(c, e)
		if err != nil {
			if c.Writer.Status() != http.StatusOK {
				return
			} else if er, ok := err.(*ErrorResp); ok {
				logger.Error(er, "")
				c.JSON(er.Code, ServerErrorResponse{
					Error: er.Error(),
				})
			} else {
				logger.Error(err, "Internal server error")
				c.JSON(http.StatusInternalServerError, ServerErrorResponse{
					Error: err.Error(),
				})
			}
		} else {
			c.JSON(http.StatusOK, data)
		}
	}
}

func ErrorResponseBadRequest(c *gin.Context, err string, exception error) error {
	return ErrorResponse(c, http.StatusBadRequest, err, exception)
}

func ErrorResponse(c *gin.Context, code int, err string, exception error) error {
	logger.Error(nil, err)
	if exception != nil {
		logger.Error(exception, "detailed error")
	}
	c.JSON(code, gin.H{
		"message": err,
	})
	return errors.New(err)
}

func GetUserFromContext(ctx *gin.Context) (*UserInfo, error) {
	if user, ok := ctx.Request.Context().Value(UserKey).(*UserInfo); ok {
		return user, nil
	} else if CheckSkipEndpointAuth(ctx) {
		id := ctx.Param("id")
		return &UserInfo{&gocloak.UserInfo{Sub: &id}}, nil
	} else {
		return nil, ErrorResponseBadRequest(ctx, "cannot extract user data from request context", nil)
	}
}

func CheckSkipEndpointAuth(c *gin.Context) bool {
	logger.Debug("CheckSkipEndpointAuth", "exc", config.ServerConfiguration.KeyCloak.ExcludeEndpoints, "path", c.FullPath())
	for _, excl := range strings.Split(config.ServerConfiguration.KeyCloak.ExcludeEndpoints, ",") {
		if excl != "" && strings.Contains(c.FullPath(), excl) {
			return true
		}
	}
	return false
}

func GetContextWithUserId(ctx *gin.Context, id string) *gin.Context {
	user := &UserInfo{&gocloak.UserInfo{Sub: &id}}
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), UserKey, user))
	return ctx
}
