package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/websockets"
)

func WsThread(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	vars := mux.Vars(r)
	idThread, err := strconv.Atoi(vars["id"])

	if err != nil {
		panic(err)
	}

	sQB := orm.GetSelectQueryBuilder(models.NewThread()).
		Where(orm.And(orm.H{"IdThread": idThread}))
	err = sQB.ApplyQueryRow()

	if err != nil {
		panic(err)
	}

	targetThread := sQB.EffectiveModel.(*models.Thread)

	if targetThread.InitiatorThreadIdUser != user.IdUser && targetThread.ReceiverThreadIdUser != user.IdUser {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(nil)
		return
	}

	var connexion *websocket.Conn
	connexion, err = config.WebsocketUpgrade.Upgrade(w, r, nil)

	if err != nil {
		panic(err)
	}

	sQB.Consider("Messages").
		Consider("ReceiverThread").
		Consider("InitiatorThread").
		Where(orm.And(orm.H{"IdThread": targetThread.IdThread}))

	if user.IdUser == targetThread.InitiatorThreadIdUser {
		sQB.Select([]string{"*", "InitiatorThread.*", "Messages.*",
			"ReceiverThread.IdUser",
			"ReceiverThread.CreatedAt", "ReceiverThread.Pseudo",
			"ReceiverThread.PhotoPath"})
	} else {
		sQB.Select([]string{"*", "ReceiverThread.*", "Messages.*",
			"InitiatorThread.IdUser",
			"InitiatorThread.CreatedAt", "InitiatorThread.Pseudo",
			"InitiatorThread.PhotoPath"})
	}

	sQB.Aggregate(orm.H{"COUNT": "Messages.IdMessage"}).
		OrderBy(orm.H{"Messages.CreatedAt": "ASC"})
	err = sQB.ApplyQuery()

	if err != nil {
		_ = connexion.Close()
		panic(err)
	}

	targetThread = sQB.EffectiveModel.(*models.Thread)

	threadHub := websockets.GetThreadHub(targetThread)
	client := &websockets.ThreadClient{IdUser: user.IdUser, ThreadHub: threadHub, Conn: connexion}
	client.ThreadHub.Register <- client

	wsEnterSend := orm.H{"thread": targetThread, "aggregates": sQB.EffectiveAggregates}
	err = client.Conn.WriteJSON(wsEnterSend)

	if err != nil {
		_ = connexion.Close()
		panic(err)
	}

	go client.WritePump()
	go client.ReadPump()

	return
}
