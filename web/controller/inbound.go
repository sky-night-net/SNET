package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/web/entity"
	"github.com/sky-night-net/snet/web/service"
)

type InboundController struct {
	BaseController
	inboundService *service.InboundService
}

func NewInboundController(g *gin.RouterGroup) *InboundController {
	c := &InboundController{}
	// Provide these in NewServer
	return c
}

func (c *InboundController) SetInboundService(s *service.InboundService) {
	c.inboundService = s
}

func (c *InboundController) RegisterRoutes(g *gin.RouterGroup) {
	inbound := g.Group("/inbound")
	inbound.GET("/list", c.List)
	inbound.POST("/add", c.Add)
	inbound.POST("/update/:id", c.Update)
	inbound.POST("/del/:id", c.Del)
	
	client := g.Group("/client")
	client.POST("/add", c.AddClient)
	client.POST("/update/:id", c.UpdateClient)
	client.POST("/del/:id", c.DelClient)
}

func (c *InboundController) List(ctx *gin.Context) {
	inbounds, err := c.inboundService.GetAllInbounds()
	if err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Obj: inbounds})
}

func (c *InboundController) Add(ctx *gin.Context) {
	var inbound model.Inbound
	if err := ctx.ShouldBindJSON(&inbound); err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
		return
	}
	err := c.inboundService.AddInbound(&inbound)
	if err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "Inbound added"})
}

func (c *InboundController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var inbound model.Inbound
	if err := ctx.ShouldBindJSON(&inbound); err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
		return
	}
	inbound.Id = id
	err := c.inboundService.UpdateInbound(&inbound)
	if err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "Inbound updated"})
}

func (c *InboundController) Del(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	err := c.inboundService.DelInbound(id)
	if err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "Inbound deleted"})
}

func (c *InboundController) AddClient(ctx *gin.Context) {
    var client model.Client
    if err := ctx.ShouldBindJSON(&client); err != nil {
        ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
        return
    }
    err := c.inboundService.AddClient(&client)
    if err != nil {
        ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "Client added"})
}

func (c *InboundController) UpdateClient(ctx *gin.Context) {
    id, _ := strconv.Atoi(ctx.Param("id"))
    var client model.Client
    if err := ctx.ShouldBindJSON(&client); err != nil {
        ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
        return
    }
    client.Id = id
    err := c.inboundService.UpdateClient(&client)
    if err != nil {
        ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "Client updated"})
}

func (c *InboundController) DelClient(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	err := c.inboundService.DelClient(id)
	if err != nil {
		ctx.JSON(http.StatusOK, entity.Msg{Success: false, Msg: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, entity.Msg{Success: true, Msg: "Client deleted"})
}
