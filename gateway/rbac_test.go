package gateway

import "testing"

func TestRBAC_IsCommandAllowed(t *testing.T) {
	rbac := &RBAC{}

	tests := []struct {
		name    string
		role    UserRole
		command string
		want    bool
	}{
		// Admin: everything allowed
		{"admin can rm", RoleAdmin, "rm -rf /tmp/test", true},
		{"admin can ls", RoleAdmin, "ls", true},
		{"admin can shutdown", RoleAdmin, "shutdown /s", true},

		// Viewer: only inspection commands
		{"viewer can ls", RoleViewer, "ls", true},
		{"viewer can dir", RoleViewer, "dir", true},
		{"viewer can pwd", RoleViewer, "pwd", true},
		{"viewer can echo", RoleViewer, "echo hello", true},
		{"viewer can cat", RoleViewer, "cat file.txt", true},
		{"viewer can whoami", RoleViewer, "whoami", true},
		{"viewer can Get-Process", RoleViewer, "Get-Process", true},
		{"viewer cannot rm", RoleViewer, "rm file.txt", false},
		{"viewer cannot mkdir", RoleViewer, "mkdir test", false},
		{"viewer cannot python", RoleViewer, "python script.py", false},
		{"viewer cannot shutdown", RoleViewer, "shutdown /s", false},

		// Operator: everything except destructive
		{"operator can ls", RoleOperator, "ls", true},
		{"operator can mkdir", RoleOperator, "mkdir test", true},
		{"operator can python", RoleOperator, "python script.py", true},
		{"operator cannot rm", RoleOperator, "rm file.txt", false},
		{"operator cannot del", RoleOperator, "del file.txt", false},
		{"operator cannot rmdir", RoleOperator, "rmdir test", false},
		{"operator cannot Remove-Item", RoleOperator, "Remove-Item file.txt", false},
		{"operator cannot shutdown", RoleOperator, "shutdown /s", false},
		{"operator cannot sudo", RoleOperator, "sudo apt update", false},
		{"operator cannot reg", RoleOperator, "reg add HKLM\\test", false},

		// Edge cases
		{"empty command allowed", RoleViewer, "", true},
		{"whitespace only allowed", RoleViewer, "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rbac.IsCommandAllowed(tt.role, tt.command)
			if got != tt.want {
				t.Errorf("IsCommandAllowed(%q, %q) = %v, want %v", tt.role, tt.command, got, tt.want)
			}
		})
	}
}

func TestRBAC_CanPerform(t *testing.T) {
	rbac := &RBAC{}

	tests := []struct {
		role UserRole
		perm Permission
		want bool
	}{
		{RoleViewer, PermTerminalRead, true},
		{RoleViewer, PermTerminalWrite, false},
		{RoleOperator, PermTerminalWrite, true},
		{RoleOperator, PermSessionManage, false},
		{RoleAdmin, PermSessionManage, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.role)+"_"+string(tt.perm), func(t *testing.T) {
			if got := rbac.CanPerform(tt.role, tt.perm); got != tt.want {
				t.Errorf("CanPerform(%q, %q) = %v, want %v", tt.role, tt.perm, got, tt.want)
			}
		})
	}
}
