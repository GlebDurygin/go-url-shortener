package repository

type URLRepository interface {
	Save(url string) string

	FindById(id string) (string, bool)
}
