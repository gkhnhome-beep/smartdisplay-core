// --- Role-based UI ---
// --- Backend/HA Error Handling ---
const backendErrorBanner = document.getElementById('backend-error-banner');
let backendUnreachable = false;
let backendRetryTimer = null;
let debug = false; // set true for console logs

function showBackendErrorBanner(show) {
  if (!backendErrorBanner) return;
  backendErrorBanner.style.display = show ? '' : 'none';
}

function handleApiError(err, context) {
  if (debug) console.error('API error:', context, err);
  if (typeof err === 'string') return err;
  if (err && err.status === 404) return 'Not found.';
  if (err && err.status === 401) return 'Not authorized.';
  if (err && err.status === 500) return 'Server error.';
  return 'Something went wrong.';
}

function checkBackendHealth() {
  fetch('/api/health').then(r=>{
    if (!r.ok) throw new Error('Backend unreachable');
    return r.json();
  }).then(()=>{
    backendUnreachable = false;
    showBackendErrorBanner(false);
    if (backendRetryTimer) clearTimeout(backendRetryTimer);
  }).catch(()=>{
    backendUnreachable = true;
    showBackendErrorBanner(true);
    backendRetryTimer = setTimeout(checkBackendHealth, 5000);
  });
}
// Start health check loop
setTimeout(checkBackendHealth, 1000);
// --- Idle/Kiosk Mode Logic ---
const idleScreen = document.getElementById('idle-screen');
const idleClock = document.getElementById('idle-clock');
const idleDate = document.getElementById('idle-date');
const idleAlarm = document.getElementById('idle-alarm');
const idleAI = document.getElementById('idle-ai');
let idleTimer = null;
let isIdle = false;
let idlePollInterval = null;
let normalPollInterval = 2000;
let idlePollIntervalMs = 10000;
let lastHomeData = null;
let lastOverviewData = null;
let pollTimeout = null;
function debouncePoll(fn, delay) {
  if (pollTimeout) clearTimeout(pollTimeout);
  pollTimeout = setTimeout(fn, delay);
}

function resetIdleTimer() {
  if (isIdle) return;
  if (idleTimer) clearTimeout(idleTimer);
  idleTimer = setTimeout(goIdle, 60000);
}

function goIdle() {
  isIdle = true;
  if (idleScreen) idleScreen.style.display = '';
  document.body.classList.add('dimmed');
  updateIdleScreen();
  if (idlePollInterval) clearInterval(idlePollInterval);
  idlePollInterval = setInterval(updateIdleScreen, idlePollIntervalMs);
}

function wakeFromIdle() {
  if (!isIdle) return;
  isIdle = false;
  if (idleScreen) idleScreen.style.display = 'none';
  document.body.classList.remove('dimmed');
  if (idlePollInterval) clearInterval(idlePollInterval);
  resetIdleTimer();
  showView('home');
}

function updateIdleScreen() {
  // Large clock and date
  try {
    if (idleClock) {
      const now = new Date();
      idleClock.textContent = now.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
    }
    if (idleDate) {
      const now = new Date();
      idleDate.textContent = now.toLocaleDateString(undefined, {weekday:'long', year:'numeric', month:'long', day:'numeric'});
    }
    // Alarm state and AI summary
    fetch('/api/ui/home').then(r=>r.ok?r.json():Promise.reject()).then(data=>{
      if (JSON.stringify(data) !== JSON.stringify(lastHomeData)) {
        lastHomeData = data;
        if (idleAlarm) idleAlarm.textContent = data.alarm_state ? `Alarm: ${data.alarm_state}` : '';
        if (idleAI) idleAI.textContent = data.ai_insight && data.ai_insight.summary ? data.ai_insight.summary.split('.').slice(0,1).join('.') : '';
      }
    }).catch(()=>{
      if (idleAlarm) idleAlarm.textContent = '';
      if (idleAI) idleAI.textContent = '';
    });
  } catch (e) {
    if (debug) console.error('Idle screen update error', e);
  }
}

