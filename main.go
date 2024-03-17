package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

type AppGlobals struct {
	Debug    bool
	DbName   string
	DbParams string
}

 

var constants = &AppGlobals{
	Debug:    true,
	DbName:   "duckdb",
	DbParams: "?access_mode=READ_WRITE",}

type Timestamp time.Time

type PugData struct {
	Id          uint64
	TheLink     string
	Description sql.NullString
	Created     time.Time
	Updated     sql.NullTime
	Deleted     sql.NullTime

	bob *Timestamp
}

func (data PugData) String() string {
	return fmt.Sprintf("%d,%s,%s,%v", data.Id, data.TheLink, data.Description.String, data.Created)
}

func NewPugData(alink, desc string) *PugData {
	now := time.Now()
	return &PugData{
		TheLink:     alink,
		Description: sql.NullString{String: desc, Valid: true},
		Created:     now,
		bob:         (*Timestamp)(&now),
	}
}

type PugData2Label struct {
	LabelId uint64
	DataId  uint64
}

type PugUser struct {
	Id          uint64
	Name        string
	Description sql.NullString
}

type PugLabel struct {
	Id          uint64
	ParentId    uint64
	Label       string
	Description sql.NullString
	Created     time.Time
	Deleted     sql.NullTime
}

const (

	// defaults to an in memory database. Put a full path for a persistent one

	DB_TBL1_SEQ = "CREATE SEQUENCE IF NOT EXISTS pugdata_seq START 1;"
	DB_TBL2_SEQ = "CREATE SEQUENCE IF NOT EXISTS puglabel_seq START 100;"
	DB_TBL1_CRT = "CREATE TABLE IF NOT EXISTS pug_data(id int8 primary key DEFAULT nextval('pugdata_seq'), thelink VARCHAR NOT NULL, description VARCHAR, created TIMESTAMP NOT NULL DEFAULT current_timestamp, updated TIMESTAMP, deleted TIMESTAMP)"
	DB_TBL2_CRT = "CREATE TABLE IF NOT EXISTS pug_label(id int8 primary key DEFAULT nextval('puglabel_seq'), parent_id int8, label VARCHAR NOT NULL, description VARCHAR, created TIMESTAMP NOT NULL DEFAULT current_timestamp, deleted TIMESTAMP)"
	DB_TBL3_CRT = "CREATE TABLE IF NOT EXISTS pug_data_label(label_id int8, data_id int8, deleted TIMESTAMP, PRIMARY KEY (label_id, data_id), FOREIGN KEY (label_id) REFERENCES pug_label (id), FOREIGN KEY (data_id) REFERENCES pug_data (id) )"

	DB_TBL1_INS_PST = "INSERT INTO pug_data (thelink,description) VALUES (?,?)"
	DB_TBL2_INS_PST = "INSERT INTO pug_label (parent_id,label,description) VALUES (?,?,?)"
	DB_TBL3_INS_PST = "INSERT INTO pug_data_label (label_id,data_id) VALUES (?,?)"

	DB_TBL1_INS = "INSERT INTO pug_data (thelink,description) VALUES ('http://google.de/bla/blub','Search')"
)

// https://pkg.go.dev/fmt#Printf  --> All Format-Identifiers
func main() {

	defer initLogging("pugnet").Close()

	data := NewPugData("http://www.netflix.de", "Search Thingy")
	log.Trace().Msgf("New PugData created: %s", data.String())

	db := OpenDatabase(constants.DbName, constants.DbParams)
	defer db.Close()

	check(db.Ping())

	createSchemaIfNeeded(db)
	createSomeTestdatafWanted(db)

	insertPugDataRow(db, &PugData{
		TheLink:     "https://regexper.com/",
		Description: sql.NullString{String: "Regex visualisation", Valid: true},
	})

	row := db.QueryRow("SELECT id, thelink, description, created, updated FROM pug_data where id = 1  ")
	r1 := &PugData{}
	err := row.Scan(&r1.Id, &r1.TheLink, &r1.Description, &r1.Created, &r1.Updated)
	handle(err, "Cannot Select single row from pug_data! : %v", r1)

	log.Print("r1.updated", r1.Updated)

	log.Info().Uint64("id", r1.Id).
		Str("TheLink", r1.TheLink).
		Str("description", r1.Description.String).
		Time("created", r1.Created).
		Msgf("updated %v", r1.Updated)
	var ctx context.Context = context.TODO()
	fmt.Println("start selecting all from pugdata..")
	rows, err := db.QueryContext(ctx, "SELECT thelink FROM pug_data ")

	handle(err, "Cannot Select from pug_data!")

	defer rows.Close()
	names := make([]string, 0)

	for rows.Next() {
		fmt.Println("calling next..")
		var name string
		if err := rows.Scan(&name); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			log.Error().Err(err).Msg("")
		}
		fmt.Println("name", name)
		names = append(names, name)
	}
	// If the database is being written to ensure to check for Close
	// errors that may be returned from the driver. The query may
	// encounter an auto-commit error and be forced to rollback changes.
	rerr := rows.Close()
	handle(rerr, "Cannot rows.close")

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("")
	}
	fmt.Printf("%s are %d years old", strings.Join(names, ", "), 2)

}

// DDLs are with 'IF NOT EXISTS' so we can try to create without any hassle
func createSchemaIfNeeded(db *sql.DB) {
	log.Debug().Msg("DDL: Check Schema: Sequences...")
	ExecuteUpdate(db, DB_TBL1_SEQ)
	ExecuteUpdate(db, DB_TBL2_SEQ)

	log.Debug().Msg("DDL: Check Schema: Create tables...")
	ExecuteUpdate(db, DB_TBL1_CRT)

	ExecuteUpdate(db, DB_TBL2_CRT)

	ExecuteUpdate(db, DB_TBL3_CRT)

	ExecuteUpdate(db, "CREATE TABLE person (id INTEGER, name VARCHAR)")
}

