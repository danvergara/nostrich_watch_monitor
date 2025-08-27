# Infinite Scroll Implementation Plan

## Overview
Implement infinite scroll pagination for the relays table using HTMX to load more relays as the user scrolls down, fetching data from the database using OFFSET and LIMIT.

## Current State
- ✅ Basic table structure exists (`RelayTable` component)
- ✅ Pagination logic exists in service layer (`RelayFilters` with Limit/Offset)
- ✅ HTMX is already included in the project
- ✅ Handler currently loads first 10 relays with hardcoded limit/offset

## Error Handling Strategy

### Problem with Current Approach
The current error handling uses `http.Error()` which completely replaces the page with plain text error messages:
```go
if err != nil {
    http.Error(w, "Error fetching relays", http.StatusInternalServerError)
    return
}
```

This breaks the user experience by:
- **Destroying the UI**: Replaces entire dashboard with error text
- **Losing Context**: Users lose navigation, previously loaded content
- **Poor Infinite Scroll UX**: API failures completely break the page

### Improved Error Handling
Instead of breaking the page, we'll use **graceful error handling**:
- **Preserve UI Structure**: Errors appear within the table layout
- **Contextual Messages**: Show specific, actionable error information
- **Retry Mechanisms**: Provide easy ways to recover from failures
- **Non-Destructive**: Keep existing content visible

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
        // Return error row instead of breaking the page
        components.ErrorRow("Failed to load more relays. Please try again.").Render(r.Context(), w)
        return
    }
    
    // Return only the table rows, not the full page
    // Use graceful error handling instead of http.Error
    if err := components.RelayTableRows(ToRelayTableViewModels(relays), offset+limit).Render(r.Context(), w); err != nil {
        // Return error row that preserves table structure
        components.ErrorRow("Failed to render relay data. Please try again.").Render(r.Context(), w)
        return
    }
}
```

### 2. Create Error Handling Templates

#### 2a. Error Row Template
**File**: `web/views/components/error_row.templ`
**Status**: Pending

Template for inline error handling within table structure:

```templ
package components

templ ErrorRow(message string) {
    <tr class="bg-red-50 dark:bg-red-900/20">
        <td colspan="6" class="px-4 py-8 text-center">
            <div class="flex flex-col items-center space-y-3">
                <div class="text-red-600 dark:text-red-400">
                    <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                    </svg>
                </div>
                <p class="text-red-600 dark:text-red-400 font-medium">{message}</p>
                <button 
                    class="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
                    hx-get="/api/relays?offset=0&limit=10" 
                    hx-target="#relay-table-body" 
                    hx-swap="innerHTML">
                    Try Again
                </button>
            </div>
        </td>
    </tr>
}
```

#### 2b. Empty State Template
**File**: `web/views/components/empty_state.templ`
**Status**: Pending

Template for when no relays are available:

```templ
package components

templ EmptyState(message string) {
    <tr>
        <td colspan="6" class="px-4 py-12 text-center text-gray-500 dark:text-gray-400">
            <div class="flex flex-col items-center space-y-3">
                <svg class="w-12 h-12 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2 2v-5m16 0h-2M4 13h2m0 0V9a2 2 0 012-2h2m0 0V6a2 2 0 012-2h2.01"></path>
                </svg>
                <p class="text-lg font-medium">{message}</p>
            </div>
        </td>
    </tr>
}
```

### 3. Create RelayTableRows Template
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
    if len(relays) == 0 {
        @EmptyState("No relays available at the moment.")
    } else {
        for _, relay := range relays {
            @RelayTableRow(relay)
        }
        
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

### 4. Modify RelayTable Template
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

### 5. Improve Dashboard Error Handling
**File**: `internal/handlers/relays.go`
**Status**: Pending

Update `HandleRelayIndex` to preserve UI on errors by using existing dashboard with empty table:

```go
func (rh *RelaysHandler) HandleRelayIndex(w http.ResponseWriter, r *http.Request) {
    limit := 10
    offset := 0
    filters := &services.RelayFilters{
        Limit:  &limit,
        Offset: &offset,
    }

    relays, err := rh.service.GetRelays(r.Context(), filters)
    if err != nil {
        // Show dashboard with empty table - template will show error state via EmptyState component
        if err := views.Dashboard([]presentation.RelayTableViewModel{}).Render(r.Context(), w); err != nil {
            // If even empty dashboard fails, try once more (template rendering rarely fails twice)
            views.Dashboard([]presentation.RelayTableViewModel{}).Render(r.Context(), w)
        }
        return
    }

    if err := views.Dashboard(ToRelayTableViewModels(relays)).Render(r.Context(), w); err != nil {
        // Same approach - show empty dashboard instead of breaking the page
        views.Dashboard([]presentation.RelayTableViewModel{}).Render(r.Context(), w)
    }
}
```

**Key Benefits of This Approach:**
- ✅ **Reuses Existing UI**: Navigation, styling, layout all preserved
- ✅ **Consistent Experience**: Looks exactly like normal dashboard  
- ✅ **Template Logic**: Empty state handling built into RelayTable/RelayTableRows components
- ✅ **Simple**: No new templates or HTML needed
- ✅ **Maintainable**: Leverages existing EmptyState component

**How It Works:**
1. When `GetRelays()` fails, pass empty array to Dashboard
2. Dashboard renders normally with navigation and layout
3. RelayTable receives empty array and shows EmptyState component
4. User sees professional error message within familiar UI
5. No broken pages, no lost context

### 6. Improve Relay Detail Error Handling
**File**: `internal/handlers/relays.go`
**Status**: Pending

Update `HandleRelayDetail` to preserve UI on errors by showing error state within existing template:

```go
func (rh *RelaysHandler) HandleRelayDetail(w http.ResponseWriter, r *http.Request) {
    relayURL := r.URL.Query().Get("url")
    if relayURL == "" {
        http.Error(w, "URL parameter required", http.StatusBadRequest)
        return
    }

    relay, err := rh.service.GetRelayByURL(r.Context(), relayURL)
    if err != nil {
        // Create error state relay view instead of breaking the page
        errorRelay := createErrorRelayViewModel(relayURL, "Relay not found or temporarily unavailable")
        if err := views.RelayDetail(errorRelay).Render(r.Context(), w); err != nil {
            // Final fallback: try once more
            views.RelayDetail(errorRelay).Render(r.Context(), w)
        }
        return
    }

    if err := views.RelayDetail(ToRelayDetailViewModel(relay)).Render(r.Context(), w); err != nil {
        // Same approach - show error state instead of breaking
        errorRelay := createErrorRelayViewModel(relayURL, "Error loading relay details")
        views.RelayDetail(errorRelay).Render(r.Context(), w)
    }
}