// Reset idle timer on any user interaction
['touchstart','mousedown','mousemove','keydown'].forEach(evt=>{
  window.addEventListener(evt, resetIdleTimer, {passive:true});
});
// Wake from idle on any touch/click
if (idleScreen) {
  idleScreen.addEventListener('touchstart', wakeFromIdle);
  idleScreen.addEventListener('mousedown', wakeFromIdle);
}
resetIdleTimer();
// --- Guest Flow UI Logic ---
const guestOverlay = document.getElementById('guest-overlay');
const guestRequestBtn = document.getElementById('guest-request-btn');
const guestWaitingSection = document.getElementById('guest-waiting-section');
const guestWaitingCountdown = document.getElementById('guest-waiting-countdown');
const guestWaitingStatus = document.getElementById('guest-waiting-status');
const guestRequestSection = document.getElementById('guest-request-section');
const guestApprovedSection = document.getElementById('guest-approved-section');
const guestDeniedSection = document.getElementById('guest-denied-section');
const guestDeniedMsg = document.getElementById('guest-denied-msg');
const guestLeaveBtn = document.getElementById('guest-leave-btn');

let guestState = 'idle'; // idle | waiting | approved | denied
let guestCountdown = null;

function showGuestOverlay(state, opts={}) {
  if (!guestOverlay) return;
  guestOverlay.style.display = '';
  guestRequestSection.style.display = (state === 'idle') ? '' : 'none';
  guestWaitingSection.style.display = (state === 'waiting') ? '' : 'none';
  guestApprovedSection.style.display = (state === 'approved') ? '' : 'none';
  guestDeniedSection.style.display = (state === 'denied') ? '' : 'none';
  if (state === 'waiting') {
    guestWaitingCountdown.textContent = opts.countdown ? `Time left: ${opts.countdown}s` : '';
    guestWaitingStatus.textContent = opts.status || '';
  }
  if (state === 'denied') {
    guestDeniedMsg.textContent = opts.status || '';
  }
}

function hideGuestOverlay() {
  if (guestOverlay) guestOverlay.style.display = 'none';
}

if (guestRequestBtn) {
  guestRequestBtn.onclick = function() {
    fetch('/api/guest/request', {method:'POST'}).then(r=>{
      showGuestOverlay('waiting', {status:'Request sent. Waiting for approval...'});
      guestState = 'waiting';
    }).catch(()=>{
      showToast('Failed to request access', '#b71c1c');
    });
  };
}
if (guestLeaveBtn) {
  guestLeaveBtn.onclick = function() {
    fetch('/api/guest/exit', {method:'POST'}).then(r=>{
      guestState = 'idle';
      hideGuestOverlay();
      window.location.reload();
    }).catch(()=>{
      showToast('Failed to leave', '#b71c1c');
    });
  };
}

function pollGuestState() {
  if (uiRole !== 'guest') {
    hideGuestOverlay();
    return;
  }
  fetch('/api/guest/state').then(r=>r.ok?r.json():Promise.reject()).then(data=>{
    // Possible states: idle, waiting, approved, denied, expired
    if (data.state === 'idle') {
      guestState = 'idle';
      showGuestOverlay('idle');
      enableNav(true);
    } else if (data.state === 'waiting') {
      guestState = 'waiting';
      showGuestOverlay('waiting', {countdown:data.countdown, status:data.status||''});
      enableNav(false);
    } else if (data.state === 'approved') {
      guestState = 'approved';
      showGuestOverlay('approved');
      enableNav(true);
    } else if (data.state === 'denied' || data.state === 'expired') {
      guestState = 'denied';
      showGuestOverlay('denied', {status:data.status||'Request denied or expired.'});
      enableNav(false);
    }
  }).catch(()=>{
    // fallback: treat as idle
    guestState = 'idle';
    showGuestOverlay('idle');
    enableNav(true);
  });
}

