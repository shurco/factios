package database

import (
	"errors"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/sdomino/scribble"
	"github.com/shurco/factios/logger"
	"github.com/shurco/factios/model"
)

var (
	log = logger.GetLogger("database")

	errInvalidSignature = errors.New("Invalid signature")
)

// DB is ...
type DB interface {
	GetRandomFact(lng string) (*model.Fact, error)
	GetFactByID(lng, short string) (*model.Fact, error)
}

// JSONDB is ...
type JSONDB struct {
	folder string
	db     *scribble.Driver
}

// NewDB is ...
func NewDB(folder string) DB {
	db, err := scribble.New(folder, nil)
	if err != nil {
		log.Error().Err(err)
	}

	return JSONDB{
		folder: folder,
		db:     db,
	}
}

// GetRandomFact is ...
func (d JSONDB) GetRandomFact(lang string) (*model.Fact, error) {
	files, err := ioutil.ReadDir(d.folder + lang)
	if err != nil {
		log.Error().Err(err)
		return nil, errInvalidSignature
	}
	rand.Seed(time.Now().UTC().UnixNano())
	file := strings.Split(files[rand.Intn(len(files))].Name(), ".")

	fact := model.Fact{}
	err = d.db.Read(lang, file[0], &fact)
	if err != nil {
		log.Error().Err(err)
		return nil, errInvalidSignature
	}
	fact.Short = file[0]
	return &fact, nil
}

// GetFactByID is ...
func (d JSONDB) GetFactByID(lang, Short string) (*model.Fact, error) {
	fact := model.Fact{}
	err := d.db.Read(lang, Short, &fact)
	if err != nil {
		log.Error().Err(err)
		return nil, errInvalidSignature
	}
	fact.Short = Short
	return &fact, nil
}
