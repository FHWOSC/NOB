package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
)

const CustomSPlanURLFormat = "https://intern.fh-wedel.de/~splan/index.html?typ=benutzer_vz_ausgabe&id=%s" //1. Param: id

const (
	DefDbHost          = "localhost"
	DefDbPort          = "3306"
	DefDbUser          = "root"
	DefDbPass          = "mypassword"
	DefDatabase        = "stundenplan"
	DbConnectionFormat = "%s:%s@tcp(%s:%s)/%s"
)

func ConnectTo(host, port, user, pass string) (*sql.DB, error) {
	connection := fmt.Sprintf(DbConnectionFormat+"?parseTime=true",
		user,
		pass,
		host,
		port,
		DefDatabase,
	)

	return sql.Open("mysql", connection)
}

func sanitize(str string) string {
	str = strings.Replace(str, "&nbsp;", " ", -1)
	str = strings.Replace(str, " ", " ", -1)
	return str
}

func main() {
	log.Println("started...")

	// Load the HTML document of the alphabetical overview site
	doc, err := GetDoc("https://intern.fh-wedel.de/~splan/index.html?typ=benutzer_vz")
	if err != nil {
		log.Fatalln(err)
	}

	// connect to database
	db, err := ConnectTo(DefDbHost, DefDbPort, DefDbUser, DefDbPass)
	if err != nil {
		log.Fatalln(err)
	}

	// find all employees
	for i, lecturer := range GetLecturers() {
		fmt.Printf("%2d. %s\n", i, lecturer)
		// TODO DB add type attribute
		_, err := db.Exec(`
			INSERT INTO Employee(id, name)
			VALUES (?, ?)
		`, lecturer.Short, strings.ToValidUTF8(lecturer.Full, "[?]"))
		if err != nil {
			log.Println(err)
		}
	}

	fmt.Println("========================================================================")

	// find all rooms
	for i, room := range GetRooms() {
		fmt.Printf("%2d. %s\n", i, room)

		_, err := db.Exec(`
			INSERT INTO Room(id, name) 
			VALUES (?, ?)
		`, room.Short, strings.ToValidUTF8(room.Full, "?"))
		if err != nil {
			log.Fatalln(err)
		}
	}

	// find all module table rows
	doc.Find(`tr[style="text-align:left;"]`).Each(func(i int, s *goquery.Selection) {
		// For each row, get the cells
		tds := make([]*goquery.Selection, 0)
		s.Find("td").Each(func(i int, selection *goquery.Selection) {
			tds = append(tds, selection)
		})

		splanId, _ := tds[0].Find("input").First().Attr("value")
		moduleName := sanitize(tds[1].Text())
		timetable := GetTimeslots(splanId)

		fmt.Println("========================================================================")
		fmt.Println("Id:\t\t", splanId)
		fmt.Println("Verantstaltung:\t", tds[1].Text())
		fmt.Println("Mitarbeiter:\t", tds[2].Text())
		fmt.Println("Link:\t\t", fmt.Sprintf(CustomSPlanURLFormat, splanId))

		r, err := db.Exec(`
			INSERT INTO Module(splanId, name) 
			VALUES (?, ?)
		`, splanId, moduleName)
		if err != nil {
			log.Fatalln(err)
		}

		id, err := r.LastInsertId()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("LastInsertId", id)

		for day, v := range timetable {
			for _, timeslot := range v {
				if day != "VV" {
					// Normaler Timetable
					_, err := db.Exec(`
					INSERT INTO Lecture(start, end, day, moduleId) 
					VALUES (?, ?, ?, ?)
				`, timeslot.Start.Add(24*400*time.Hour), timeslot.End.Add(24*366*time.Hour), day, id)
					if err != nil {
						log.Println(err)
						continue
					}

					_, err = db.Exec(`
						INSERT INTO Lecture_in_Room(roomId, lectureId)
						VALUES (?, ?)
					`, timeslot.Room, id)
					if err != nil {
						log.Println(err)
					}
				} else {
					// Veranstaltung nach Vereinbaarung
					_, err := db.Exec(`
						INSERT INTO Lecture(day, moduleId) 
						VALUES ('VV', ?)
					`, id)
					if err != nil {
						log.Println(err)
						continue
					}
				}

				for _, employee := range timeslot.Employees {
					_, err = db.Exec(`
						INSERT INTO Lecture_by_Employee(employeeId, lectureId)
						VALUES (?, ?)
					`, employee, id)
				}
			}

		}

		str, _ := json.MarshalIndent(timetable, "", " ")
		fmt.Println(string(str))

	})

	db.Close()
	log.Println("finished")
	wait := make(chan int, 1)
	<-wait
}

func GetDoc(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status))
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

type Room struct {
	Short string
	Full  string
}

