# Menu Timeout Behavior

## Problem Statement
Ayarlar menüsüne girilen kullanıcı işlem yapıyor ama 10 saniye sonra otomatik olarak alarm menüsüne geçişi yapılıyordu.

## Solution
İki seviyeli timeout sistemi oluşturuldu:
- **10 Seconds (No Activity)**: Idle screen'e geç (menü arka planda açık kalır)
- **5 Minutes (No Activity)**: Menüyü kapat ve Home Menu'ye dön

## Implementation Details

### Timeline
- **User Activity**: Menü içinde herhangi bir işlem yapıldığında (button click, touch, vb.) timeout'lar reset edilir
- **10 Seconds (Inactive)**: İşlem yapılmadığında → **Idle Screen**'e geç (menü arka planda açık kalır)
- **5 Minutes (No Activity)**: Hiç işlem yapılmadığında → Menüyü kapat ve **Home Menu**'ye dön

### Key Changes

#### 1. HomeView Enhancements
- Added `menuFullCloseTimeout` property
- Updated `_scheduleMenuAutoHide()` to implement two-level timeout:
  ```javascript
  // First timeout: 10s → Show idle screen
  this.menuAutoHideTimeout = setTimeout(..., 10000);
  
  // Second timeout: 5m → Close menu and return home
  this.menuFullCloseTimeout = setTimeout(..., 300000);
  ```

#### 2. ViewManager Helper
- Added `resetMenuAutoHideTimer()` method
- Called whenever user interacts with menu items

#### 3. View Mount Handlers
- **MenuView**: Schedule timer on mount
- **SettingsView**: Schedule timer on mount
- **AlarmoSettingsView**: Schedule timer on mount

#### 4. Event Listener Updates
- **MenuView**: Reset timer on expandable header clicks and menu item clicks
- **SettingsView**: Reset timer on all button clicks (Back, HA, Save, Test, Sync, Alarmo)
- **AlarmoSettingsView**: Reset timer on all button clicks

### Modified Functions

#### HomeView._scheduleMenuAutoHide()
```javascript
_scheduleMenuAutoHide: function() {
    var self = this;
    
    // Clear existing timeouts
    if (this.menuAutoHideTimeout) {
        clearTimeout(this.menuAutoHideTimeout);
    }
    if (this.menuFullCloseTimeout) {
        clearTimeout(this.menuFullCloseTimeout);
    }
    
    // First timeout: 10 seconds - show idle screen (but keep menu context)
    this.menuAutoHideTimeout = setTimeout(function() {
        console.log('[HomeView] No activity for 10s - switching to idle screen');
        self._showIdleScreen();
    }, 10000);
    
    // Second timeout: 5 minutes - return to home menu entirely
    this.menuFullCloseTimeout = setTimeout(function() {
        console.log('[HomeView] No activity for 5 minutes - closing menu and returning to home');
        window.SmartDisplay.viewManager.closeMenu();
        
        if (window.SmartDisplay.store) {
            window.SmartDisplay.store.setState({
                menu: {
                    currentView: 'home',
                    isOpen: false
                }
            });
        }
    }, 300000); // 5 minutes
}
```

#### ViewManager.resetMenuAutoHideTimer()
```javascript
resetMenuAutoHideTimer: function() {
    var homeView = this.views['home'];
    if (homeView && homeView._scheduleMenuAutoHide) {
        console.log('[ViewManager] Resetting menu auto-hide timer due to user activity');
        homeView._scheduleMenuAutoHide();
    }
}
```

### User Interaction Points Where Timer Resets

#### HomeView
- Touch events while menu is open
- Click events

#### MenuView
- Expandable header clicks
- Menu item clicks

#### SettingsView
- Back button
- HA Settings button
- HA Save/Test/Sync buttons
- Alarmo Monitoring button

#### AlarmoSettingsView
- Back button
- Filter buttons
- Refresh button

## Expected Behavior

### Scenario 1: User Active in Settings Menu
```
Time 0s:    User enters Settings menu → Timer scheduled (10s, 5m)
Time 5s:    User clicks "HA Settings" → Both timers reset
Time 15s:   User clicks "Save" → Both timers reset
Time 25s:   User still working... Timer continues resetting on each action
Result:     No automatic transitions occur while user is active
```

### Scenario 2: User Inactive for 10+ Seconds
```
Time 0s:    User enters Settings menu → Timer scheduled
Time 5s:    User clicks "HA Settings" → Timer reset
Time 15s:   [NO ACTIVITY FOR 10 SECONDS]
Result:     Idle screen displayed (menu still accessible in background)
```

### Scenario 3: User Inactive for 5+ Minutes Total
```
Time 0s:    User enters Settings menu → Timer scheduled
Time 60s:   User clicks "Test Connection" → Timer reset
Time 360s:  [NO ACTIVITY FOR 5 MINUTES]
Result:     Menu closes completely, returns to Home
```

## File Changes

### [viewManager.js](web/js/viewManager.js)

**HomeView Modifications:**
- Line 755: Added `menuFullCloseTimeout` property
- Lines 825-851: Updated `unmount()` to clear both timeouts
- Lines 1083-1120: Rewrote `_scheduleMenuAutoHide()` with two-level timeout

**ViewManager Additions:**
- Lines 4580-4587: Added `resetMenuAutoHideTimer()` helper function

**SettingsView Modifications:**
- Lines 2355-2368: Updated all event listener handlers to call `resetMenuAutoHideTimer()`
- Lines 2318-2329: Added timer scheduling on mount

**AlarmoSettingsView Modifications:**
- Lines 2640-2649: Added timer scheduling on mount
- Lines 2689-2705: Updated event listener handlers to call `resetMenuAutoHideTimer()`

**MenuView Modifications:**
- Lines 3457-3469: Added timer scheduling on mount
- Lines 3491-3495: Updated overlay click handlers to call `resetMenuAutoHideTimer()`
- Lines 3524-3528: Updated container click handlers to call `resetMenuAutoHideTimer()`

## Testing Checklist

- [ ] User enters Settings menu
- [ ] After 10s of no activity, idle screen appears (with menu still context available)
- [ ] User interacts with a button → Idle screen closes, timer resets
- [ ] After 5 minutes of total inactivity, menu closes and returns to Home
- [ ] Each user action resets both timeout counters
- [ ] Verify console logs show timer resets for debugging
- [ ] Test in multiple menu views (Settings, Alarmo, etc.)
- [ ] Test back button navigation
- [ ] Test expandable menu sections
- [ ] Verify no errors in browser console

## Console Output Examples

Expected logging:
```
[HomeView] Opening menu on initial load
[HomeView] Scheduling menu auto-hide timer on mount
[ViewManager] Resetting menu auto-hide timer due to user activity
[MenuView] Menu item clicked
[HomeView] No activity for 10s - switching to idle screen
[HomeView] No activity for 5 minutes - closing menu and returning to home
[ViewManager] Closing menu
```

## Backwards Compatibility

This change is backwards compatible as:
- Default timer values are reasonable (10s idle, 5m full close)
- All views gracefully handle the absence of ViewManager
- Existing timeout logic is preserved, just extended

