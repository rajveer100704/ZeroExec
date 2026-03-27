package gateway

import (
	"regexp"
	"strings"
)

type Permission string

const (
	PermTerminalRead  Permission = "terminal:read"
	PermTerminalWrite Permission = "terminal:write"
	PermSessionManage Permission = "session:manage"
)

var RolePermissions = map[UserRole][]Permission{
	RoleViewer:   {PermTerminalRead},
	RoleOperator: {PermTerminalRead, PermTerminalWrite},
	RoleAdmin:    {PermTerminalRead, PermTerminalWrite, PermSessionManage},
}

// CommandPolicy defines which commands a role is allowed or denied.
type CommandPolicy struct {
	// If AllowAll is true, no command filtering is applied.
	AllowAll bool
	// DeniedPatterns are regexes. If any match, the command is blocked.
	DeniedPatterns []*regexp.Regexp
	// AllowedPatterns are regexes. If set (and AllowAll is false),
	// only commands matching at least one pattern are allowed.
	AllowedPatterns []*regexp.Regexp
}

// RoleCommandPolicies maps each role to its command-level restrictions.
var RoleCommandPolicies = map[UserRole]*CommandPolicy{
	RoleViewer: {
		AllowAll: false,
		AllowedPatterns: compilePatterns([]string{
			`^(ls|dir|pwd|cd|echo|cat|type|help|cls|clear|whoami|hostname|date|time|Get-\w+|get-\w+)(\s.*)?$`,
		}),
	},
	RoleOperator: {
		AllowAll: false,
		DeniedPatterns: compilePatterns([]string{
			`^(rm|del|rmdir|rd|Remove-Item|format|shutdown|restart|Stop-Computer|Restart-Computer)(\s.*)?$`,
			`^sudo(\s.*)?$`,
			`^(net\s+user|net\s+localgroup)(\s.*)?$`,
			`^(reg|regedit|schtasks)(\s.*)?$`,
			`^powershell\s+.*-enc.*$`,
		}),
	},
	RoleAdmin: {
		AllowAll: true,
	},
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		compiled = append(compiled, regexp.MustCompile("(?i)"+p))
	}
	return compiled
}

type RBAC struct{}

func (r *RBAC) CanPerform(role UserRole, perm Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// IsCommandAllowed checks if the given command string is permitted for the role.
func (r *RBAC) IsCommandAllowed(role UserRole, command string) bool {
	command = strings.TrimSpace(command)
	if command == "" {
		return true
	}

	policy, ok := RoleCommandPolicies[role]
	if !ok {
		return false // Unknown role: deny all
	}

	if policy.AllowAll {
		return true
	}

	// Check denied patterns first (for Operator)
	for _, deny := range policy.DeniedPatterns {
		if deny.MatchString(command) {
			return false
		}
	}

	// If there are allowed patterns (for Viewer), command must match at least one
	if len(policy.AllowedPatterns) > 0 {
		for _, allow := range policy.AllowedPatterns {
			if allow.MatchString(command) {
				return true
			}
		}
		return false // No allowed pattern matched
	}

	// If no denied pattern matched and no allowed patterns are defined, allow
	return true
}