func createSomeTestdatafWanted(db *sql.DB) {
	ExecuteUpdate(db, DB_TBL1_INS)

	ExecuteUpdate(db, "INSERT INTO person VALUES (1, 'John')")
	ExecuteUpdate(db, "INSERT INTO person VALUES (2, 'Claudia')")
	ExecuteUpdate(db, "INSERT INTO person VALUES (3, 'Martha')")
	db.Exec(`INSERT INTO person VALUES (44, 'Sepp')`)
	log.Trace().Msg("rows inserted")
}

// Logging will be json to a file called 'pugnet.log.json'. Also the log.Logger will be globally set ?!
func initLogging(app string) *os.File {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if constants.Debug {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimestampFieldName = "tmstmp"
	zerolog.LevelFieldName = "lvl"
	zerolog.MessageFieldName = "msg"

	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix // default is time.RFC3339

	file, err := os.OpenFile("pugnet.log.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		handle(err, "Cannot open logfile!")
	}
	hname, _ := os.Hostname()

	mw := io.MultiWriter(os.Stdout, file)

	// global configuration how the basic structure of our logger should look like
	mylog := zerolog.New(mw).With().Dict("nfo", zerolog.Dict().Timestamp().Str("app", app).Str("host", hname)).Logger()

	log.Logger = mylog // this should set the logger globally?
	mylog.Info().
		Bool("TraceEnabled", mylog.Trace().Enabled()).
		Bool("DebugEnabled", mylog.Debug().Enabled()).
		Bool("InfoEnabled", mylog.Info().Enabled()).
		Str("LogfileName", file.Name()).
		Msg("Logging successfully configured!")

	if e := log.Debug(); e.Enabled() {

		// Compute log output only if enabled.
		value := "bar"
		e.Str("foo", value).Msg("some debug message")
	}

	return file // to defer the Close() method!
}

func OpenDatabase(drv, params string) *sql.DB {
	db, err := sql.Open(drv, params)
	handle(err, "cannot open database: drv=%s, params=%s", drv, params)
	return db
}

func ExecuteUpdate(db *sql.DB, sql string, args ...any) {
	_, err := db.Exec(sql, args...)
	handle(err, "ExecuteUpdate failed at %s", sql)
}

func check(args ...interface{}) {
	err := args[len(args)-1]
	if err != nil {
		log.Error().Stack().Caller().Msg("Panic!")
		panic(err)
	}
}

func handle(err error, msg string, args ...any) {
	if err != nil {
		log.Panic().Stack().Caller().Err(err).Msgf(msg, args...)
	}
}

func InsertData(db *sql.DB, sql string) {
	stmt, err := db.PrepareContext(context.Background(), "INSERT INTO pug_data(thelink,description) VALUES( ?, ?)")
	check(err)
	defer stmt.Close()

	check(stmt.ExecContext(context.Background(), "Kevin", 11, 0.55, true, "2013-07-06"))
	check(stmt.ExecContext(context.Background(), "Bob", 12, 0.73, true, "2012-11-04"))
	check(stmt.ExecContext(context.Background(), "Stuart", 13, 0.66, true, "2014-02-12"))

	stmt, err = db.PrepareContext(context.Background(), "SELECT * FROM users WHERE name = ?")
	check(err)

	rows, err := stmt.QueryContext(context.Background(), "Bob")
	check(err)
	defer rows.Close()

	for rows.Next() {
		u := new(PugData)
		err := rows.Scan(&u.TheLink, &u.Description, &u.Id, &u.Created)
		if err != nil {
			log.Error().Err(err).Msg("")
		}
		//log.Printf("%s is %d years old, %.2f tall, bday on %s and has awesomeness: %t\n",			u.servername.String, u.scheme, u.id, u.created.Format(time.RFC3339), u.id)
	}
}

func insertPugDataRow(db *sql.DB, data *PugData) {
	tx, err := db.Begin()
	if err != nil {
		log.Error().Err(err).Msg("Cannot Begin Tx")
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(DB_TBL1_INS_PST)
	if err != nil {
		log.Error().Err(err).Msg("Cannot prepare Statement")
	}
	defer stmt.Close() // danger!
	//for i := 0; i < 10; i++ {
	//_, err = stmt.Exec(data.scheme, data.servername, data.domain, data.path, data.suffix, data.description)
	_, err = stmt.Exec(data.TheLink, data.Description)
	if err != nil {
		log.Error().Err(err).Msg("Cannot execute Statement")
	}
	//}
	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("Cannot commit Tx")
	}

}

// Matches URLs and can also be used to split parts in groups
// group 1: https
// group 2: server and domain
// group 3: port
// group 4: path till file
// group 5: parameters
// group 6: #xxxx-> stuff after #
func ValidateUrl(url string) bool {
	//var re = regexp.MustCompile(`^(?:([A-Za-z]+):)?(?:\/{0,3})([0-9.\-A-Za-z]+)(?::(\d+))?(?:\/([^?#]*))?(?:\?([^#]*))?(?:#(.*))?$`)
	var re = regexp.MustCompile(`^(?:(http|https):){1}?(?:\/{0,3})([0-9.\-A-Za-z]+)(?::(\d+))?(?:\/([^?#]*))?(?:\?([^#]*))?(?:#(.*))?$`)
	return re.MatchString(url)
}
