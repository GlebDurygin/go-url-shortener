package repository

type URLRepository interface {
	Save(url string) string

	FindByID(id string) (string, bool)
}
