package url

type URLRepository interface {
	Save(url string) (string, error)

	FindByID(id string) (string, bool)

	SaveAll(urls []string) ([]string, error)
}
