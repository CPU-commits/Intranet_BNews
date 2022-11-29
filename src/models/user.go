package models

type UserTypes string

const (
	DIRECTOR          = "f"
	DIRECTIVE         = "e"
	TEACHER           = "d"
	ATTORNEY          = "c"
	STUDENT_DIRECTIVE = "b"
	STUDENT           = "a"
)

type User struct {
	Name           string `json:"name,omitempty" bson:"name" extensions:"x-omitempty" example:"Karen"`
	FirstLastname  string `json:"first_lastname,omitempty" bson:"first_lastname" extensions:"x-omitempty" example:"Rojas"`
	SecondLastname string `json:"second_lastname,omitempty" bson:"second_lastname" extensions:"x-omitempty" example:"Valdes"`
	ID             string `json:"_id,omitempty" bson:"_id" extensions:"x-omitempty" example:"638660ca141aa4ee9faf07e8"`
}
