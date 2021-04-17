package main

import (
  "log"

  "github.com/ABFranco/roomz-media-server/rms"

  "github.com/gin-gonic/gin"
)

func GinMiddleware(allowOrigin string) gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
    c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
    c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

    if c.Request.Method == "OPTIONS" {
      c.AbortWithStatus(204)
      return
    }

    c.Request.Header.Del("Origin")

    c.Next()
  }
}

func main() {
  router := gin.New()
  server := &rms.RoomzMediaServer{}
  server = server.Init()

  go server.SioServer.Serve()
  defer server.SioServer.Close()

  router.Use(GinMiddleware("http://localhost:3000"))
  router.GET("/socket.io/*any", gin.WrapH(server.SioServer))
  router.POST("/socket.io/*any", gin.WrapH(server.SioServer))

  log.Fatal(router.Run(":5000"))
}