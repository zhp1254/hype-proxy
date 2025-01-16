package routers

import (
	"github.com/gin-gonic/gin"
	"hype-proxy/httpserver/api"
	"net/http"
)

func Init() *gin.Engine {
	// 修改模式
	gin.SetMode(gin.ReleaseMode)
	// 创建路由
	engine := gin.New()
	engine.Use(gin.Recovery())
	// 404默认值
	engine.NoRoute(NoResponse)
	// MiddleWare中间件-解决跨域
	//engine.Use(middlewares.Cors())
	// 设置请求映射
	initRouter(engine)
	return engine
}

func NoResponse(c *gin.Context) {
	// 返回404状态码
	c.JSON(http.StatusNotFound, gin.H{
		"status": 404,
		"error":  "404, page not exists!",
	})
}

func initRouter(router *gin.Engine) {
	router.GET("/", api.Index)
	router.POST("/info", api.ProxyHypeInfo)
	router.POST("/exchange", api.ProxyHypeExchange)
	router.POST("/explorer", api.ProxyHypeExplorer)
	// 获取缓存最新块高
	router.POST("/getBestBlock", api.GetBestBlock)
	// 获取缓存交易
	router.POST("/getBlock", api.GetBlock)
}
