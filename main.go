package main

import (
	"authorization"
	"encoding/json"
	"log"
	"net/http"
	"service"
	"time"

	"dao"
	"model"

	"github.com/gorilla/mux"
)

// our main function
func main() {

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

				go func(sessData model.SessionData) {
					<-timer30Min.C
					processTimeout(sessData, 30)
					timer30Min.Stop()
				}(sessionData)
				go func(sessData model.SessionData) {
					<-timer1Hour.C
					processTimeout(sessData, 60)
					timer1Hour.Stop()
				}(sessionData)
				go func(sessData model.SessionData) {
					<-timer2Hour.C
					processTimeout(sessData, 120)
					timer2Hour.Stop()
				}(sessionData)
				go func(sessData model.SessionData) {
					<-timer12Hour.C
					processTimeout(sessData, 720)
					timer12Hour.Stop()
				}(sessionData)
				go func(sessData model.SessionData) {
					<-timer24Hour.C
					processTimeout(sessData, 1440)
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

func processTimeout(sessionData model.SessionData, minutes int) {
	service.ProcessNotification(&sessionData, minutes)
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
