package auth_utils

type ShortRole string

const rolePrefix = "PegNu-Short."

const (
	RoleCreate          ShortRole = rolePrefix + "CREATE"
	RoleDelete          ShortRole = rolePrefix + "DELETE"
	RoleOverwrite       ShortRole = rolePrefix + "OVERWRITE"
	RoleClearBackground ShortRole = rolePrefix + "CLEAR-BACKGROUND"
	RoleAdmin           ShortRole = rolePrefix + "ADMIN"
)
