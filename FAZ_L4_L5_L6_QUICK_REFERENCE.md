# FAZ L4-L6 Integration Quick Reference

## How to Use Each Module

### L4: Advisor

**Initialization (automatic in bootstrap):**
```javascript
window.SmartDisplay.advisor.init();
```

**Manual hint check:**
```javascript
var hint = window.SmartDisplay.advisor.getHint({
    role: 'admin',
    alarmState: 'disarmed',
    haIsConnected: false,
    haIsConfigured: true,
    haSyncDone: false,
    guestIsActive: false,
    guestApprovedAt: null,
    currentView: 'home'
});
// Returns: { id: 'ha-disconnected', text: '...' } or null
```

**Show context-aware hint:**
```javascript
window.SmartDisplay.advisor.checkAndShow(context);
```

**Manual trigger (for testing):**
```javascript
window.SmartDisplay.advisor.showManual('This is a test hint');
```

**Dismiss current hint:**
```javascript
window.SmartDisplay.advisor.dismiss();
```

---

### L5: Intro

**Check if intro should play:**
```javascript
if (window.SmartDisplay.intro.shouldShow()) {
    // firstBoot === true and not yet shown
}
```

**Play intro sequence:**
```javascript
window.SmartDisplay.intro.play()
    .then(() => console.log('Intro complete'))
    .catch(err => console.error('Intro failed', err));
```

**Skip intro immediately:**
```javascript
window.SmartDisplay.intro.skip();
```

**Automatic flow (bootstrap):**
```javascript
// Bootstrap automatically:
// 1. Checks if should show
// 2. Plays if yes
// 3. Skips on error
// 4. Continues to LoginView
```

---

### L6: Trace

**Initialization (automatic in bootstrap):**
```javascript
window.SmartDisplay.trace.init();
```

**Add admin action:**
```javascript
window.SmartDisplay.trace.add('HA credentials saved');
window.SmartDisplay.trace.add('HA connection verified');
window.SmartDisplay.trace.add('Guest access ended');
```

**Show/hide trace UI:**
```javascript
window.SmartDisplay.trace.show();
window.SmartDisplay.trace.hide();
```

**Clear all entries:**
```javascript
window.SmartDisplay.trace.clear();
```

---

## Store State Access

### Get Full State
```javascript
var state = window.SmartDisplay.store.getState();

console.log(state.aiAdvisorState.currentHint);
console.log(state.adminTrace.recent);
```

### Subscribe to Changes
```javascript
var unsubscribe = window.SmartDisplay.store.subscribe(function(updates) {
    if (updates.aiAdvisorState) {
        console.log('Advisor state changed');
    }
    if (updates.adminTrace) {
        console.log('Trace updated');
    }
});

// Later: unsubscribe();
```

---

## Adding New Trace Entries

### Pattern: Only Admin Actions
```javascript
// Check role before adding trace
var state = window.SmartDisplay.store.getState();
if (state.authState.role === 'admin' && window.SmartDisplay.trace) {
    window.SmartDisplay.trace.add('Your action description');
}
```

### Where to Add
- **Settings:** HA credentials save, HA test success, sync completion
- **Guest:** When admin approves/ends guest access
- **Alarm:** (NOT - alarm is never admin's domain)
- **Any other admin-only control**

---

## Adding New Advisor Contexts

### Edit advisor.js getHint() function
```javascript
getHint: function(context) {
    // ... existing checks ...
    
    // Add new context check:
    if (context.someCondition) {
        return {
            id: 'unique-hint-id',
            text: 'Single sentence hint.'
        };
    }
    
    return null;
}
```

### Context Parameters Available
```javascript
{
    role: 'admin' | 'user' | 'guest',
    alarmState: 'unknown' | 'disarmed' | 'armed' | 'pending' | 'arming' | 'triggered',
    haIsConnected: boolean,
    haIsConfigured: boolean,
    haSyncDone: boolean,
    guestIsActive: boolean,
    guestApprovedAt: timestamp or null,
    currentView: 'home' | 'alarm' | 'settings' | ...
}
```

---

## CSS Customization

### Advisor Bubble Colors
```css
.advisor-bubble {
    background-color: rgba(33, 150, 243, 0.95); /* Blue */
    /* Change to your color */
}
```

### Intro Animation Speed
```css
.intro-glow {
    animation: intro-glow-pulse 1.5s ease-in-out infinite;
    /* Change 1.5s to different duration */
}
```

### Trace Entry Colors
```css
.trace-entry {
    background-color: rgba(76, 175, 80, 0.1); /* Light green */
    border-left-color: rgba(76, 175, 80, 0.6); /* Green accent */
    color: #2e7d32; /* Dark green text */
}
```

---

## Common Issues & Solutions

### Advisor not appearing
- Check: Is user admin role?
- Check: Is alarm triggered? (hidden during lockdown)
- Check: Did hint get dismissed? (6s timeout)
- Debug: `console.log(window.SmartDisplay.advisor.getHint(context))`

### Intro playing every time
- Check: `firstBoot` state is being set to false after completion
- Check: Browser console for errors during playback
- Check: Reduced-motion preference is working

### Trace not showing
- Check: Admin role confirmed
- Check: trace.add() called after authentication
- Check: CSS display not overridden
- Debug: `console.log(window.SmartDisplay.store.getState().adminTrace)`

---

## Testing Checklist

```javascript
// Test Advisor
window.SmartDisplay.advisor.showManual('Test hint');
// Should see blue bubble, bottom-right, disappear in 6s

// Test Intro
window.SmartDisplay.intro.play();
// Should see full-screen sequence (or static if prefers-reduced-motion)

// Test Trace
window.SmartDisplay.trace.add('Test entry');
window.SmartDisplay.trace.show();
// Should see green entry, bottom-left

// Test State
var state = window.SmartDisplay.store.getState();
console.log(state.aiAdvisorState);
console.log(state.adminTrace.recent);
```

---

## Performance Notes

- **Advisor:** Pure function, O(1) hint generation
- **Intro:** One-time 2-3s animation, then removed from DOM
- **Trace:** Max 5 entries in memory, minimal overhead
- **Overall:** <1KB combined JS, CSS fits in main stylesheet

---

## Accessibility

All three modules respect:
- `prefers-reduced-motion` media query
- Role-based access (guest/user suppression)
- Graceful failure (non-blocking on error)
- Keyboard/touch neutral (no required interaction)

---

## Future Enhancement Ideas

### L4 Advisor
- [ ] Contextual hints for guest session pending
- [ ] Hints for alarm armed/triggered transitions
- [ ] Hint throttling to prevent spam
- [ ] User preference to disable advisor

### L5 Intro
- [ ] Customizable intro text from backend
- [ ] Brand/logo customization
- [ ] Intro skip button (kiosk mode)

### L6 Trace
- [ ] Time-based expiry (auto-clear after 30min)
- [ ] Trace export for debugging
- [ ] Configurable max entries
- [ ] Optional backend sync for audit trail

---

**Documentation Last Updated:** January 5, 2026
**Implementation Status:** Production Ready ✓