function enableNav(enable) {
  const nav = document.getElementById('bottom-nav');
  if (!nav) return;
  Array.from(nav.children).forEach(btn=>{
    btn.disabled = !enable;
    if (!enable) btn.classList.add('disabled');
    else btn.classList.remove('disabled');
  });
}

// Poll guest state every 2s
setInterval(pollGuestState, 2000);
window.addEventListener('DOMContentLoaded', pollGuestState);
function getRole() {
  const m = window.location.search.match(/[?&]role=(admin|user|guest)/);
  return m ? m[1] : 'guest';
}
let uiRole = getRole();

function updateRoleUI() {
  // Nav visibility
  navBtns.settings.style.display = (uiRole === 'guest') ? 'none' : '';
  navBtns.alarm.style.display = (uiRole !== 'guest') ? '' : 'none';
  // Devices tab: always shown
  // Views: restrict alarm controls for guest
  if (uiRole === 'guest') {
    document.getElementById('alarm-btns').style.display = 'none';
    document.getElementById('disarm-confirm').style.display = 'none';
  } else {
    document.getElementById('alarm-btns').style.display = '';
  }
  // Devices guest banner
  let guestBanner = document.getElementById('guest-banner');
  if (uiRole === 'guest') {
    if (!guestBanner) {
      guestBanner = document.createElement('div');
      guestBanner.id = 'guest-banner';
      guestBanner.textContent = 'Guest mode limited';
      guestBanner.style.background = '#b71c1c';
      guestBanner.style.color = '#fff';
      guestBanner.style.textAlign = 'center';
      guestBanner.style.padding = '0.7em 0';
      guestBanner.style.letterSpacing = '0.1em';
      guestBanner.style.marginBottom = '1em';
      guestBanner.style.borderRadius = '0.7em';
      guestBanner.style.opacity = '0.85';
      document.getElementById('view-devices').prepend(guestBanner);
    }
  } else if (guestBanner) {
    guestBanner.remove();
  }
  // Devices filtering for guest (if backend provides allowed list, else show all)
  // (Stub: no backend, so just show all for now)
  // Show current role in settings
  const el = document.getElementById('current-role');
  if (el) el.textContent = uiRole;
  // Hide Settings view for guest
  const settingsView = document.getElementById('view-settings');
  if (settingsView) settingsView.style.display = (uiRole === 'guest') ? 'none' : '';
// --- Settings View Logic ---
const settingsCards = {
  backend: document.getElementById('settings-backend-status'),
  ha: document.getElementById('settings-ha-status'),
  platform: document.getElementById('settings-platform-value'),
  profile: document.getElementById('settings-profile-value'),
  uptime: document.getElementById('settings-uptime-value'),
};
const advancedBtn = document.getElementById('toggle-advanced');
const advancedJson = document.getElementById('advanced-json');
let lastOverview = null;

function fetchSettingsData() {
  // Fetch all required endpoints in parallel
  Promise.all([
    fetch('/api/state/overview').then(r=>r.ok?r.json():Promise.reject()),
    fetch('/api/health').then(r=>r.ok?r.json():Promise.reject()).catch(()=>({})),
    fetch('/api/ui/home').then(r=>r.ok?r.json():Promise.reject()).catch(()=>({})),
  ]).then(([overview, health, home]) => {
    lastOverview = overview;
    // Backend status
    settingsCards.backend.textContent = (health && health.backend_online !== false) ? 'Online' : 'Offline';
    settingsCards.backend.style.color = (health && health.backend_online !== false) ? '#4caf50' : '#b71c1c';
    // HA status
    let haOnline = (overview && overview.ha_online !== undefined) ? overview.ha_online : (home && home.ha_online);
    settingsCards.ha.textContent = haOnline ? 'Online' : 'Offline';
    settingsCards.ha.style.color = haOnline ? '#4caf50' : '#b71c1c';
    // Show HA offline badge if needed
    let haStatusEl = document.getElementById('ha-status');
    if (haStatusEl) {
      let badge = document.getElementById('ha-offline-badge');
      if (!haOnline) {
        if (!badge) {
          badge = document.createElement('span');
          badge.id = 'ha-offline-badge';
          badge.className = 'ha-offline-badge';
          badge.textContent = 'Home Assistant not reachable';
          haStatusEl.parentNode.insertBefore(badge, haStatusEl.nextSibling);
        }
        badge.style.display = '';
      } else if (badge) {
        badge.style.display = 'none';
      }
    }
    // Platform
    let platform = (overview && overview.platform) || (health && health.platform) || 'Unknown';
    settingsCards.platform.textContent = platform;
    // Profile
    let profile = (overview && overview.profile) || (health && health.profile) || 'Unknown';
    settingsCards.profile.textContent = profile;
    // Uptime
    let uptime = (health && health.uptime) || (overview && overview.uptime) || '';
    settingsCards.uptime.textContent = uptime ? formatUptime(uptime) : '--';
    // Advanced JSON (admin only)
    if (uiRole === 'admin') {
      advancedBtn.style.display = '';
    } else {
      advancedBtn.style.display = 'none';
      advancedJson.style.display = 'none';
    }
  }).catch(()=>{
    Object.values(settingsCards).forEach(el=>{ el.textContent = '--'; el.style.color = '#fff'; });
    advancedBtn.style.display = 'none';
    advancedJson.style.display = 'none';
    // Show backend error banner if not already
    showBackendErrorBanner(true);
  });
}

function formatUptime(uptime) {
  // Accepts seconds or formatted string
  if (typeof uptime === 'number') {
    const d = Math.floor(uptime/86400), h = Math.floor((uptime%86400)/3600), m = Math.floor((uptime%3600)/60);
    return `${d}d ${h}h ${m}m`;
  }
  return uptime;
}

if (advancedBtn) {
  advancedBtn.onclick = function() {
    if (!advancedJson) return;
    if (advancedJson.style.display === 'none') {
      advancedJson.textContent = lastOverview ? JSON.stringify(lastOverview, null, 2) : '--';
      advancedJson.style.display = '';
      advancedBtn.textContent = 'Hide Advanced';
    } else {
      advancedJson.style.display = 'none';
      advancedBtn.textContent = 'Show Advanced';
    }
  };
}

// Poll settings data when Settings view is active
setInterval(()=>{
  if (window.location.hash === '#settings' && uiRole !== 'guest') fetchSettingsData();
}, 3000);
// Also fetch once on view switch
window.addEventListener('hashchange', ()=>{
  if (window.location.hash === '#settings' && uiRole !== 'guest') fetchSettingsData();
});
}
updateRoleUI();
window.addEventListener('hashchange', updateRoleUI);
// --- Devices View Logic ---
const devicesTabs = document.querySelectorAll('.devices-tab');
const devicesList = document.getElementById('devices-list');
const devicesSearch = document.getElementById('devices-search');
let devicesTab = 'lights';
let devicesData = {lights:[], covers:[], plugs:[]};
let devicesAvailable = {lights:true, covers:true, plugs:true};

