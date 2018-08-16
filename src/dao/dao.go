package dao

import (
	"errors"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"model"
)

func getConn() *sqlx.DB {

	db, err := sqlx.Connect("mysql", "msdb:msdb@(localhost:3306)/msdb?parseTime=true&loc=Local")
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	return db
}

func GetSeries() []model.Series {
	series := []model.Series{}

	getConn().Select(&series, "SELECT id, name FROM series ORDER BY name ASC")

	return series
}

func GetUserSuscriptions(userId int) []model.Suscription {
	suscriptions := []model.Suscription{}

	err := getConn().Select(&suscriptions, "SELECT * FROM suscription WHERE user_id = ?", userId)
	if err != nil {
		log.Fatal(err)
	}

	return suscriptions
}

func GetUsersSuscribedToSeries(seriesId int, minutes int) []model.User {
	usersToBeNotified := []model.User{}

	err := getConn().Select(&usersToBeNotified,
		`select 
			user.id id, user.email email
		from
			suscription s left join jhi_user user on s.user_id = user.id
		where 
			s.series_id = ? and s.minutes_notification = ?`, seriesId, minutes)

	if err != nil {
		log.Fatal(err)
	}

	return usersToBeNotified
}

func GetUpcomingSessions() []model.SessionData {
	futureSessions := []model.SessionData{}

	err := getConn().Select(&futureSessions,
		`select 
			s.id series_id, se.edition_name series_name, 
			ee.long_event_name event_name, es.name session_name, es.id session_id, 
			es.session_start_time session_start_time 
		from 
			event_session es left join event_edition ee on es.event_edition_id = ee.id 
			left join events_series evse on ee.id = evse.event_id 
			left join series_edition se on evse.series_id = se.id 
			left join series s on s.id = se.series_id 
		where 
			es.session_start_time > now()`)
	if err != nil {
		log.Fatal(err)
	}

	return futureSessions
}

func GetUsers() []model.User {
	users := []model.User{}

	getConn().Select(&users, "SELECT id, email FROM jhi_user")

	return users
}

func GetUserByUsername(username string) (*model.User, error) {
	user := []model.User{}

	err := getConn().Select(&user, "SELECT id, email FROM jhi_user WHERE login = ?", username)

	if err != nil || len(user) != 1 {
		log.Printf("Error retrieving user using username %v", err)
		return nil, errors.New("Invalid username")
	}
	return &user[0], nil
}

func UpdateUserSuscriptions(user model.User) model.User {
	tx, _ := getConn().Begin()

	_, err := tx.Exec("DELETE FROM suscription WHERE user_id = ?", user.Id)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	sqlStr := "INSERT INTO suscription(user_id, series_id, sessions_suscription, results_suscription, minutes_notification) VALUES "
	const rowSQL = "(?, ?, ?, ?, ?)"
	var inserts []string
	var vals = []interface{}{}
	for _, row := range user.Suscriptions {
		inserts = append(inserts, rowSQL)
		if row.MinutesNotif == 0 {
			row.MinutesNotif = 60
		}
		vals = append(vals, user.Id, row.SeriesId, row.SessionsNotif, row.ResultsNotif, row.MinutesNotif)
	}
	sqlStr = sqlStr + strings.Join(inserts, ",")

	txStmt, err := tx.Prepare(sqlStr)
	_, err = txStmt.Exec(vals...)
	if err != nil {
		tx.Rollback()
		log.Println(err)
	}
	_ = tx.Commit()

	return user
}
