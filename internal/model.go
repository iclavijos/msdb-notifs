package internal

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
	MinutesNotif  int  `json:"minutesNotification" db:"minutes_notification"`
}

type Series struct {
	Id   int    `json:"id"`
	Name string `json:"seriesName"`
}

type SessionData struct {
	SeriesId         int       `json:"seriesId" db:"series_id"`
	SeriesName       string    `json:"seriesName" db:"series_name"`
	EventName        string    `json:"eventName" db:"event_name"`
	SessionId        int       `json:"sessionId" db:"session_id"`
	SessionName      string    `json:"sessionName" db:"session_name"`
	SessionStartTime time.Time `json:"sessionStartTime" db:"session_start_time"`
}

type TokenAuthentication struct {
	Token string `json:"token" form:"token"`
}
