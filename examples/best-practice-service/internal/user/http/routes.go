package userhttp

import "github.com/gin-gonic/gin"

func Routes(r *gin.Engine) {
	group := r.Group("/users")

	group.POST("", Register)
}
