package main

import (
    "time"
    "sync"
    "net/http"
	"crypto/rand"
	"encoding/hex"
	"io"
	"github.com/codegangsta/martini"
    "github.com/martini-contrib/encoder"
	"github.com/martini-contrib/binding"
)

// In-memory database
var db = DB{
    cals:   make(map[string]*Calendar),
}

// Our single martini instance
var m = martini.Classic()

func main() {
	// Middleware
    m.Use(func(c martini.Context, w http.ResponseWriter) {
        c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
    })

	// Calendar API routes
    m.Group("/calendars", func(r martini.Router) {
        r.Get("/", calendarIndex)
        r.Get("/:id", calendarShow)
        r.Post("/", binding.Bind(Calendar{}), calendarCreate)
        r.Put("/:id", binding.Bind(Calendar{}), calendarUpdate)
        r.Delete("/:id", calendarDestroy)
    })
    m.Group("/calendars/:calendar_id/entries", func(r martini.Router) {
        r.Get("/", entryIndex)
        r.Get("/:id", entryShow)
        r.Post("/", binding.Bind(Entry{}), entryCreate)
        r.Put("/:id", binding.Bind(Entry{}), entryUpdate)
        r.Delete("/:id", entryDestroy)
    })
    m.Run()
}

type DB struct {
    cals    map[string]*Calendar
    rw      sync.RWMutex
}

func (db DB) Reader(read func(db DB)) {
    db.rw.RLock()
    defer db.rw.RUnlock()
    read(db)
}

func (db DB) Writer(write func(db DB)) {
    db.rw.Lock()
    defer db.rw.Unlock()
    write(db)
}

func (db DB) makeKey() string {
	b := make([]byte, 8)
	n, err := io.ReadFull(rand.Reader, b)
	if n != len(b) || err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

type Calendar struct {
    Id      string     `json:"-"`
    Name    string     `json:"name"`
    Entries map[string]*Entry `json:"entries"`
}

func (cal Calendar) Validate(errors *binding.Errors, req *http.Request) {
	if len(cal.Name) == 0 {
		errors.Fields["name"] = "Required attribute missing"
	}
}

type Entry struct {
    Id      string     `json:"-"`
    Start   string     `json:"startTime"`
    End     string     `json:"endTime"`
    Desc    string     `json:"description"`
}

func (ent Entry) Validate(errors *binding.Errors, req *http.Request) {
	const (
		Layout = "2006-01-02 15:04:05 -0700"
		ErrorMsg = "Malformed time specification"
	)
	if len(ent.Start) != 0 {
		if _, err := time.Parse(Layout, ent.Start); err != nil {
			errors.Fields["startTime"] = ErrorMsg
		}
	}
	if len(ent.End) != 0 {
		if _, err := time.Parse(Layout, ent.End); err != nil {
			errors.Fields["endTime"] = ErrorMsg
		}
	}
}

/*
 * Utility functions
 */
func findCalendar(id string, success func(cal *Calendar) (int, []byte)) (int, []byte) {
    var (
        ok bool
        status int
        result []byte
    )
    db.Reader(func(db DB) {
        var cal *Calendar
        if cal, ok = db.cals[id]; ok {
            status, result = success(cal)
        }
    })
    if ok {
        return status, result
    }
    return http.StatusNotFound, nil
}

func findEntry(cal *Calendar, id string, success func(ent *Entry) (int, []byte)) (int, []byte) {
    if ent, ok := cal.Entries[id]; ok {
        return success(ent)
    }
    return http.StatusNotFound, nil
}

/*
 * Request handlers
 */
func calendarIndex(enc encoder.Encoder) (int, []byte) {
    var result []byte
    db.Reader(func (db DB) {
        result = encoder.Must(enc.Encode(db.cals))
    })
    return http.StatusOK, result
}

func calendarShow(p martini.Params, enc encoder.Encoder) (int, []byte) {
    return findCalendar(p["id"], func(cal *Calendar) (int, []byte) {
        return http.StatusOK, encoder.Must(enc.Encode(*cal))
    })
}

func calendarCreate(cal Calendar, enc encoder.Encoder) (int, []byte) {
    cal.Id = db.makeKey()
    cal.Entries = make(map[string]*Entry)
    db.Writer(func(db DB) {
        db.cals[cal.Id] = &cal
    })
    return http.StatusCreated, encoder.Must(enc.Encode(cal))
}

func calendarUpdate(p martini.Params, update Calendar, enc encoder.Encoder) (int, []byte) {
    return findCalendar(p["id"], func(cal *Calendar) (int, []byte) {
		if len(update.Name) != 0 {
			db.Writer(func(db DB) {
				cal.Name = update.Name
			})
		}
        return http.StatusOK, encoder.Must(enc.Encode(*cal))
    })
}

func calendarDestroy(p martini.Params, enc encoder.Encoder) (int, []byte) {
    return findCalendar(p["id"], func(cal *Calendar) (int, []byte) {
        db.Writer(func(db DB) {
            delete(db.cals, cal.Id)
        })
        return http.StatusNoContent, nil
    })
}

func entryIndex(p martini.Params, enc encoder.Encoder) (int, []byte) {
    return findCalendar(p["calendar_id"], func(cal *Calendar) (int, []byte) {
        return http.StatusOK, encoder.Must(enc.Encode(cal.Entries))
    })
}

func entryShow(p martini.Params, enc encoder.Encoder) (int, []byte) {
    return findCalendar(p["calendar_id"], func(cal *Calendar) (int, []byte) {
        return findEntry(cal, p["id"], func(ent *Entry) (int, []byte) {
            return http.StatusOK, encoder.Must(enc.Encode(*ent))
        })
    })
}

func entryCreate(p martini.Params, ent Entry, enc encoder.Encoder) (int, []byte) {
    return findCalendar(p["calendar_id"], func(cal *Calendar) (int, []byte) {
        ent.Id = db.makeKey()
		db.Writer(func(db DB) {
            cal.Entries[ent.Id] = &ent
        })
        return http.StatusCreated, encoder.Must(enc.Encode(ent))
    })
}

func entryUpdate(p martini.Params, update Entry, enc encoder.Encoder) (int, []byte) {
    return findCalendar(p["calendar_id"], func(cal *Calendar) (int, []byte) {
        return findEntry(cal, p["id"], func(ent *Entry) (int, []byte) {
			if len(update.Desc) != 0 {
				db.Writer(func(db DB) {
					ent.Desc = update.Desc
				})
			}
			if len(update.Start) != 0 {
				db.Writer(func(db DB) {
					ent.Start = update.Start
				})
			}
			if len(update.End) != 0 {
				db.Writer(func(db DB) {
					ent.End = update.End
				})
			}
            return http.StatusOK, encoder.Must(enc.Encode(*ent))
        })
    })
}

func entryDestroy(p martini.Params, enc encoder.Encoder) (int, []byte) {
    return findCalendar(p["calendar_id"], func(cal *Calendar) (int, []byte) {
        return findEntry(cal, p["id"], func(ent *Entry) (int, []byte) {
            db.Writer(func(db DB) {
                delete(cal.Entries, ent.Id)
            })
            return http.StatusNoContent, nil
        })
    })
}
