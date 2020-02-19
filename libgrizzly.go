package grizzly

import (
	"database/sql"
	"log"
	"os/user"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
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

func openDB() *sql.DB {
	homeDirStr := homeDir() + "/Library/Group Containers/9K33E3U3T4.net.shinyfrog.bear/Application Data/database.sqlite"
	db, err := sql.Open("sqlite3", homeDirStr)
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


func GetAllNotes(notes *[]Note) {
	db := openDB()
	defer db.Close()

	rows, err := db.Query(`
		select n.Z_PK as id,
       		n.ZTEXT as text,
			n.ZTITLE as title,
			n.ZUNIQUEIDENTIFIER as identifier,
       		group_concat(t.ZTITLE) as tags
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
		group by id;`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var row Note
		var tagStr sql.NullString
		err = rows.Scan(&row.Id, &row.Text, &row.Title, &row.Identifier, &tagStr)
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

func GetAllWithTags(notes *[]NoteTag) {
	db := openDB()
	defer db.Close()

	rows, err := db.Query(`
		select n.Z_PK as id,
       		n.ZTITLE as title,
       		n.ZUNIQUEIDENTIFIER as identifier,
       		group_concat(t.ZTITLE) as tags
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        	left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
		where t.ZTITLE is not null
		group by id;`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var row NoteTag
		var tagStr string
		err = rows.Scan(&row.Id, &row.Title, &row.Identifier, &tagStr)
		row.Tags = strings.Split(tagStr, ",")
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

func GetAllMarked(notes *[]Note) {
	db := openDB()
	defer db.Close()

	rows, err := db.Query(`
		select n.Z_PK as id,
       		n.ZTEXT as text,
       		n.ZTITLE as title,
       		group_concat(t.ZTITLE) as tags,
       		n.ZUNIQUEIDENTIFIER as identifier
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
        where text like "%::%::%"
		group by id;`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var row Note
		var text sql.NullString
		var tagStr sql.NullString
		err = rows.Scan(&row.Id, &text, &row.Title, &tagStr, &row.Identifier)
		if text.Valid {
			row.Text = text.String
		} else {
			row.Text = ""
		}
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

func GetTailWithTags(notes *[]NoteTag, limit int) {
	db := openDB()
	defer db.Close()

	rows, err := db.Query(`
		select n.Z_PK as id,
       		n.ZTITLE as title,
       		n.ZUNIQUEIDENTIFIER as identifier,
       		group_concat(t.ZTITLE) as tags
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        	left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
		group by id
		order by id asc
		limit ?;`, limit)
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

func GetHeadWithTags(notes *[]NoteTag, limit int) {
	db := openDB()
	defer db.Close()

	rows, err := db.Query(`
		select n.Z_PK as id,
       		n.ZTITLE as title,
       		n.ZUNIQUEIDENTIFIER as identifier,
       		group_concat(t.ZTITLE) as tags
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        	left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
		group by id
		order by id desc
		limit ?;`, limit)
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

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func GetDuplicates(notes *[]NoteDuplicate) {
	db := openDB()
	defer db.Close()

	rows, err := db.Query(`
			select Z_PK as id, ZTITLE as title, count(ZTITLE) as count
			from ZSFNOTE 
			group by ZTITLE having ( count > 1 );
		`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var row NoteDuplicate
		err = rows.Scan(&row.Id, &row.Title, &row.Count)
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

func GetUnlinked() map[string][]string {
	var allNotes []Note
	GetAllNotes(&allNotes)
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

func SearchTitles(partialTitle string, notes *[]NoteTag) {
	db := openDB()
	defer db.Close()

	rows, err := db.Query(`select n.Z_PK as id,
       		n.ZTITLE as title,
       		n.ZUNIQUEIDENTIFIER as identifier,
       		group_concat(t.ZTITLE) as tags
		from ZSFNOTE as n left join Z_7TAGS as tn on n.Z_PK=tn.Z_7NOTES
        	left join ZSFNOTETAG as t on tn.Z_14TAGS=t.Z_PK
		where t.ZTITLE like ?
		group by id;`, "%" + partialTitle + "%")
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