// Helper function to create error state relay
func createErrorRelayViewModel(url, errorMessage string) presentation.RelayDetailViewModel {
    return presentation.RelayDetailViewModel{
        URL:            url,
        Name:           "Relay Unavailable",
        Description:    errorMessage,
        IsOnline:       false,
        Classification: "Unknown",
        LastCheckTime:  "Never",
        // All other fields will be zero values, template should handle gracefully
    }
}
```

**Key Benefits of This Approach:**
- ✅ **Consistent UX**: User stays on detail page with familiar layout
- ✅ **Clear Feedback**: Shows exactly what went wrong in description field
- ✅ **Navigation Preserved**: Back button and navigation still work
- ✅ **SEO Friendly**: Still a valid page, not a redirect or broken state
- ✅ **Template Reuse**: Uses existing RelayDetail template with error data

**How It Works:**
1. When `GetRelayByURL()` fails, create error state relay view model
2. RelayDetail template renders normally with error data
3. User sees "Relay Unavailable" with error message in description
4. All UI elements (navigation, back button) remain functional
5. Template handles zero values gracefully (shows "N/A" or hides sections)

### 7. Update Routes
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

### 8. Add CSS for Loading Indicator and Error States
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

/* Error state styles */
.error-row {
    background-color: rgba(239, 68, 68, 0.1);
}

.error-row:hover {
    background-color: rgba(239, 68, 68, 0.15);
}
```

### 9. Add Required Import
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
✅ **Graceful Error Handling**: Errors don't break the UI  
✅ **Professional Empty States**: Clear feedback when no data available  
✅ **User-Friendly Recovery**: Easy retry mechanisms for failed requests  

## Testing Checklist

### Core Functionality
- [ ] Initial page load shows first 10 relays
- [ ] Scrolling to bottom triggers loading indicator
- [ ] New relays are appended to table
- [ ] Loading indicator shows/hides correctly
- [ ] No duplicate relays are loaded
- [ ] Works on mobile devices
- [ ] Performance is acceptable with large datasets

### Error Handling
- [ ] Database connection failures show graceful error (not broken page)
- [ ] API endpoint failures show inline error rows with retry buttons
- [ ] Empty results show professional empty state
- [ ] Retry buttons work correctly
- [ ] Network timeouts are handled gracefully
- [ ] Malformed responses don't break the UI

### Edge Cases
- [ ] Very slow network connections
- [ ] Rapid scrolling doesn't cause duplicate requests
- [ ] Browser back/forward navigation works correctly
- [ ] Page refresh maintains expected state

## Future Enhancements

- Add filters (relay type, status, etc.) to infinite scroll
- Implement virtual scrolling for very large datasets
- Add "Load More" button as fallback
- Cache loaded data in browser
- Add skeleton loading states

## Files to Modify

### New Files
1. `web/views/components/error_row.templ` - Error handling template
2. `web/views/components/empty_state.templ` - Empty state template  
3. `web/views/components/relay_table_rows.templ` - Infinite scroll rows template

### Modified Files
4. `internal/handlers/relays.go` - Add new handler method, improve error handling for both dashboard and detail pages, add imports and helper function
5. `web/views/components/relay_table.templ` - Modify to use new components
6. `internal/server/routes.go` - Add new API route
7. `web/static/css/styles.css` - Add HTMX indicator and error state styles

## Notes

- Current implementation uses hardcoded limit of 10 in `HandleRelayIndex`
- Service layer already supports pagination via `RelayFilters`
- HTMX is already included in the project (`web/static/htmx.min.js`)
- Table structure is well-defined and ready for infinite scroll
- Relative time formatting was recently implemented and will work with new rows
- **Critical**: Current error handling breaks UX by replacing entire page with error text
- **Improvement**: New approach preserves UI structure and provides actionable error recovery
- **User Experience**: Empty states and error handling create professional, polished interface
- **Detail Page**: Same error handling approach applied to relay detail page for consistency
- **Template Reuse**: Both dashboard and detail pages leverage existing templates for error states