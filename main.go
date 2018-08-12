package main

import (
	"encoding/json"
	"log"
	"net/http"
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
				notifInfo := sessionData.EventName + " - " + sessionData.SessionName + "(" + sessionData.SessionStartTime.String() + ")"
				go func() {
					<-timer30Min.C
					log.Print("Aviso 30 min: " + notifInfo)
					timer30Min.Stop()
				}()
				go func() {
					<-timer1Hour.C
					log.Print("Aviso 1 hora: " + notifInfo)
					timer1Hour.Stop()
				}()
				go func() {
					<-timer2Hour.C
					log.Print("Aviso 2 horas: " + notifInfo)
					timer2Hour.Stop()
				}()
				go func() {
					<-timer12Hour.C
					log.Print("Aviso 12 horas: " + notifInfo)
					timer12Hour.Stop()
				}()
				go func() {
					<-timer24Hour.C
					log.Print("Aviso 24 horas: " + notifInfo)
					timer24Hour.Stop()
				}()
			}(session)
		}
	}()

	router := mux.NewRouter()
	router.HandleFunc("/suscriptions/series", getSeries).Methods("GET")
	router.HandleFunc("/suscriptions/{id}", getUserSuscriptions).Methods("GET")
	router.HandleFunc("/suscriptions", updateUserSuscriptions).Methods("POST")
	//router.HandleFunc("/people/{id}", GetPerson).Methods("GET")
	//router.HandleFunc("/people/{id}", CreatePerson).Methods("POST")
	//router.HandleFunc("/people/{id}", DeletePerson).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8000", router))

	//start := time.Date(2018, 8, 11, 10, 15, 0, 0, time.UTC)
}

func getUpcomingSessions(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, dao.GetUpcomingSessions())
}

func getSeries(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, dao.GetSeries())
}

func getUserSuscriptions(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	suscriptions := dao.GetUserSuscriptions(params["id"])

	respondWithJson(w, http.StatusOK, suscriptions)
}

func updateUserSuscriptions(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var userSuscs model.User
	if err := json.NewDecoder(r.Body).Decode(&userSuscs); err != nil {
		log.Fatalln(err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	dao.UpdateUserSuscriptions(userSuscs)
	//if err := dao.Insert(movie); err != nil {
	//	respondWithError(w, http.StatusInternalServerError, err.Error())
	//	return
	//}
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
