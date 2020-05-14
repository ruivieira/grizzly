package grizzly

import (
	"database/sql"
	"log"
	"os/user"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type NoteTag struct {
	Id    int
	Title string
	Tags  []string
	Identifier string
}

type Note struct {
	Id         int
	Title      string
	Text       string
	Tags       []string
	Identifier string
}

func OpenDB() *gorm.DB {
	homeDirStr := homeDir() + "/Library/Group Containers/9K33E3U3T4.net.shinyfrog.bear/Application Data/database.sqlite"
	db, err := gorm.Open("sqlite3", homeDirStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

type NoteDuplicate struct {
	Id    int
	Title string
	Count int
}


func GetAllNotes(db *gorm.DB, notes *[]Note) {
	rows, err := db.Raw(`
		select n.Z_PK as id,
       		n.ZTEXT as text,
			n.ZTITLE as title,
			n.ZUNIQUEIDENTIFIER as identifier,
       		group_concat(t.ZTITLE) as tags
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
		group by id;`).Rows()
	defer rows.Close()
	if err != nil {
		log.Fatal("Could not process query")
	}
	for rows.Next() {
		var note Note
		var text sql.NullString
		var tagStr sql.NullString
		err = rows.Scan(&note.Id, &text, &note.Title, &note.Identifier, &tagStr)
		if text.Valid {
			note.Text = text.String
		}
		if tagStr.Valid {
			note.Tags = strings.Split(tagStr.String, ",")
		}
		*notes = append(*notes, note)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetAllWithTags(db *gorm.DB, notes *[]NoteTag) {
	rows, err := db.Raw(`
		select n.Z_PK as id,
       		n.ZTITLE as title,
       		n.ZUNIQUEIDENTIFIER as identifier,
       		group_concat(t.ZTITLE) as tags
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        	left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
		where t.ZTITLE is not null
		group by id;`).Rows()
	defer rows.Close()
	if err != nil {
		log.Fatal("Could not process query")
	}
	for rows.Next() {
		var note NoteTag
		var tagStr sql.NullString
		err = rows.Scan(&note.Id, &note.Title, &note.Identifier, &tagStr)
		if tagStr.Valid {
			note.Tags = strings.Split(tagStr.String, ",")
		}
		*notes = append(*notes, note)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetAllMarked(db *gorm.DB, notes *[]Note) {
	db.Raw(`
		select n.Z_PK as id,
       		n.ZTEXT as text,
       		n.ZTITLE as title,
       		group_concat(t.ZTITLE) as tags,
       		n.ZUNIQUEIDENTIFIER as identifier
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
        where text like "%::%::%"
		group by id;`).Scan(notes)
}

func GetTailWithTags(db *gorm.DB, notes *[]NoteTag, limit int) {
	rows, err := db.Raw(`
		select n.Z_PK as id,
       		n.ZTITLE as title,
       		n.ZUNIQUEIDENTIFIER as identifier,
       		group_concat(t.ZTITLE) as tags
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        	left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
		group by id
		order by id asc
		limit ?;`, limit).Rows()
	defer rows.Close()
	if err != nil {
		log.Fatal("Could not process query")
	}
	for rows.Next() {
		var note NoteTag
		var tagStr sql.NullString
		err = rows.Scan(&note.Id, &note.Title, &note.Identifier, &tagStr)
		if tagStr.Valid {
			note.Tags = strings.Split(tagStr.String, ",")
		}
		*notes = append(*notes, note)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetHeadWithTags(db *gorm.DB, notes *[]NoteTag, limit int) {
	rows, err := db.Raw(`
		select n.Z_PK as id,
       		n.ZTITLE as title,
       		n.ZUNIQUEIDENTIFIER as identifier,
       		group_concat(t.ZTITLE) as tags
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        	left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
		group by id
		order by id desc
		limit ?;`, limit).Rows()
	defer rows.Close()
	if err != nil {
		log.Fatal("Could not process query")
	}
	for rows.Next() {
		var note NoteTag
		var tagStr sql.NullString
		err = rows.Scan(&note.Id, &note.Title, &note.Identifier, &tagStr)
		if tagStr.Valid {
			note.Tags = strings.Split(tagStr.String, ",")
		}
		*notes = append(*notes, note)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func GetDuplicates(db *gorm.DB, notes *[]NoteDuplicate) {
	db.Raw(`
			select Z_PK as id, ZTITLE as title, count(ZTITLE) as count
			from ZSFNOTE 
			group by ZTITLE having ( count > 1 );
		`).Scan(notes)
}

func GetUnlinked(db *gorm.DB) map[string][]string {
	var allNotes []Note
	GetAllNotes(db, &allNotes)
	reference := make(map[string][]string)
	r, _ := regexp.Compile("\\(bear:\\/\\/x-callback-url\\/open-note?(.*)\\)")
	for _, note := range allNotes {
		reference[note.Identifier] = make([]string, 0)
		matches := r.FindAllString(note.Text, -1)
		for _, mark := range matches {
			identifier := mark[36 : len(mark)-1]
			reference[identifier] = append(reference[identifier], note.Identifier)
		}
	}
	return reference
}

func SearchTitles(db *gorm.DB, partialTitle string, notes *[]NoteTag) {
	rows, err := db.Raw(`select n.Z_PK as id,
       		n.ZTITLE as title,
       		n.ZUNIQUEIDENTIFIER as identifier,
       		group_concat(t.ZTITLE) as tags
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        	left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
		where t.ZTITLE like ?
		group by id;`, "%" + partialTitle + "%").Rows()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var row NoteTag
		var tagStr sql.NullString
		err = rows.Scan(&row.Id, &row.Title, &row.Identifier, &tagStr)
		if tagStr.Valid {
			row.Tags = strings.Split(tagStr.String, ",")
		}
		*notes = append(*notes, row)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}