devicesTabs.forEach(tab => {
  tab.onclick = () => {
    devicesTab = tab.dataset.tab;
    devicesTabs.forEach(t=>t.classList.remove('active'));
    tab.classList.add('active');
    renderDevices();
  };
});
devicesTabs[0].classList.add('active');

devicesSearch.oninput = renderDevices;

function fetchDevices() {
  if (window.location.hash !== '#devices') return;
  Promise.all([
    fetch('/api/devices/lights').then(r=>r.ok?r.json():Promise.reject()).catch(()=>{devicesAvailable.lights=false;return[];}),
    fetch('/api/devices/covers').then(r=>r.ok?r.json():Promise.reject()).catch(()=>{devicesAvailable.covers=false;return[];}),
    fetch('/api/devices/plugs').then(r=>r.ok?r.json():Promise.reject()).catch(()=>{devicesAvailable.plugs=false;return[];})
  ]).then(([lights,covers,plugs])=>{
    devicesData = {lights, covers, plugs};
    renderDevices();
  });
}
setInterval(fetchDevices, 3000);
window.addEventListener('hashchange', fetchDevices);

function renderDevices() {
  const list = devicesData[devicesTab]||[];
  const available = devicesAvailable[devicesTab];
  const filter = (devicesSearch.value||'').toLowerCase();
  devicesList.innerHTML = '';
  // If HA offline, disable all device actions
  let haOffline = false;
  let badge = document.getElementById('ha-offline-badge');
  if (badge && badge.style.display !== 'none') haOffline = true;
  if (!available || haOffline) {
    devicesList.innerHTML = '<div style="opacity:0.7;text-align:center;padding:2em;">' + (haOffline ? 'Home Assistant not reachable' : 'Not available') + '</div>';
    return;
  }
  const filtered = list.filter(d => (d.name||'').toLowerCase().includes(filter));
  if (filtered.length === 0) {
    devicesList.innerHTML = '<div style="opacity:0.7;text-align:center;padding:2em;">No devices found</div>';
    return;
  }
  filtered.forEach(dev => {
    devicesList.appendChild(deviceCard(dev, devicesTab, haOffline));
  });
}

