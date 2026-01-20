# Dashboard UI Features Quick Reference

## Navigation

### Main Pages
- **Dashboard** - Overview with summary cards and system status
- **Servers** - List of all configured servers with real-time monitoring
- **Virtual Machines** - List of all configured VMs with real-time monitoring

### Navigation Methods
1. **Click sidebar links** - Desktop view (left sidebar)
2. **Click hamburger menu** - Mobile view (top-right)
3. **Direct URL** - Navigate directly to `/servers`, `/vms`, etc.

### Collapsible Sidebar (Desktop Only)
1. Click the **sidebar toggle icon** (📋) next to the theme toggle
2. Sidebar smoothly slides out of view
3. Content area expands to full width
4. Click again to restore sidebar
5. **Preference saved** - Your choice persists across sessions

**Features:**
- Smooth 0.3s animations
- State saved to localStorage
- Icon changes to indicate collapsed/expanded state
- Only visible on large screens (desktop)

---

## Dark Mode

### Enable Dark Mode
1. Click the **moon icon** (🌙) in the top-right corner
2. Page theme switches to dark mode
3. Preference automatically saved

### Disable Dark Mode
1. Click the **sun icon** (☀️) in the top-right corner
2. Page returns to light mode
3. Preference automatically saved

### Auto-Detection
- Dashboard automatically detects your OS dark mode preference
- First-time users get their OS preference
- Toggle button overrides the automatic detection

---

## System Status

### Status Indicators
Located in the sidebar (left side):

| Status | Color | Meaning |
|--------|-------|---------|
| **Active** | Green | All systems monitoring normally |
| **Degraded** | Yellow | Some systems unreachable |
| **Offline** | Red | Monitoring stopped or error |

### Real-Time Updates
- Status checks every 30 seconds
- Auto-refresh without manual intervention
- Smooth color transitions

---

## Tables

### Servers & VMs Tables

#### Columns
- **Name** - System name (bold)
- **IP Address** - IP or hostname (monospace font)
- **Status** - Online/Offline/Running (badge with icon)
- **Uptime** - System uptime duration
- **Processes** - Number of running processes
- **Disk** - Disk usage percentage with color coding
- **Last Checked** - Timestamp of last update

#### Color-Coded Disk Usage
- **Green**: < 70% usage (healthy)
- **Yellow**: 70-80% usage (caution)
- **Red**: > 80% usage (warning)

#### Interactive Features
- **Click any row** - View detailed metrics
- **Keyboard**: Press **Enter** on a row to view details
- **Hover effect** - Rows highlight on hover
- **Responsive** - Tables adapt to screen size

---

## Detail Pages

### What You'll See
- System name and IP address
- Real-time status
- System metrics:
  - Uptime
  - Running processes
  - Disk space usage
  - System load
- Port connectivity:
  - SSH (port 22)
  - HTTP (port 80)
  - HTTPS (port 443)
  - MySQL (port 3306)
  - PostgreSQL (port 5432)

### Port Status
- **Green**: Port is open/responsive
- **Red**: Port is closed/unreachable
- **Gray**: Not checked yet

### Refresh Data
- Data refreshes automatically every 30 seconds
- No manual refresh needed
- Check "Last Updated" timestamp

---

## Cards & Metrics

### Card Features
- **Shadow effect** - Depth and elevation
- **Hover animation** - Cards lift slightly on hover
- **Status badge** - Quick status at a glance
- **Pulsing indicator** - Shows live monitoring

### Metric Display
- **Large numbers** - Easy to read metrics
- **Units** - Clear labeling (%, hours, processes)
- **Color-coded** - Status clear from colors
- **Icons** - Visual indicators for quick scanning

---

## Notifications & Alerts

### Notification Types
1. **Success** - Green - Operation completed successfully
2. **Warning** - Yellow - Attention needed
3. **Error** - Red - Something went wrong
4. **Info** - Blue - Informational message

### Notification Behavior
- Auto-dismiss after 3 seconds
- Click to dismiss immediately
- No pop-up windows (non-intrusive)

---

## Mobile Experience

### Mobile Features
- **Responsive layout** - Optimized for touch devices
- **Hamburger menu** - Access navigation on small screens
- **Stacked tables** - Tables adapt to narrow screens
- **Touch-friendly** - Larger tap targets
- **Full functionality** - All features work on mobile

