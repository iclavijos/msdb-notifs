package service

import (
	"dao"
	"log"
	"model"
)

func ProcessNotification(sessionData *model.SessionData, minutes int) {

	users := dao.GetUsersSuscribedToSeries(sessionData.SeriesId, minutes)

	for _, user := range users {
		//Send notification to user
		log.Print("Sending notification to user ")
		log.Println(user.Email)

		//TODO: Implement proper handling of notifications by sending events through Kafka
		//That will also work as a nice integration PoC
	}
}