function deviceCard(dev, type) {
  const card = document.createElement('div');
  card.className = 'device-card';
  const icon = document.createElement('div');
  icon.className = 'dev-icon';
  icon.textContent = type==='lights'?'💡':type==='plugs'?'🔌':type==='covers'?'🪟':'';
  card.appendChild(icon);
  const name = document.createElement('div');
  name.className = 'dev-name';
  name.textContent = dev.name || dev.id || 'Device';
  card.appendChild(name);
  const state = document.createElement('div');
  state.className = 'dev-state';
  state.textContent = dev.state || '';
  card.appendChild(state);
  const btn = document.createElement('button');
  btn.className = 'dev-action';
  // haOffline disables all actions
  if (arguments.length > 2 && arguments[2]) {
    btn.textContent = 'Unavailable';
    btn.disabled = true;
    btn.style.opacity = 0.5;
    btn.title = 'Home Assistant not reachable';
  } else if (type==='lights'||type==='plugs') {
    btn.textContent = (dev.state==='on'||dev.state===true)?'Turn Off':'Turn On';
    btn.onclick = () => {
      fetch('/api/devices/toggle', {
        method:'POST',
        headers:{'Content-Type':'application/json'},
        body:JSON.stringify({id:dev.id,type:type.slice(0,-1)})
      }).then(r=>r.ok?r.json():Promise.reject()).then(()=>{
        showToast('Toggled','');
        fetchDevices();
      }).catch(()=>showToast('Not implemented','#b71c1c'));
    };
  } else if (type==='covers') {
    btn.textContent = (dev.state==='open'||dev.state===true)?'Close':'Open';
    btn.onclick = () => {
      fetch('/api/devices/cover', {
        method:'POST',
        headers:{'Content-Type':'application/json'},
        body:JSON.stringify({id:dev.id,action:(dev.state==='open'||dev.state===true)?'close':'open'})
      }).then(r=>r.ok?r.json():Promise.reject()).then(()=>{
        showToast('Cover action','');
        fetchDevices();
      }).catch(()=>showToast('Not implemented','#b71c1c'));
    };
    if (dev.percent !== undefined && dev.percent !== null) {
      const pct = document.createElement('div');
      pct.className = 'dev-state';
      pct.textContent = 'Position: ' + dev.percent + '%';
      card.appendChild(pct);
    }
  }
  card.appendChild(btn);
  return card;
}
// --- Routing and View Management ---
const views = {
  home: document.getElementById('view-home'),
  alarm: document.getElementById('view-alarm'),
  devices: document.getElementById('view-devices'),
  settings: document.getElementById('view-settings'),
};
const navBtns = {
  home: document.getElementById('nav-home'),
  alarm: document.getElementById('nav-alarm'),
  devices: document.getElementById('nav-devices'),
  settings: document.getElementById('nav-settings'),
};
function showView(name) {
  // Restrict guest
  if (uiRole === 'guest' && name === 'alarm') name = 'home';
  if (uiRole === 'guest' && name === 'settings') name = 'home';
  Object.keys(views).forEach(k => {
    if (k === name) {
      views[k].classList.add('active');
    } else {
      views[k].classList.remove('active');
    }
  });
  window.location.hash = '#' + name;
}
function route() {
  const h = window.location.hash.replace('#','') || 'home';
  showView(h in views ? h : 'home');
}
window.addEventListener('hashchange', route);
Object.entries(navBtns).forEach(([k, btn]) => {
  btn.onclick = () => showView(k);
});
route();

