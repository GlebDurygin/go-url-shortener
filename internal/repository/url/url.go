package url

import "errors"

var ErrConflict = errors.New("original URL already exists")

type URLRepository interface {
	Save(url, userID string) (string, error)

	FindByID(id string) (string, bool)

	SaveAll(urls []string, userID string) ([]string, error)

	FindByUserID(userID string) (map[string]string, error)
}
