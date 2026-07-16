package repository

type Repository interface {
	InsertUser()
	GetUser()
	DeleteUser()
	UpdateUser()
}