// --- Toast Banner ---
let toastTimeout = null;
function showToast(msg, color='#263238') {
  let t = document.getElementById('toast-banner');
  if (!t) {
    t = document.createElement('div');
    t.id = 'toast-banner';
    t.style.position = 'fixed';
    t.style.bottom = '4.5em';
    t.style.left = '50%';
    t.style.transform = 'translateX(-50%)';
    t.style.background = color;
    t.style.color = '#fff';
    t.style.padding = '1em 2em';
    t.style.borderRadius = '1em';
    t.style.fontSize = '1.2em';
    t.style.zIndex = 2000;
    t.style.boxShadow = '0 2px 8px #0006';
    t.style.opacity = 0;
    t.className = '';
    document.body.appendChild(t);
  }
  t.textContent = msg;
  t.style.background = color;
  t.className = 'show';
  setTimeout(()=>{ t.style.opacity = 1; }, 10);
  clearTimeout(toastTimeout);
  toastTimeout = setTimeout(()=>{
    t.style.opacity = 0;
    setTimeout(()=>{ t.className = ''; }, 300);
  }, 2500);
}


// --- Alarm View Logic ---
const alarmStateEl = document.getElementById('alarm-state-indicator');
const alarmCountdownEl = document.getElementById('alarm-countdown');
const alarmGuestEl = document.getElementById('alarm-guest');
const alarmAIHintEl = document.getElementById('alarm-ai-hint');
const btnArmHome = document.getElementById('btn-arm-home');
const btnArmAway = document.getElementById('btn-arm-away');
const btnDisarm = document.getElementById('btn-disarm');
const disarmConfirm = document.getElementById('disarm-confirm');
const btnDisarmConfirm = document.getElementById('btn-disarm-confirm');
const btnDisarmCancel = document.getElementById('btn-disarm-cancel');
let disarmStep = false;
if (uiRole === 'guest') {
  btnArmHome.style.display = 'none';
  btnArmAway.style.display = 'none';
  btnDisarm.style.display = 'none';
}

function renderAlarmView(data) {
  alarmStateEl.textContent = data.alarm_state || 'Unknown';
  alarmCountdownEl.textContent = data.countdown ? 'Countdown: ' + data.countdown : '';
  alarmGuestEl.textContent = data.guest ? 'Guest: ' + data.guest : '';
  alarmAIHintEl.textContent = data.ai_hint || '';
  // Disable alarm actions if HA offline
  let badge = document.getElementById('ha-offline-badge');
  let haOffline = badge && badge.style.display !== 'none';
  [btnArmHome, btnArmAway, btnDisarm].forEach(btn => {
    if (haOffline) {
      btn.disabled = true;
      btn.title = 'Home Assistant not reachable';
      btn.style.opacity = 0.5;
    } else {
      btn.disabled = false;
      btn.title = '';
      btn.style.opacity = 1;
    }
  });
  // Pulse border if countdown
  if (data.countdown) {
    alarmCountdownEl.classList.add('pulse');
  } else {
    alarmCountdownEl.classList.remove('pulse');
  }
}

