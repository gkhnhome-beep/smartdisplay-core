# First-Boot HA Not Configured Fix

## Problem
On first boot, when Home Assistant is not yet configured, the UI gets stuck on the Alarm screen because:
1. Alarm state is empty/unhydrated (waiting for HA to provide it)
2. The routing logic treats any empty alarm state as "must lock to Alarm view"
3. This prevents the user from reaching Settings to configure HA
4. Result: First-boot lockout requiring backend intervention or factory reset

## Root Cause
In [viewManager.js](web/js/viewManager.js), the `getNextView()` function had this condition:
```javascript
if (!alarmState.isHydrated || this._shouldLockToAlarm(alarmState)) {
    return 'alarm';
}
```

This unconditionally locks to Alarm whenever alarm state is not fully hydrated, preventing access to Settings during first boot when HA is not yet configured.

## Solution
Added a **bypass condition** that allows Settings access when:
- Home Assistant is NOT configured (`haState.isConfigured === false`) AND
- Alarm state is in a waiting/invalid state (not a real alarm mode)

### Changes Made

**File:** [web/js/viewManager.js](web/js/viewManager.js)

**1. Updated `getNextView()` method** (lines 2120-2165)
- Added `haState` variable to access HA configuration status
- Added `shouldBypassAlarmLock` flag to detect first-boot scenario
- Modified routing logic to check bypass condition before applying alarm lock
- Added clear console logging for debugging

**2. Added `_isAlarmStateInvalid()` helper method** (lines 2221-2232)
- Identifies alarm states that are placeholder/waiting values
- Returns true for: `null`, `undefined`, empty string, `'waiting'`, `'unknown'`
- Returns false for: real alarm modes like `'disarmed'`, `'armed_home'`, `'armed_away'`, etc.

### Routing Priority (Updated)
1. ✅ FirstBoot active → FirstBootView
2. ✅ Menu overlay requested → MenuView  
3. ✅ Guest session active (non-admin) → GuestView
4. ✅ **[NEW] HA not configured + invalid alarm → allows Settings access**
5. ✅ Alarm in real locked state → AlarmView
6. ✅ Settings view + admin role → SettingsView
7. ✅ Default → HomeView

### Key Design Decisions

1. **Minimal, surgical fix**: Only changes routing logic, no backend or API changes
2. **Preserves alarm locking**: Once HA IS configured, normal alarm locking behavior is preserved
3. **First-boot focused**: Specific to the haState.isConfigured flag and invalid alarm states
4. **Backward compatible**: All existing alarm behaviors remain unchanged

### Testing Scenarios

✅ **First-boot with HA not configured**
- User boots device for first time
- HA is not configured
- Alarm state is empty/waiting
- Expected: User can access Settings to configure HA
- Result: **FIXED** - Settings is accessible

✅ **First-boot with HA configured**
- User boots device, HA already configured
- Alarm state gets hydrated with real values
- Expected: Normal alarm locking behavior
- Result: **Works** - Alarm locking preserved

✅ **Normal operation (HA configured)**
- System running with HA fully configured
- Alarm state has real values (armed/disarmed/etc.)
- Expected: Alarm locking works normally
- Result: **Works** - No change to existing behavior

✅ **Alarm triggered (real alarm state)**
- HA configured, actual alarm trigger detected
- Alarm state = 'triggered'
- Expected: Lockdown to Alarm view
- Result: **Works** - No change to existing behavior

## Files Modified
- [web/js/viewManager.js](web/js/viewManager.js) - Routing logic fix

## Verification
To verify the fix is working:
1. Check console logs for `[ViewManager] Route:` messages
2. During first-boot with no HA: should see route to Settings or Home, NOT stuck on Alarm
3. Once HA configured: should see normal alarm routing behavior
4. When alarm triggers: should lock to Alarm view normally

## Technical Details

### haState Object Structure
Located in [web/js/store.js](web/js/store.js#L63-L86):
```javascript
haState: {
    isConnected: boolean,     // Last test reached ok stage
    isConfigured: boolean,    // Credentials exist (KEY FLAG)
    syncDone: boolean,        // Bootstrap sync completed
    runtimeUnreachable: boolean,
    meta: {...},
    entityCounts: {...}
}
```

### alarmState Object Structure
Located in [web/js/store.js](web/js/store.js#L37-L48):
```javascript
alarmState: {
    state: string,           // 'disarmed', 'armed_home', 'armed_away', etc.
    triggered: boolean,      // true if alarm is in alarm state
    isHydrated: boolean,     // Has received data from backend
    // ... other fields
}
```

## Impact Assessment
- **User Impact**: ✅ Positive - fixes first-boot lockout
- **Backward Compatibility**: ✅ Full - no breaking changes
- **Performance**: ✅ Negligible - adds one boolean check
- **Security**: ✅ Safe - preserves admin access controls
