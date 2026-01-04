package auth

type Role string

const (
	Admin Role = "admin"
	User  Role = "user"
	Guest Role = "guest"
)

type Permission string

const (
	PermAlarm  Permission = "alarm"
	PermDevice Permission = "device"
	PermGuest  Permission = "guest"
)

var rolePermissions = map[Role][]Permission{
	Admin: {PermAlarm, PermDevice, PermGuest},
	User:  {PermAlarm, PermDevice},
	Guest: {PermGuest},
}

func HasPermission(role Role, perm Permission) bool {
	perms := rolePermissions[role]
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}
