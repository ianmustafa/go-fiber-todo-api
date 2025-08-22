package interfaces

// Repositories contains all repository interfaces
type Repositories struct {
	User UserRepository
	Todo TodoRepository
}

// NewRepositories creates a new repositories container
func NewRepositories(user UserRepository, todo TodoRepository) *Repositories {
	return &Repositories{
		User: user,
		Todo: todo,
	}
}
