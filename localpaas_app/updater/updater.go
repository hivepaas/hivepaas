package updater

type Updater interface {
	Start() error
	Shutdown() error
}