### Mobile Navigation
1. Tap **hamburger menu** (≡) in top-left
2. Select page from dropdown menu
3. Menu automatically closes after selection

---

## Icons Used

### Navigation Icons
- 🏠 **Dashboard** - Home dashboard
- 🖥️ **Servers** - Physical or cloud servers
- 🖲️ **Virtual Machines** - VMs and containers
- ⚙️ **Settings** - Configuration options

### Status Icons
- 🟢 **Online/Running** - System is operational
- 🔴 **Offline** - System is unreachable
- ⏱️ **Clock** - Time-related info
- 📊 **Chart** - Metrics and data

### Action Icons
- 🌙 **Moon** - Enable dark mode
- ☀️ **Sun** - Enable light mode
- 🔍 **Search** - Search functionality
- ⚡ **Lightning** - Quick actions

---

## Keyboard Shortcuts

### Navigation
| Key | Action |
|-----|--------|
| <kbd>Tab</kbd> | Move to next element |
| <kbd>Shift+Tab</kbd> | Move to previous element |
| <kbd>Enter</kbd> | Activate focused button/link |

### Table Navigation
| Key | Action |
|-----|--------|
| <kbd>Tab</kbd> | Move to next row |
| <kbd>Enter</kbd> | View details for row |
| <kbd>Space</kbd> | Toggle row selection (if available) |

### Theme Toggle
| Key | Action |
|-----|--------|
| <kbd>M</kbd> | Toggle theme (if implemented) |

---

## Performance Tips

1. **Use dark mode in low light** - Reduces eye strain
2. **Allow auto-refresh** - Don't manually refresh
3. **Close unused tabs** - Reduces memory usage
4. **Clear browser cache** - Improves performance
5. **Use modern browser** - Better compatibility

---

## Troubleshooting

### Table not loading?
1. Wait 30 seconds for auto-refresh
2. Check network connection
3. Reload the page (F5)
4. Check browser console (F12) for errors

### Dark mode not saving?
1. Enable localStorage in browser
2. Clear browser cache
3. Try different browser
4. Check if cookies are allowed

### Icons not showing?
1. Check internet connection
2. Wait for Bootstrap Icons CDN to load
3. Hard refresh page (Ctrl+Shift+R or Cmd+Shift+R)
4. Check browser console for errors

### Clickable rows not working?
1. Ensure JavaScript is enabled
2. Check browser console (F12)
3. Try different browser
4. Reload the page

---

## Accessibility Features

### Screen Reader Support
- Semantic HTML for proper structure
- ARIA labels for custom components
- Meaningful link text
- Form labels properly associated

### Keyboard Navigation
- All features accessible via keyboard
- Focus indicators visible
- Logical tab order
- No keyboard traps

### Color Contrast
- WCAG AA compliant
- Text readable on all backgrounds
- Not relying solely on color
- Alternative indicators provided

### Responsive Design
- Works on all screen sizes
- Touch-friendly on mobile
- Readable at all zoom levels
- Font sizes appropriate

---

## Getting Help

### In-App Help
- Hover over elements for tooltips
- Status badges explain meaning
- Descriptive column headers
- Clear section labels

### Documentation
- See `UI-MODERNIZATION.md` for full guide
- Check `README.md` for project overview
- View production guides for deployment

### Error Messages
- Clear error descriptions
- Suggested solutions
- Actionable next steps
- Contact information if needed

---

## System Requirements

### Browser Requirements
- Modern browser (Chrome, Firefox, Safari, Edge)
- JavaScript enabled
- Cookies/localStorage enabled
- Viewport support (mobile-friendly)

### Network Requirements
- Stable internet connection
- Access to dashboard server
- No proxy/firewall blocking
- API endpoint reachability

---

## Tips & Tricks

1. **Bookmark detail pages** - Direct access to specific systems
2. **Use browser back button** - Navigate back from detail pages
3. **Refresh at any time** - Manual refresh with F5
4. **Zoom in/out** - Adjust with Ctrl+/- (Cmd+/- on Mac)
5. **Full screen** - Press F11 for full-screen dashboard
6. **Open in new tab** - Right-click to open detail pages in new tabs
7. **Save page offline** - Save HTML for offline reference

---

## Version Info

**Dashboard Version**: 2.0  
**UI Framework**: Bootstrap 5.3  
**Last Updated**: 2026  
**Status**: Production Ready
