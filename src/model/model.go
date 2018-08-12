package model

import "time"

type User struct {
	Id           int            `json:"id"`
	Email        string         `json:"email"`
	Suscriptions []*Suscription `json:"suscriptions"`
}

type Suscription struct {
	UserId        int  `json:"userId" db:"user_id"`
	SeriesId      int  `json:"seriesId" db:"series_id"`
	SessionsNotif bool `json:"sessionsNotifications" db:"sessions_suscription"`
	ResultsNotif  bool `json:"resultsNotifications" db:"results_suscription"`
	HoursNotif    int  `json:"hoursNotification" db:"hours_notification"`
}

type Series struct {
	Id   int    `json:"id"`
	Name string `json:"seriesName"`
}

type SessionData struct {
	SeriesId         int       `json:"seriesId" db:"series_id"`
	SeriesName       string    `json:"seriesName" db:"series_name"`
	EventName        string    `json:"eventName" db:"event_name"`
	SessionName      string    `json:"sessionName" db:"session_name"`
	SessionStartTime time.Time `json:"sessionStartTime" db:"session_start_time"`
}
