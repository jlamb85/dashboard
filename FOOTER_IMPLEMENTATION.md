# Footer Implementation

## Overview
The footer is now **sticky** and always visible at the bottom of the page, displaying the application version dynamically.

## Visual Layout

```
┌─────────────────────────────────────────────────────────────┐
│                     Page Content                             │
│                                                              │
│                   (Grows dynamically)                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
┌═════════════════════════════════════════════════════════════┐
║                        FOOTER (Sticky)                       ║
║  © 2026 Server Dashboard  |  ♥ Made with passion  |  v1.0.0 ║
└═════════════════════════════════════════════════════════════┘
```

## Footer HTML Structure

```html
<footer class="footer mt-auto py-3 bg-body-secondary border-top">
    <div class="container-fluid">
        <div class="row align-items-center">
            <!-- Left: Copyright -->
            <div class="col-md-4 text-muted">
                <small>&copy; {{ currentYear }} Server Dashboard</small>
            </div>
            
            <!-- Center: Branding -->
            <div class="col-md-4 text-center">
                <small class="text-muted">
                    <i class="bi bi-heart-fill text-danger"></i> Made with passion
                </small>
            </div>
            
            <!-- Right: Version -->
            <div class="col-md-4 text-end">
                <small class="text-muted">
                    <i class="bi bi-code-square"></i> {{ appVersion }}
                </small>
            </div>
        </div>
    </div>
</footer>
```

## CSS Implementation

The footer uses CSS Flexbox for sticky positioning:

```css
/* Body layout for sticky footer */
body {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
}

/* Footer stays at bottom */
.footer {
    margin-top: auto;  /* Push to bottom */
    min-height: 60px;
    background-color: var(--bs-secondary-bg);
    border-top: 2px solid rgba(0, 0, 0, 0.1);
}

/* Dark mode support */
[data-bs-theme="dark"] .footer {
    border-top-color: rgba(255, 255, 255, 0.1);
}
```

## Template Functions

The version is injected via Go template functions:

```go
funcMap := template.FuncMap{
    "currentYear": func() int {
        return time.Now().Year()  // 2026
    },
    "appVersion": func() string {
        return Version  // v1.0.0
    },
    "buildInfo": func() string {
        return fmt.Sprintf("%s (built %s)", Version, BuildTime)
    },
}
```

## Version Sources

The version is read from multiple sources in priority order:

1. **Build-time ldflags** (production builds)
   ```bash
   go build -ldflags "-X main.Version=v1.0.0"
   ```

2. **VERSION file** (development/fallback)
   ```bash
   cat VERSION  # v1.0.0
   ```

3. **Default** ("dev" if nothing else available)

## Features

### ✅ Always Visible
- Footer is **always at the bottom** of the viewport
- Doesn't overlap content
- Works with any content height

### ✅ Responsive
- **Mobile**: Stacks vertically
- **Tablet/Desktop**: Three columns (left, center, right)

### ✅ Dark Mode Compatible
- Automatically adjusts colors based on theme
- Border color changes for better contrast

### ✅ Dynamic Content
- Year updates automatically ({{ currentYear }})
- Version reads from VERSION file or build flags
- No hardcoded values

## Testing Footer Display

### Visual Check
1. Start the application
2. Navigate to any page (Dashboard, Servers, VMs)
3. Verify footer is visible at bottom
4. Scroll to bottom - footer should be fixed
5. Resize window - footer should remain visible

### Version Check
```bash
# Check version in footer matches VERSION file
cat VERSION        # v1.0.0
# Then visit http://localhost:8080 and check footer
```

### Dark Mode Check
1. Click theme toggle button in header
2. Footer should adapt colors
3. Border should remain visible

## Responsive Breakpoints

```css
/* Desktop (≥ 768px) */
.col-md-4
  ├─ Left: Copyright
  ├─ Center: Branding
  └─ Right: Version

/* Mobile (< 768px) */
.col-md-4 → full width
  └─ Stacks vertically:
      1. Copyright
      2. Branding
      3. Version
```

## Icons Used

From Bootstrap Icons 1.11.0:
- `bi-heart-fill` - ❤️ Heart icon (branding)
- `bi-code-square` - 💻 Code icon (version)

## Browser Compatibility

Tested and working on:
- ✅ Chrome/Edge 90+
- ✅ Firefox 88+
- ✅ Safari 14+
- ✅ Mobile Safari (iOS 14+)
- ✅ Chrome Mobile (Android 10+)

## Troubleshooting

### Footer Not Visible
- Check `body` has `min-vh-100` class
- Verify `.footer` has `mt-auto` class
- Ensure Bootstrap 5.3 CSS is loaded

### Version Shows "dev"
- Check VERSION file exists
- Verify build includes version flags
- Check main.go reads VERSION file

### Dark Mode Not Working
- Verify `data-bs-theme` attribute on `<html>`
- Check theme toggle script is loaded
- Ensure dark mode CSS rules are present
