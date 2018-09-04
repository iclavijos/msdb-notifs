package src

import (
	"authorization"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"service"
	"sync"
	"time"

	"dao"
	"model"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
)

var sessionTimersMap = make(map[int][]*time.Timer)
var lock = sync.RWMutex{}

func main() {

	//Initialize Kafka producer object
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092"})
	if err != nil {
		panic(err)
	}

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	defer p.Close()

	go service.ProcessEventEditionEvents()

	go func() {
		upcomingSessions := dao.GetUpcomingSessions()
		now := time.Now()
		for _, session := range upcomingSessions {

			go func(sessionData model.SessionData) {
				min30Notif := sessionData.SessionStartTime.Add(time.Minute * -30).Sub(now)
				hour1Notif := sessionData.SessionStartTime.Add(time.Hour * -1).Sub(now)
				hour2Notif := sessionData.SessionStartTime.Add(time.Hour * -2).Sub(now)
				hours12Notif := sessionData.SessionStartTime.Add(time.Hour * -12).Sub(now)
				hours24Notif := sessionData.SessionStartTime.Add(time.Hour * -24).Sub(now)
				timer30Min := time.NewTimer(min30Notif)
				timer1Hour := time.NewTimer(hour1Notif)
				timer2Hour := time.NewTimer(hour2Notif)
				timer12Hour := time.NewTimer(hours12Notif)
				timer24Hour := time.NewTimer(hours24Notif)

				timers := []*time.Timer{timer30Min, timer1Hour, timer2Hour, timer12Hour, timer24Hour}
				addTimersToMap(sessionData.SessionId, timers)

				go func(sessData model.SessionData) {
					<-timer30Min.C
					processTimeout(p, sessionTimersMap, sessData, 30)
					timer30Min.Stop()
				}(sessionData)
				go func(sessData model.SessionData) {
					<-timer1Hour.C
					processTimeout(p, sessionTimersMap, sessData, 60)
					timer1Hour.Stop()
				}(sessionData)
				go func(sessData model.SessionData) {
					<-timer2Hour.C
					processTimeout(p, sessionTimersMap, sessData, 120)
					timer2Hour.Stop()
				}(sessionData)
				go func(sessData model.SessionData) {
					<-timer12Hour.C
					processTimeout(p, sessionTimersMap, sessData, 720)
					timer12Hour.Stop()
				}(sessionData)
				go func(sessData model.SessionData) {
					<-timer24Hour.C
					processTimeout(p, sessionTimersMap, sessData, 1440)
					timer24Hour.Stop()
				}(sessionData)
			}(session)
		}
	}()

	router := mux.NewRouter()
	router.HandleFunc("/suscriptions/series", getSeries).Methods("GET")
	router.HandleFunc("/suscriptions/{id}", getUserSuscriptions).Methods("GET")
	router.HandleFunc("/suscriptions", updateUserSuscriptions).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func addTimersToMap(sessionId int, timers []*time.Timer) {
	lock.Lock()
	defer lock.Unlock()
	sessionTimersMap[sessionId] = timers
}

func processTimeout(producer *kafka.Producer, sessionTimersMap map[int][]*time.Timer, sessionData model.SessionData, minutes int) {
	service.ProcessNotification(producer, &sessionData, minutes)
	if minutes == 30 {
		sessionTimersMap[sessionData.SessionId] = nil
	}
}

func getUpcomingSessions(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, dao.GetUpcomingSessions())
}

func getSeries(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, dao.GetSeries())
}

func getUserSuscriptions(w http.ResponseWriter, r *http.Request) {
	//This should be better handled at router level using libraries such as
	//http://github.com/codegangsta/negroni that allows better chaining of invocations
	authorized, username := authorization.ValidateJWT(r)
	if !authorized {
		respondWithError(w, http.StatusUnauthorized, "")
	}

	user, err := dao.GetUserByUsername(username)
	if err != nil {
		log.Println(err)
		respondWithError(w, http.StatusBadRequest, err.Error())
	} else {
		suscriptions := dao.GetUserSuscriptions(user.Id)
		respondWithJson(w, http.StatusOK, suscriptions)
	}

}

func updateUserSuscriptions(w http.ResponseWriter, r *http.Request) {
	if authorized, _ := authorization.ValidateJWT(r); !authorized {
		respondWithError(w, http.StatusUnauthorized, "")
	}
	defer r.Body.Close()
	var userSuscs model.User
	if err := json.NewDecoder(r.Body).Decode(&userSuscs); err != nil {
		log.Fatalln(err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	dao.UpdateUserSuscriptions(userSuscs)

	respondWithJson(w, http.StatusCreated, userSuscs)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