func GetRooms() []Room {
	doc, err := GetDoc(`https://intern.fh-wedel.de/~splan/index.html?layout=0`)
	if err != nil {
		return nil
	}

	sources := make([]*goquery.Selection, 0)
	doc.Find(`div[style="margin-left:40px; margin-bottom:15px;"]`).Each(func(i int, s *goquery.Selection) {
		sources = append(sources, s)
	})

	rooms := make([]Room, 0)
	sources[4].Find("a").Each(func(i int, s *goquery.Selection) {
		rooms = append(rooms, Room{
			Full:  s.AttrOr("title", ""),
			Short: s.Text(),
		})
	})

	return rooms
}

type Lecturer struct {
	Type  string `json:"type"`
	Short string `json:"short"`
	Full  string `json:"full"`
}

func GetLecturers() []Lecturer {
	doc, err := GetDoc("https://intern.fh-wedel.de/~splan/index.html?layout=0")
	if err != nil {
		return nil
	}

	sources := make([]*goquery.Selection, 0)
	doc.Find(`div[style="margin-left:40px;"]`).Each(func(i int, s *goquery.Selection) {
		sources = append(sources, s)
	})

	lecturers := make([]Lecturer, 0)

	sources[2].Find("a").Each(func(i int, s *goquery.Selection) {
		lecturers = append(lecturers, Lecturer{
			Type:  "Dozent",
			Full:  s.AttrOr("title", ""),
			Short: s.Text(),
		})
	})
	sources[3].Find("a").Each(func(i int, s *goquery.Selection) {
		lecturers = append(lecturers, Lecturer{
			Type:  "Assistent",
			Full:  s.AttrOr("title", ""),
			Short: s.Text(),
		})
	})
	sources[4].Find("a").Each(func(i int, s *goquery.Selection) {
		lecturers = append(lecturers, Lecturer{
			Type:  "Extern",
			Full:  s.AttrOr("title", ""),
			Short: s.Text(),
		})
	})

	return lecturers
}

type Timeslot struct {
	Subject   string    `json:"subject"`
	Employees []string  `json:"lecturer"`
	Room      string    `json:"room"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
}

func GetTimeslots(id string) map[string][]Timeslot {
	doc, err := GetDoc(fmt.Sprintf(CustomSPlanURLFormat, id))
	if err != nil {
		return nil
	}

	rows := make([]*goquery.Selection, 0)
	doc.Find("table[style=\"font-size: 8pt; text-align:center;\"] > tbody > tr").Each(func(i int, selection *goquery.Selection) {
		rows = append(rows, selection)
	})

	timetable := make(map[string][]Timeslot)

	timeslots := make([]Timeslot, 0)
	rows[0].Find("td[style=\"width:14%;\"]").Each(func(i int, selection *goquery.Selection) {
		parts := strings.Split(selection.Text(), " - ")
		s, err := time.Parse("15:04", parts[0])
		if err != nil {
			return
		}
		e, err := time.Parse("15:04", parts[1])
		if err != nil {
			return
		}

		timeslots = append(timeslots, Timeslot{
			Start: s,
			End:   e,
		})
	})

	for _, row := range rows[1:] {
		first := true

		day := ""
		row.Children().Each(func(i int, selection *goquery.Selection) {
			if selection.Find("table").Size() <= 0 {
				if first {
					day = selection.Text()
					first = false
				} else {
				}
			} else {
				span, err := strconv.ParseInt(selection.AttrOr("colspan", "1"), 10, 32)
				if err == nil {
					event := selection.Find("td.splan_veranstaltung > a").First().Text()
					event = strings.Trim(event, " ")

					start := timeslots[i-1].Start
					end := timeslots[i-2+int(span)].End

					room := selection.Find("td.splan_hoerer_raum > a").First().Text()

					employees := make([]string, 0)
					selection.Find("td.splan_mitarbeiter > a").Each(func(i int, s *goquery.Selection) {
						employees = append(employees, s.Text())
					})

					if timetable[day] == nil {
						timetable[day] = make([]Timeslot, 0)
					}

					timetable[day] = append(timetable[day], Timeslot{
						Subject:   event,
						Employees: employees,
						Room:      room,
						Start:     start,
						End:       end,
					})

				}
			}
		})
	}

	doc.Find("table[style=\"font-size: 8pt; text-align:center; width:100%;\"] > tbody > tr").
		First().
		Each(func(i int, selection *goquery.Selection) {
			event := selection.Find("td.splan_veranstaltung > a").First().Text()
			event = strings.Trim(event, " ")

			room := selection.Find("td.splan_hoerer_raum > a").First().Text()

			employees := make([]string, 0)
			selection.Find("td.splan_mitarbeiter > a").Each(func(i int, s *goquery.Selection) {
				employees = append(employees, s.Text())
			})

			timetable["VV"] = []Timeslot{{
				Subject:   event,
				Employees: employees,
				Room:      room,
			}}
		})

	return timetable
}
