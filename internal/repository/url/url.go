package url

import "errors"

var ErrConflict = errors.New("original URL already exists")

type URLRepository interface {
	Save(url string) (string, error)

	FindByID(id string) (string, bool)

	SaveAll(urls []string) ([]string, error)
}