function pollAlarm() {
  fetch('/api/ui/alarm')
    .then(r => r.ok ? r.json() : Promise.reject())
    .then(data => {
      renderAlarmView(data);
    })
    .catch(() => {
      // fallback to /api/alarm/state + /api/ai/insight
      Promise.all([
        fetch('/api/alarm/state').then(r=>r.ok?r.json():{}).catch(()=>({})),
        fetch('/api/ai/insight').then(r=>r.ok?r.json():{}).catch(()=>({}))
      ]).then(([alarm, ai]) => {
        renderAlarmView({
          alarm_state: alarm.state,
          countdown: alarm.countdown,
          guest: alarm.guest,
          ai_hint: ai.summary || ''
        });
      });
    });
}

btnArmHome.onclick = function() {
  fetch('/api/alarm/arm', {
    method: 'POST',
    headers: {'Content-Type':'application/json'},
    body: JSON.stringify({mode:'home'})
  }).then(r=>r.ok?r.json():Promise.reject()).then(()=>{
    showToast('Armed Home', '#388e3c');
    pollAlarm();
  }).catch(()=>showToast('Failed to arm home', '#b71c1c'));
};
btnArmAway.onclick = function() {
  fetch('/api/alarm/arm', {
    method: 'POST',
    headers: {'Content-Type':'application/json'},
    body: JSON.stringify({mode:'away'})
  }).then(r=>r.ok?r.json():Promise.reject()).then(()=>{
    showToast('Armed Away', '#388e3c');
    pollAlarm();
  }).catch(()=>showToast('Failed to arm away', '#b71c1c'));
};
btnDisarm.onclick = function() {
  disarmStep = true;
  disarmConfirm.style.display = '';
};
btnDisarmCancel.onclick = function() {
  disarmStep = false;
  disarmConfirm.style.display = 'none';
};
btnDisarmConfirm.onclick = function() {
  fetch('/api/alarm/disarm', {
    method: 'POST',
    headers: {'Content-Type':'application/json'},
    body: JSON.stringify({method:'ui'})
  }).then(r=>r.ok?r.json():Promise.reject()).then(()=>{
    showToast('Disarmed', '#1976d2');
    pollAlarm();
    disarmStep = false;
    disarmConfirm.style.display = 'none';
  }).catch(()=>{
    showToast('Failed to disarm', '#b71c1c');
    disarmStep = false;
    disarmConfirm.style.display = 'none';
  });
};

// Poll alarm view only if visible
setInterval(()=>{
  if (window.location.hash === '#alarm') pollAlarm();
}, 2000);

const statusBar = document.getElementById('status-bar');
const datetimeEl = document.getElementById('datetime');
const haStatusEl = document.getElementById('ha-status');
const alarmStatusEl = document.getElementById('alarm-status');
const aiInsightEl = document.getElementById('ai-insight');
const aiSeverityEl = document.getElementById('ai-severity');
const tileTemp = document.getElementById('tile-temp');
const offlineBanner = document.getElementById('offline-banner');

function updateTime() {
  const now = new Date();
  datetimeEl.textContent = now.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) + ' ' + now.toLocaleDateString();
}
setInterval(updateTime, 1000);
updateTime();

