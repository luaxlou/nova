package userhttp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/examples/best-practice-service/internal/user"
)

func Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := user.Register(
		c.Request.Context(),
		user.RegisterCommand{
			Email: req.Email,
			Name:  req.Name,
		},
	)
	if err != nil {
		writeRegisterError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toUserResponse(result))
}

func writeRegisterError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, user.ErrEmailExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, user.ErrInvalidEmail), errors.Is(err, user.ErrInvalidName):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "register user failed"})
	}
}
