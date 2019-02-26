package discovery

type (
	Registry interface {
		Update(*Address)
	}
)
