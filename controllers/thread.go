package controllers

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/websockets"
)

func WsThread(ctx echo.Context) error {
	user := ctx.Get("user").(*models.User)
	idThread, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		return err
	}

	sQB := orm.GetSelectQueryBuilder(models.NewThread()).
		Where(orm.And(orm.H{"IdThread": idThread}))
	err = sQB.ApplyQueryRow()

	if err != nil {
		return err
	}

	targetThread := sQB.EffectiveModel.(*models.Thread)

	if targetThread.InitiatorThreadIdUser != user.IdUser && targetThread.ReceiverThreadIdUser != user.IdUser {
		return ctx.JSON(http.StatusUnauthorized, echo.Map{})
	}

	connexion, err := websockets.Upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)

	if err != nil {
		return err
	}

	sQB.Consider("Messages").
		Consider("ReceiverThread").
		Consider("InitiatorThread").
		Where(orm.And(orm.H{"IdThread": targetThread.IdThread}))

	if user.IdUser == targetThread.InitiatorThreadIdUser {
		sQB.Select([]string{"*", "InitiatorThread.*", "Messages.*",
			"ReceiverThread.CreatedAt", "ReceiverThread.Pseudo",
			"ReceiverThread.PhotoPath"})
	} else {
		sQB.Select([]string{"*", "ReceiverThread.*", "Messages.*",
			"InitiatorThread.IdUser",
			"InitiatorThread.CreatedAt", "InitiatorThread.Pseudo",
			"InitiatorThread.PhotoPath"})
	}

	sQB.Aggregate(orm.H{"COUNT": "Messages.IdMessage"})
	err = sQB.ApplyQuery()

	if err != nil {
		return err
	}

	targetThread = sQB.EffectiveModel.(*models.Thread)

	threadHub := websockets.GetThreadHub(targetThread)
	client := &websockets.ThreadClient{ThreadHub: threadHub, Conn: connexion}
	client.ThreadHub.Register <- client

	wsEnterSend := echo.Map{"thread": targetThread, "aggregates": sQB.EffectiveAggregates}
	err = client.Conn.WriteJSON(wsEnterSend)

	if err != nil {
		return err
	}

	go client.WritePump()
	go client.ReadPump()

	return err
}
