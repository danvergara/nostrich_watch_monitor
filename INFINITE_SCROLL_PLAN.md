# Infinite Scroll Implementation Plan

## Overview
Implement infinite scroll pagination for the relays table using HTMX to load more relays as the user scrolls down, fetching data from the database using OFFSET and LIMIT.

## Current State
- ✅ Basic table structure exists (`RelayTable` component)
- ✅ Pagination logic exists in service layer (`RelayFilters` with Limit/Offset)
- ✅ HTMX is already included in the project
- ✅ Handler currently loads first 10 relays with hardcoded limit/offset

## Implementation Tasks

### 1. Create New API Endpoint for Paginated Relay Rows
**File**: `internal/handlers/relays.go`
**Status**: Pending

Add new method `HandleRelayRows` that:
- Accepts `limit` and `offset` query parameters
- Returns only table rows (not full page)
- Uses existing `RelayFilters` and service layer

```go
func (rh *RelaysHandler) HandleRelayRows(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")
    
    limit := 10 // default
    offset := 0 // default
    
    if limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil {
            limit = l
        }
    }
    
    if offsetStr != "" {
        if o, err := strconv.Atoi(offsetStr); err == nil {
            offset = o
        }
    }
    
    filters := &services.RelayFilters{
        Limit:  &limit,
        Offset: &offset,
    }
    
    relays, err := rh.service.GetRelays(r.Context(), filters)
    if err != nil {
        http.Error(w, "Error fetching relays", http.StatusInternalServerError)
        return
    }
    
    // Return only the table rows, not the full page
    if err := components.RelayTableRows(ToRelayTableViewModels(relays), offset+limit).Render(r.Context(), w); err != nil {
        http.Error(w, "Error rendering relay rows", http.StatusInternalServerError)
        return
    }
}
```

### 2. Create RelayTableRows Template
**File**: `web/views/components/relay_table_rows.templ`
**Status**: Pending

New template that renders only table rows with infinite scroll trigger:

```templ
package components

import (
    "fmt"
    "github.com/danvergara/nostrich_watch_monitor/pkg/presentation"
)

templ RelayTableRows(relays []presentation.RelayTableViewModel, nextOffset int) {
    for _, relay := range relays {
        @RelayTableRow(relay)
    }
    
    if len(relays) > 0 {
        <!-- Infinite scroll trigger -->
        <tr id="load-more-trigger" 
            hx-get={ fmt.Sprintf("/api/relays?offset=%d&limit=10", nextOffset) }
            hx-trigger="revealed"
            hx-target="#relay-table-body"
            hx-swap="beforeend"
            hx-indicator="#loading-indicator">
            <td colspan="6" class="text-center py-4">
                <div id="loading-indicator" class="htmx-indicator">
                    <div class="flex items-center justify-center space-x-2">
                        <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
                        <span class="text-gray-400">Loading more relays...</span>
                    </div>
                </div>
            </td>
        </tr>
    }
}
```

### 3. Modify RelayTable Template
**File**: `web/views/components/relay_table.templ`
**Status**: Pending

Update existing template to:
- Add `id="relay-table-body"` to tbody
- Use new `RelayTableRows` component
- Pass initial offset (10) for next batch

```templ
templ RelayTable(relays []presentation.RelayTableViewModel) {
    <div class="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        <div class="overflow-x-auto">
            <table class="w-full min-w-[1200px]">
                <thead class="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10">
                    <!-- ... existing header unchanged ... -->
                </thead>
                <tbody id="relay-table-body" class="divide-y divide-gray-200 dark:divide-gray-700">
                    @RelayTableRows(relays, 10)
                </tbody>
            </table>
        </div>
    </div>
}
```

### 4. Update Routes
**File**: `internal/server/routes.go`
**Status**: Pending

Add new API endpoint:

```go
func addRoutes(mux *http.ServeMux, cfg *config.Config, fs fs.FS, handler handlers.RelaysHandler) {
    mux.Handle(
        "/static/",
        http.StripPrefix("/static/", http.FileServer(http.FS(fs))),
    )

    mux.HandleFunc("/", handler.HandleRelayIndex)
    mux.HandleFunc("/relay", handler.HandleRelayDetail)
    mux.HandleFunc("/api/relays", handler.HandleRelayRows) // New endpoint
}
```

### 5. Add CSS for Loading Indicator
**File**: `web/static/css/styles.css`
**Status**: Pending

Add HTMX indicator styles:

```css
.htmx-indicator {
    display: none;
}

.htmx-request .htmx-indicator {
    display: block;
}

.htmx-request.htmx-indicator {
    display: block;
}
```

### 6. Add Required Import
**File**: `internal/handlers/relays.go`
**Status**: Pending

Add `strconv` import for parsing query parameters:

```go
import (
    "net/http"
    "strconv"  // Add this
    
    "github.com/danvergara/nostrich_watch_monitor/pkg/domain"
    "github.com/danvergara/nostrich_watch_monitor/pkg/services"
    "github.com/danvergara/nostrich_watch_monitor/web/views"
    "github.com/danvergara/nostrich_watch_monitor/web/views/components"  // Add this
)
```

## How It Works

1. **Initial Load**: Dashboard loads first 10 relays normally via `HandleRelayIndex`
2. **Scroll Detection**: When user scrolls to the "load-more-trigger" row, HTMX `revealed` trigger fires
3. **AJAX Request**: HTMX makes GET request to `/api/relays?offset=10&limit=10`
4. **Handler Response**: `HandleRelayRows` returns only new table rows
5. **Append Content**: HTMX appends new rows to `#relay-table-body` using `beforeend`
6. **Update Trigger**: New batch includes updated trigger with next offset
7. **Repeat**: Process continues until no more relays returned

## Benefits

✅ **Smooth UX**: No page reloads, seamless scrolling  
✅ **Performance**: Only loads data as needed  
✅ **SEO Friendly**: Initial content loads normally  
✅ **Lightweight**: Uses existing HTMX, no additional JS frameworks  
✅ **Flexible**: Easy to adjust page size and add filters later  
✅ **Database Efficient**: Uses proper OFFSET/LIMIT pagination  

## Testing Checklist

- [ ] Initial page load shows first 10 relays
- [ ] Scrolling to bottom triggers loading indicator
- [ ] New relays are appended to table
- [ ] Loading indicator shows/hides correctly
- [ ] No duplicate relays are loaded
- [ ] Handles empty results gracefully
- [ ] Works on mobile devices
- [ ] Performance is acceptable with large datasets

## Future Enhancements

- Add filters (relay type, status, etc.) to infinite scroll
- Implement virtual scrolling for very large datasets
- Add "Load More" button as fallback
- Cache loaded data in browser
- Add skeleton loading states

## Files to Modify

1. `internal/handlers/relays.go` - Add new handler method and import
2. `web/views/components/relay_table_rows.templ` - New template file
3. `web/views/components/relay_table.templ` - Modify existing template
4. `internal/server/routes.go` - Add new route
5. `web/static/css/styles.css` - Add HTMX indicator styles

## Notes

- Current implementation uses hardcoded limit of 10 in `HandleRelayIndex`
- Service layer already supports pagination via `RelayFilters`
- HTMX is already included in the project (`web/static/htmx.min.js`)
- Table structure is well-defined and ready for infinite scroll
- Relative time formatting was recently implemented and will work with new rows