function renderOverview(data) {
  if (!data) return;
  haStatusEl.textContent = data.ha_online ? 'HA: Online' : 'HA: Offline';
  haStatusEl.style.color = data.ha_online ? '#4caf50' : '#b71c1c';
  // Alarm state indicator color by severity
  let alarmState = (data.alarm_state || '').toLowerCase();
  alarmStatusEl.textContent = 'Alarm: ' + (data.alarm_state || 'Unknown');
  alarmStatusEl.classList.remove('alarm-critical','alarm-warning','alarm-ok');
  if (alarmState === 'triggered' || alarmState === 'alarm') {
    alarmStatusEl.classList.add('alarm-critical');
  } else if (alarmState === 'pending' || alarmState === 'arming') {
    alarmStatusEl.classList.add('alarm-warning');
  } else if (alarmState === 'disarmed' || alarmState === 'ok' || alarmState === 'ready') {
    alarmStatusEl.classList.add('alarm-ok');
  }
  // AI card style consistency
  aiInsightEl.textContent = data.ai_insight ? data.ai_insight.summary : 'No Insight';
  aiSeverityEl.textContent = data.ai_insight ? 'Severity: ' + data.ai_insight.severity : '';
  aiSeverityEl.className = '';
  if (data.ai_insight && data.ai_insight.severity) {
    let sev = data.ai_insight.severity.toLowerCase();
    if (sev === 'critical') aiSeverityEl.classList.add('alarm-critical');
    else if (sev === 'warning') aiSeverityEl.classList.add('alarm-warning');
    else aiSeverityEl.classList.add('alarm-ok');
  }
  if (data.temp !== undefined && data.temp !== null) {
    tileTemp.textContent = 'Temp\n' + data.temp + '°C';
  } else {
    tileTemp.textContent = 'Temp\n--°C';
  }
}

function poll() {
  try {
    if (window.location.hash && window.location.hash !== '#home') return;
    // Show skeleton while loading
    const aiCard = document.getElementById('ai-card');
    if (aiCard) aiCard.classList.add('skeleton');
    fetch('/api/ui/home').then(r => r.ok ? r.json() : Promise.reject()).then(data => {
      if (JSON.stringify(data) !== JSON.stringify(lastHomeData)) {
        lastHomeData = data;
        if (aiCard) aiCard.classList.remove('skeleton');
        renderOverview(data);
      }
    }).catch(() => {
      // fallback to overview
      fetch('/api/state/overview').then(r => r.ok ? r.json() : Promise.reject()).then(data => {
        if (JSON.stringify(data) !== JSON.stringify(lastOverviewData)) {
          lastOverviewData = data;
          if (aiCard) aiCard.classList.remove('skeleton');
          renderOverview(data);
        }
      }).catch(() => {
        if (aiCard) aiCard.classList.remove('skeleton');
        offlineBanner.style.display = 'block';
      });
    });
  } catch (e) {
    if (debug) console.error('Poll error', e);
  }
// --- AI Card Modal WHY ---
const aiCard = document.getElementById('ai-card');
if (aiCard) {
  aiCard.onclick = function() {
    showAIModal();
  };
}
function showAIModal() {
  let bg = document.createElement('div');
  bg.id = 'modal-bg';
  let panel = document.createElement('div');
  panel.id = 'modal-panel';
  let close = document.createElement('button');
  close.id = 'modal-close';
  close.innerHTML = '&times;';
  close.onclick = ()=>bg.remove();
  panel.appendChild(close);
  let content = document.createElement('div');
  content.textContent = 'Loading...';
  panel.appendChild(content);
  bg.appendChild(panel);
  document.body.appendChild(bg);
  fetch('/api/ui/ai/why')
    .then(r=>r.ok?r.text():Promise.reject())
    .then(txt=>{ content.textContent = txt; })
    .catch(()=>{
      fetch('/api/ai/insight/explain')
        .then(r=>r.ok?r.text():Promise.reject())
        .then(txt=>{ content.textContent = txt; })
        .catch(()=>{ content.textContent = 'No explanation available.'; });
    });
}
}
function startPolling() {
  debouncePoll(poll, isIdle ? idlePollIntervalMs : normalPollInterval);
}
window.addEventListener('hashchange', startPolling);
startPolling();
