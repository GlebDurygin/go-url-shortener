package repository

type URLRepository interface {
	Save(url string) (string, error)

	FindByID(id string) (string, bool)
}
