package render

import "github.com/OskarLeirvaag/Lootsheet/src/render/dashboard"

// DefaultDashboardData returns the placeholder content used when no adapter is wired yet.
func DefaultDashboardData() DashboardData { return dashboard.DefaultData() }

// ErrorDashboardData returns a dashboard model that keeps the TUI open while surfacing an error.
func ErrorDashboardData(summary string, detail string) DashboardData {
	return dashboard.ErrorData(summary, detail)
}

func resolveDashboardData(data *DashboardData) DashboardData { return dashboard.ResolveData(data) }
func dashboardDataEmpty(data *DashboardData) bool            { return dashboard.DataEmpty(data) }
