package ui

import (
	"fmt"
	"strings"

	"github.com/jr-k/d4s/internal/dao"
	"github.com/jr-k/d4s/internal/ui/components/inspect"
	"github.com/jr-k/d4s/internal/ui/components/view"
	"github.com/jr-k/d4s/internal/ui/styles"
)

func (a *App) RefreshCurrentView() {
	page, _ := a.Pages.GetFrontPage()
	
	// Handle Inspector Filtering
	if page == "inspect" {
		if a.ActiveInspector != nil {
			// Construct breadcrumb manually for inspector
			actionName := "inspect"

			switch v := a.ActiveInspector.(type) {
			case *inspect.LogInspector:
				actionName = "logs"
			case *inspect.TextInspector:
				actionName = strings.ToLower(v.Action)
				// Simplify "Describe container" to "describe" in breadcrumb
				if strings.Contains(actionName, " ") {
					parts := strings.Split(actionName, " ")
					if len(parts) > 0 {
						actionName = parts[0]
					}
				}
			case *inspect.StatsInspector:
				actionName = "stats"
			}

			status := ""

			// Start with base context
			// Use CurrentView if available, or ActiveScope logic
			baseView := a.CurrentView
			if baseView == "" {
				baseView = "containers" // Default fallback
			}
			
			scope := a.GetActiveScope()
			// If we have ActiveScope, it means we are in drilled down mode
			if scope != nil {
				// E.g. <compose> <containers> <logs>
				var breadcrumbs []string
				curr := scope
				for curr != nil {
					if curr.OriginView != "" {
						breadcrumbs = append([]string{curr.OriginView}, breadcrumbs...)
					}
					curr = curr.Parent
				}
				breadcrumbs = append(breadcrumbs, baseView)
				breadcrumbs = append(breadcrumbs, actionName)
				
				status += " "
				for i, s := range breadcrumbs {
					color := "#00ffff" // cyan
					if i == len(breadcrumbs)-1 {
						color = "orange"
					}
					firstChar := ""
					if i > 0 {
						firstChar = " "
					}
					status += fmt.Sprintf("%s[black:%s] <%s> ", firstChar, color, strings.ToLower(s))
					if i < len(breadcrumbs)-1 {
						status += "[black:black]"
					}
				}
				status += "[-:-:-]"
			} else {
				// E.g. <containers> <logs>
				scopes := []string{baseView, actionName}
				
				status += " "
				for i, s := range scopes {
					color := "#00ffff" // cyan
					if i == len(scopes)-1 {
						color = "orange"
					}
					firstChar := ""
					if i > 0 {
						firstChar = " "
					}
					status += fmt.Sprintf("%s[black:%s] <%s> ", firstChar, color, strings.ToLower(s))
					if i < len(scopes)-1 {
						status += "[black:black]"
					}
				}
				status += "[-:-:-]"
			}

			if !a.IsFlashLocked() {
				a.Flash.SetText(status)
			}
		}
		return
	}

	// Modal check logic needs specific naming convention or check
	if page == "help" || page == "logs" || page == "confirm" || page == "result" || page == "input" || page == "textview" {
		return
	}
	
	v, ok := a.Views[page]
	if !ok || v == nil {
		return
	}
	
	filter := a.ActiveFilter

	// 1. Immediate Updates (Optimistic UI)
	// UpdateShortcuts modifies the UI. Must be called from main thread.
	// Since RefreshCurrentView is sometimes called from valid UI context (SwitchTo) 
	// and sometimes from BG (Ticker -> QueueUpdateDraw), we assume we are in Main Thread here IF caller respected rules.
	// BUT, we previously wrapped Ticker calls in QueueUpdateDraw.
	// So UpdateShortcuts is safe here.
	a.UpdateShortcuts()
	
	a.RunInBackground(func() {
		if a.IsPaused() {
			return
		}

		var err error
		var data []dao.Resource
		var headers []string

		if v.FetchFunc != nil {
			data, err = v.FetchFunc(a, v)
			headers = v.Headers
		}

		// Check pause again after fetching (fetching can take time)
		if a.IsPaused() {
			return
		}

		a.SafeQueueUpdateDraw(func() {
			// Check if page changed while fetching?
			currentPage, _ := a.Pages.GetFrontPage()
			if currentPage != page {
				return
			}
			
			v.SetFilter(filter)
			v.CurrentScope = a.GetActiveScope()

			if err != nil {
				a.Flash.SetText(fmt.Sprintf("[red]Error: %v", err))
			} else {
				// Show actual title
				title := a.formatViewTitle(page, fmt.Sprintf("%d", len(v.Data)), filter)
				a.updateViewTitle(v, title)
				
				v.Update(headers, data)

				// Only update flash if not error
				status := ""
				scope := a.GetActiveScope()
				if scope != nil {
					// Dynamic breadcrumb trail coloring: last is always orange, others cyan
					// Traverse up to get full history
					var breadcrumbs []string
					curr := scope
					for curr != nil {
						if curr.OriginView != "" {
							breadcrumbs = append([]string{curr.OriginView}, breadcrumbs...)
						}
						curr = curr.Parent
					}
					// Add current page
					breadcrumbs = append(breadcrumbs, page)

					status += " "
					for i, s := range breadcrumbs {
						color := "#00ffff" // cyan
						if i == len(breadcrumbs)-1 {
							color = "orange"
						}
						firstChar := ""
						if i > 0 {
							firstChar = " "
						}
						// l'espace entre les éléments doit toujours être noir sur noir
						status += fmt.Sprintf("%s[black:%s] <%s> ", firstChar, color, strings.ToLower(s))
						if i < len(breadcrumbs)-1 {
							status += "[black:black]" // black fg, black bg for space
						}
					}
					status += "[-:-:-]"
				} else {
					status = fmt.Sprintf(" [black:orange] <%s> [-:-]", strings.ToLower(page))
				}

				if filter != "" {
					status += fmt.Sprintf(` [black:#bd93f9] <filter: %s> [-:-]`, filter)
				}
				
				// Only update flash if not locked by temporary message
				if !a.IsFlashLocked() {
					a.Flash.SetText(status)
				}
			}
		})
	})
}

