package middleware

import (
	"context"
	"net/http"

	"github.com/tracewayapp/traceway/backend/app/db"

	"github.com/gin-gonic/gin"
)

func Transactional(c *gin.Context) {
	txHandle, err := db.DB.Begin()

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		panic(err)
	}

	defer func() {
		if r := recover(); r != nil {
			txHandle.Rollback()
			c.AbortWithStatus(http.StatusInternalServerError)
			panic(r)
		}
	}()

	c.Set(db.TransactionContextKey, txHandle)

	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), db.TransactionContextKey, txHandle))

	c.Next()

	if status := c.Writer.Status(); status >= 200 && status < 400 {
		if err := txHandle.Commit(); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			panic(err)
		}
	} else {
		txHandle.Rollback()
	}
}
