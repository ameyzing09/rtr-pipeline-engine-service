package models

// Role defines the user roles in the system
type Role string

const (
	RoleAdmin       Role = "ADMIN"
	RoleHR          Role = "HR"
	RoleInterviewer Role = "INTERVIEWER"
)

// String returns the string representation of the role
func (r Role) String() string {
	return string(r)
}