func (a *App) formatViewTitle(viewName string, countStr string, filter string) string {
	viewName = strings.ToLower(viewName)
	
	// Default simple title
	title := fmt.Sprintf(" [#00ffff::b]%s[#00ffff][[white]%s[#00ffff]] ", viewName, countStr)
	
	// Dynamic recursive breadcrumb
	scope := a.GetActiveScope()
	if scope != nil {
		var parts []string
		
		// Walk up the stack
		curr := scope
		for curr != nil {
			cleanLabel := strings.ReplaceAll(curr.Label, "@", "[white] @ [#ff00ff]")
			origin := strings.ToLower(curr.OriginView)
			
			// Format: "origin(label)"
			part := fmt.Sprintf("[#00ffff::b]%s([-][#ff00ff]%s[#00ffff])", origin, cleanLabel)
			// Prepend to list (since we walk backwards)
			parts = append([]string{part}, parts...)
			
			curr = curr.Parent
		}
		
		// Append current view name
		parts = append(parts, fmt.Sprintf("[#00ffff]%s[#00ffff][[white]%s[#00ffff]]", viewName, countStr))
		
		title = " " + strings.Join(parts, " > ") + " "
	}
	
	if filter != "" {
		title += fmt.Sprintf(" [#00ffff][[white]Filter: [::b]%s[::-][#00ffff]] ", filter)
	}
	return title
}

func (a *App) updateViewTitle(v *view.ResourceView, title string) {
	v.Table.SetTitle(title)
	v.Table.SetTitleColor(styles.ColorTitle)
	v.Table.SetBorder(true)
	v.Table.SetBorderColor(styles.ColorTableBorder)
}

func (a *App) InspectCurrentSelection() {
	page, _ := a.Pages.GetFrontPage()
	view, ok := a.Views[page]
	if !ok {
		return
	}

	row, _ := view.Table.GetSelection()
	if row < 1 || row >= view.Table.GetRowCount() {
		return
	}

	dataIndex := row - 1
	if dataIndex < 0 || dataIndex >= len(view.Data) {
		return
	}
	
	resource := view.Data[dataIndex]
	id := resource.GetID()
	
	if view.InspectFunc != nil {
		view.InspectFunc(a, id)
		// InspectFunc typically opens a new modal or changes view.
		// We should refresh shortcuts to reflect the new state immediately.
		a.UpdateShortcuts()
		return
	}
	
	resourceType := "container"
	switch page {
		case styles.TitleImages:
			resourceType = "image"
		case styles.TitleVolumes:
			resourceType = "volume"
		case styles.TitleNetworks:
			resourceType = "network"
		case styles.TitleServices:
			resourceType = "service"
		case styles.TitleNodes:
			resourceType = "node"
		case styles.TitleCompose:
			resourceType = "compose"
	}

	content, err := a.Docker.Inspect(resourceType, id)
	if err != nil {
		a.Flash.SetText(fmt.Sprintf("[red]Inspect error: %v", err))
		return
	}

	a.OpenInspector(inspect.NewTextInspector("Inspect", id, content, "json"))
}
