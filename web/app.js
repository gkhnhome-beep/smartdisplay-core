// --- Topbar Status Icons (Event-driven, no polling) ---
function renderTopbarStatus(state) {
  // Compute status for each icon
  // Network
  let net = 'unknown';
  if (state?.network?.isOnline !== undefined) net = state.network.isOnline ? 'ok' : 'bad';
  else if (state?.net?.isOnline !== undefined) net = state.net.isOnline ? 'ok' : 'bad';
  else if (state?.connection?.online !== undefined) net = state.connection.online ? 'ok' : 'bad';
  else if (typeof window.navigator.onLine !== 'undefined') net = window.navigator.onLine ? 'ok' : 'bad';
  // Home Assistant
  let ha = 'unknown';
  if (state?.haState?.isConfigured === false) ha = 'bad';
  else if (state?.haState?.isConnected === false) ha = 'warn';
  else if (state?.haState?.isConnected === true) ha = 'ok';
  else if (state?.settings?.haState?.isConfigured === false) ha = 'bad';
  else if (state?.settings?.haState?.isConnected === false) ha = 'warn';
  else if (state?.settings?.haState?.isConnected === true) ha = 'ok';
  else if (state?.homeAssistant?.isConfigured === false) ha = 'bad';
  else if (state?.homeAssistant?.isConnected === false) ha = 'warn';
  else if (state?.homeAssistant?.isConnected === true) ha = 'ok';
  else if (window.haState && window.haState.isConfigured === false) ha = 'bad';
  else if (window.haState && window.haState.ha_connected === false) ha = 'warn';
  else if (window.haState && window.haState.ha_connected === true) ha = 'ok';
  // System Health
  let health = 'unknown';
  if (state?.systemHealth?.level) {
    if (state.systemHealth.level === 'ok') health = 'ok';
    else if (state.systemHealth.level === 'warn') health = 'warn';
    else if (state.systemHealth.level === 'bad') health = 'bad';
  } else if (state?.health?.status) {
    if (state.health.status === 'ok') health = 'ok';
    else if (state.health.status === 'warn') health = 'warn';
    else if (state.health.status === 'error' || state.health.status === 'bad') health = 'bad';
  }
  // Battery
  let battery = 'unknown';
  if (state?.battery?.level !== undefined) {
    battery = state.battery.level >= 20 ? 'ok' : 'bad';
  } else if (state?.battery?.charging) {
    battery = 'ok';
  } else if (state?.battery?.level !== undefined) {
    battery = state.battery.level < 20 ? 'bad' : 'ok';
  }

  // --- STATUS ICONS ---
  let html = `
    <div class="topbar-status-icon" title="Network: ${net}">
      <svg class="icon" viewBox="0 0 24 24">
        <path d="M12 3C17.5 3 22 7.5 22 13H20C20 8.58 16.42 5 12 5S4 8.58 4 13H2C2 7.5 6.5 3 12 3ZM12 1C6.48 1 2 5.48 2 11H4C4 6.58 7.58 3 12 3C16.42 3 20 6.58 20 11H22C22 5.48 17.52 1 12 1ZM12 8C10.9 8 10 8.9 10 10V14C10 15.1 10.9 16 12 16C13.1 16 14 15.1 14 14V10C14 8.9 13.1 8 12 8ZM12 6C13.66 6 15 7.34 15 9V15C15 16.66 13.66 18 12 18C10.34 18 9 16.66 9 15V9C9 7.34 10.34 6 12 6ZM12 4C10.34 4 9 5.34 9 7V17C9 18.66 10.34 20 12 20C13.66 20 15 18.66 15 17V7C15 5.34 13.66 4 12 4Z"/>
      </svg>
    </div>
    <div class="topbar-status-icon" title="Home Assistant: ${ha}">
      <svg class="icon" viewBox="0 0 24 24">
        <path d="M12 3C17.5 3 22 7.5 22 13H20C20 8.58 16.42 5 12 5S4 8.58 4 13H2C2 7.5 6.5 3 12 3ZM12 1C6.48 1 2 5.48 2 11H4C4 6.58 7.58 3 12 3C16.42 3 20 6.58 20 11H22C22 5.48 17.52 1 12 1ZM12 8C10.9 8 10 8.9 10 10V14C10 15.1 10.9 16 12 16C13.1 16 14 15.1 14 14V10C14 8.9 13.1 8 12 8ZM12 6C13.66 6 15 7.34 15 9V15C15 16.66 13.66 18 12 18C10.34 18 9 16.66 9 15V9C9 7.34 10.34 6 12 6ZM12 4C10.34 4 9 5.34 9 7V17C9 18.66 10.34 20 12 20C13.66 20 15 18.66 15 17V7C15 5.34 13.66 4 12 4Z"/>
      </svg>
    </div>
    <div class="topbar-status-icon" title="System Health: ${health}">
      <svg class="icon" viewBox="0 0 24 24">
        <path d="M12 3C17.5 3 22 7.5 22 13H20C20 8.58 16.42 5 12 5S4 8.58 4 13H2C2 7.5 6.5 3 12 3ZM12 1C6.48 1 2 5.48 2 11H4C4 6.58 7.58 3 12 3C16.42 3 20 6.58 20 11H22C22 5.48 17.52 1 12 1ZM12 8C10.9 8 10 8.9 10 10V14C10 15.1 10.9 16 12 16C13.1 16 14 15.1 14 14V10C14 8.9 13.1 8 12 8ZM12 6C13.66 6 15 7.34 15 9V15C15 16.66 13.66 18 12 18C10.34 18 9 16.66 9 15V9C9 7.34 10.34 6 12 6ZM12 4C10.34 4 9 5.34 9 7V17C9 18.66 10.34 20 12 20C13.66 20 15 18.66 15 17V7C15 5.34 13.66 4 12 4Z"/>
      </svg>
    </div>
    <div class="topbar-status-icon" title="Battery: ${battery}">
      <svg class="icon" viewBox="0 0 24 24">
        <path d="M17 4H7C5.9 4 5 4.9 5 6V18C5 19.1 5.9 20 7 20H17C18.1 20 19 19.1 19 18V6C19 4.9 18.1 4 17 4ZM7 2C4.79 2 3 3.79 3 6V18C3 20.21 4.79 22 7 22H17C19.21 22 21 20.21 21 18V6C21 3.79 19.21 2 17 2H7ZM17 8H7V10H17V8ZM17 12H7V14H17V12ZM17 16H7V18H17V16Z"/>
      </svg>
    </div>
  `;
  // Update topbar HTML
  const topbar = document.getElementById('topbar');
  if (topbar) {
    topbar.querySelector('#topbar-right').innerHTML = html;
  }
}

// --- THERMOSTAT CARD (ADMIN ONLY) ---
function renderThermostatsPage() {
  const main = document.getElementById('main-content');
  main.innerHTML = `<div class="glass" id="thermostat-root"><h3>Termostatlar</h3><div id="thermostat-status">Yükleniyor...</div></div>`;

  // Fetch real thermostat states from backend
  // Replace with SmartDisplay.api call if available
  fetch('/api/thermostat/state')
    .then(r => r.json())
    .then(states => {
      // Defensive: Only render if valid states
      if (!Array.isArray(states)) {
        document.getElementById('thermostat-status').innerHTML = `<span style='color:#e74c3c'>Termostat durumu alınamadı.</span>`;
        return;
      }
      let html = '';
      for (const state of states) {
        const cfg = thermostatConfig ? (thermostatConfig[state.id] || {}) : {};
        const alias = cfg.alias || state.name || state.id;
        let current = state.current_temperature !== undefined ? state.current_temperature : '?';
        let target = state.target_temperature !== undefined ? state.target_temperature : '?';
        let unavailable = state.unavailable;
        let mode = state.hvac_mode || 'off';
        let sysState = state.hvac_action || 'idle';
        let lastUpdate = state.last_updated ? `<div class=\"thermo-last\">${state.last_updated}</div>` : '';
        let cardClass = 'thermo-card glass';
        if (unavailable) cardClass += ' thermo-card-unavail';
        let modeBadge = '';
        if (mode === 'heat') modeBadge = '<span class="thermo-badge heat">Isıtma</span>';
        else if (mode === 'cool') modeBadge = '<span class="thermo-badge cool">Soğutma</span>';
        else if (mode === 'auto') modeBadge = '<span class="thermo-badge auto">Otomatik</span>';
        else modeBadge = '<span class="thermo-badge off">Kapalı</span>';
        let sysBadge = '';
        if (sysState === 'heating') sysBadge = '<span class="thermo-badge heat">Isıtıyor</span>';
        else if (sysState === 'cooling') sysBadge = '<span class="thermo-badge cool">Soğutuyor</span>';
        else sysBadge = '<span class="thermo-badge idle">Beklemede</span>';
        html += `<div class="${cardClass}">
          <div class="thermo-alias">${alias}</div>
          <div class="thermo-room">${cfg.room||''}</div>
          <div class="thermo-current" style="font-size:2.2em;font-weight:700;line-height:1.1;">${current}<span style=\"font-size:0.6em;opacity:0.7;\">°C</span></div>
          ${target!==''?`<div class=\"thermo-target\">Hedef: ${target}°C</div>`:''}
          <div class="thermo-badges">${modeBadge} ${sysBadge}</div>
          ${lastUpdate}
          ${unavailable?'<div class=\"thermo-badge thermo-unavail\">Unavailable</div>':''}
        </div>`;
      }
      document.getElementById('thermostat-status').innerHTML = html;
    })
    .catch(() => {
      document.getElementById('thermostat-status').innerHTML = `<span style='color:#e74c3c'>Termostat durumu alınamadı.</span>`;
    });
}

/* ---------- ROUTER ---------- */
function route() {
  const hash = location.hash || '#/intro';

  if (hash === '#/intro') {
    renderIntro();
    return;
  }

  if (hash === '#/login') {
    renderLogin();
    return;
  }

  // Sadece ana sayfa ve alt sayfalarda layout'u render et
  let page = hash.replace('#/', '');
  renderLayout(page, currentUser);

  if (page === 'home') renderHome();
  else if (page === 'settings') renderSettingsContent('ha');
  else if (page === 'users') renderSettingsContent('users');
  else if (page === 'logs') renderSettingsContent('logs');
  else if (page === 'lighting') renderLightingPage();
  else if (page === 'thermostats') renderThermostatsPage();
  else if (page === 'alarm') renderAlarmView();


/* ---------- ALARM VIEW ---------- */
function renderAlarmView() {
  const main = document.getElementById('main-content');
  main.innerHTML = `<div class="glass" id="alarm-root"><h3>Alarm</h3><div id="alarm-status">Yükleniyor...</div></div>`;

  // Strictly fetch real alarm state from backend (never infer/fake)
  // Replace with SmartDisplay.api call if available
  fetch('/api/alarm/state')
    .then(r => r.json())
    .then(state => {
      // Defensive: Only render if valid state
      if (!state || typeof state.status !== 'string') {
        document.getElementById('alarm-status').innerHTML = `<span style='color:#e74c3c'>Alarm durumu alınamadı.</span>`;
        return;
      }
      let statusText = '';
      let color = '#fff';
      switch (state.status) {
        case 'disarmed':
          statusText = 'Devre Dışı'; color = '#2ecc40'; break;
        case 'armed_home':
          statusText = 'Evde Kurulu'; color = '#f1c40f'; break;
        case 'armed_away':
          statusText = 'Dışarıda Kurulu'; color = '#e67e22'; break;
        case 'triggered':
          statusText = 'Alarm Çaldı!'; color = '#e74c3c'; break;
        default:
          statusText = 'Bilinmeyen'; color = '#b0c4de'; break;
      }
      document.getElementById('alarm-status').innerHTML = `<div style='font-size:2em;font-weight:700;color:${color};margin:18px 0;'>${statusText}</div>`;
    })
    .catch(() => {
      document.getElementById('alarm-status').innerHTML = `<span style='color:#e74c3c'>Alarm durumu alınamadı.</span>`;
    });
}
/* ---------- LOGIN ---------- */
function renderLogin() {
  // Sadece #app içeriğini değiştir, başka DOM elemanlarını kaldırma
  const app = document.getElementById('app');
  if (app) {
    app.innerHTML = `
      <div style="height:100vh;display:flex;align-items:center;justify-content:center;background:#0a1a2f">
        <div class="glass" style="padding:30px;width:300px;text-align:center">
          <h2>PIN Giriş</h2>
          <div id="pinDots" style="margin:20px;font-size:24px">••••</div>
          <div id="keys"></div>
          <div id="loginError" style="color:#e74c3c;margin-top:10px"></div>
        </div>
      </div>
    `;
  }

  let pin = '';
  const dots = document.getElementById('pinDots');
  const keys = document.getElementById('keys');

  for (let i = 1; i <= 9; i++) {
    keys.innerHTML += `<button data-k="${i}">${i}</button>`;
  }
  keys.innerHTML += `<button data-k="0">0</button>`;

  async function tryLogin(pin) {
    document.getElementById('loginError').textContent = '';
    try {
      const resp = await fetch('/api/ui/home/state', {
        method: 'GET',
        headers: { 'X-SmartDisplay-PIN': pin }
      });
      if (resp.ok) {
        location.hash = '#/home';
      } else {
        const data = await resp.json().catch(() => ({}));
        document.getElementById('loginError').textContent = data.error || 'Hatalı PIN';
        pin = '';
        dots.textContent = '••••';
      }
    } catch (e) {
      document.getElementById('loginError').textContent = 'Sunucuya erişilemiyor';
      pin = '';
      dots.textContent = '••••';
    }
  }

  keys.addEventListener('click', e => {
    if (!e.target.dataset.k) return;
    pin += e.target.dataset.k;
    dots.textContent = '•'.repeat(pin.length).padEnd(4, '•');

    if (pin.length === 4) {
      tryLogin(pin);
    }
  });
}

/* ---------- INTRO ---------- */
function renderIntro() {
  // Sadece #app içine intro HTML'ini yaz
  const app = document.getElementById('app');
  if (app) {
    app.innerHTML = `
      <div class="intro-container center-all">
        <div class="intro-title premium-glow">Smart Display</div>
        <div class="intro-logo">
          <div class="intro-logo-circle premium-pulse">
            <img src="ha_logo.png" alt="Home Assistant Logo" class="ha-logo-img premium-glow" />
          </div>
        </div>
        <div class="intro-bar premium-bar">
          <div class="intro-bar-fill premium-bar-fill" id="introBarFill"></div>
        </div>
        <div class="intro-status premium-glow" id="introStatus">System Started...</div>
        <div class="intro-status-sub premium-glow" id="introStatusSub">Donanım kontrol ediliyor...</div>
        <div class="intro-status-ok premium-glow" id="introStatusOk" style="display:none">System Check OK</div>
      </div>
    `;
  }
  // Yıldız animasyonu için sadece #star-bg kullan
  setupPremiumStars();
  window.addEventListener('resize', setupPremiumStars);
  animateStars();
  startIntroSequence();
}

function setupPremiumStars() {
  const canvas = document.getElementById('star-bg');
  if (!canvas) return;
  canvas.width = window.innerWidth;
  canvas.height = window.innerHeight;
  premiumStarCtx = canvas.getContext('2d');
  // Generate stars
  premiumStars = [];
  const STAR_COUNT = Math.floor(window.innerWidth * window.innerHeight / 1800);
  for (let i = 0; i < STAR_COUNT; i++) {
    premiumStars.push({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      r: Math.random() * 0.8 + 0.5,
      baseAlpha: Math.random() * 0.5 + 0.3,
      twinkle: Math.random() * Math.PI * 2,
      speed: Math.random() * 0.08 + 0.02,
      color: Math.random() < 0.7 ? "#eaf6ff" : "#7fd6ff"
    });
  }
  animatePremiumStars();
}

// Add a small runtime helper: mark body to show stars when canvas is present
(function(){
  function enableIfCanvas(){
    const canvas = document.getElementById('star-bg') || document.getElementById('intro-stars') || document.querySelector('canvas');
    if(canvas){
      document.body.classList.add('show-stars');
      return true;
    }
    return false;
  }
  if(!enableIfCanvas()){
    // wait briefly for canvas to be injected by other modules
    let tries = 0;
    const iv = setInterval(()=>{ tries++; if(enableIfCanvas()||tries>40) clearInterval(iv); }, 100);
  }
})();

function spawnShootingStar() {
  const canvas = document.getElementById('star-bg');
  shootingStars.push({
    x: Math.random() * canvas.width * 0.7,
    y: Math.random() * canvas.height * 0.4,
    len: 120 + Math.random() * 80,
    speed: 7 + Math.random() * 4,
    angle: Math.PI / 4 + (Math.random() - 0.5) * 0.2,
    alpha: 1.0
  });
}

function animatePremiumStars() {
  const canvas = document.getElementById('star-bg');
  premiumStarCtx.clearRect(0, 0, canvas.width, canvas.height);
  const t = performance.now() * 0.00025;
  for (let s of premiumStars) {
    let alpha = s.baseAlpha + 0.25 * Math.sin(t * 2 + s.twinkle);
    s.x += Math.sin(s.twinkle + t) * s.speed * 0.2;
    s.y += Math.cos(s.twinkle + t) * s.speed * 0.2;
    if (s.x < 0) s.x += canvas.width;
    if (s.x > canvas.width) s.x -= canvas.width;
    if (s.y < 0) s.y += canvas.height;
    if (s.y > canvas.height) s.y -= canvas.height;
    premiumStarCtx.save();
    premiumStarCtx.globalAlpha = Math.max(0, Math.min(1, alpha));
    premiumStarCtx.beginPath();
    premiumStarCtx.arc(s.x, s.y, s.r, 0, 2 * Math.PI);
    premiumStarCtx.shadowColor = s.color;
    premiumStarCtx.shadowBlur = 8;
    premiumStarCtx.fillStyle = s.color;
    premiumStarCtx.fill();
    premiumStarCtx.restore();
  }
  // Shooting stars
  for (let i = shootingStars.length - 1; i >= 0; i--) {
    let s = shootingStars[i];
    premiumStarCtx.save();
    premiumStarCtx.globalAlpha = s.alpha;
    let dx = Math.cos(s.angle) * s.len;
    let dy = Math.sin(s.angle) * s.len;
    let grad = premiumStarCtx.createLinearGradient(s.x, s.y, s.x + dx, s.y + dy);
    grad.addColorStop(0, '#fff');
    grad.addColorStop(1, 'rgba(65,189,245,0)');
    premiumStarCtx.strokeStyle = grad;
    premiumStarCtx.lineWidth = 2.5;
    premiumStarCtx.beginPath();
    premiumStarCtx.moveTo(s.x, s.y);
    premiumStarCtx.lineTo(s.x + dx, s.y + dy);
    premiumStarCtx.stroke();
    premiumStarCtx.restore();
    s.x += Math.cos(s.angle) * s.speed;
    s.y += Math.sin(s.angle) * s.speed;
    s.alpha -= 0.012;
    if (s.alpha <= 0) shootingStars.splice(i, 1);
  }
  if (Math.random() < 0.012) spawnShootingStar();
  requestAnimationFrame(animatePremiumStars);
}

function animateStars() {
  const canvas = document.getElementById("star-bg");
  const ctx = canvas.getContext("2d");
  let w = window.innerWidth, h = window.innerHeight;
  canvas.width = w; canvas.height = h;

  // Generate stars
  const stars = [];
  for (let i = 0; i < 120; i++) {
    stars.push({
      x: Math.random() * w,
      y: Math.random() * h,
      r: Math.random() * 1.2 + 0.5,
      a: Math.random() * 2 * Math.PI,
      speed: Math.random() * 0.08 + 0.02
    });
  }

  function draw() {
    ctx.clearRect(0, 0, w, h);
    for (const star of stars) {
      // shimmer effect
      const glow = Math.sin(Date.now() / 1000 + star.a) * 0.3 + 0.7;
      ctx.save();
      ctx.beginPath();
      ctx.arc(star.x, star.y, star.r, 0, 2 * Math.PI);
      ctx.fillStyle = `rgba(255,255,255,${glow * 0.7})`;
      ctx.shadowColor = "#41bdf5";
      ctx.shadowBlur = 8 * glow;
      ctx.fill();
      ctx.restore();
      // slow movement
      star.x += Math.cos(star.a) * star.speed;
      star.y += Math.sin(star.a) * star.speed;
      // wrap
      if (star.x < 0) star.x += w;
      if (star.x > w) star.x -= w;
      if (star.y < 0) star.y += h;
      if (star.y > h) star.y -= h;
    }
    requestAnimationFrame(draw);
  }
  draw();
  window.addEventListener("resize", () => {
    w = window.innerWidth; h = window.innerHeight;
    canvas.width = w; canvas.height = h;
  });
}

function startIntroSequence() {
  const title = document.querySelector(".intro-title");
  const logo = document.querySelector(".intro-logo");
  const bar = document.querySelector(".intro-bar");
  const barFill = document.getElementById("introBarFill");
  const status = document.getElementById("introStatus");
  const statusOk = document.getElementById("introStatusOk");

  // 1. Fade-in title
  setTimeout(() => {
    if (title) title.style.opacity = 1;
  }, 200);

  // 2. Fade-out title, fade-in logo/bar
  setTimeout(() => {
    if (title) title.style.opacity = 0;
    if (logo) logo.style.opacity = 1;
    if (bar) bar.style.opacity = 1;
    updateLoadingProgress();
  }, 1800);

  // 3. Status messages and bar animation
  let msgStep = 0;
  function showNextStatus() {
    if (msgStep < INTRO_MESSAGES.length - 1) {
      if (status) {
        status.textContent = INTRO_MESSAGES[msgStep];
        status.style.color = INTRO_COLORS[msgStep];
        status.style.opacity = 1;
      }
      setTimeout(() => {
        if (status) status.style.opacity = 0;
        msgStep++;
        setTimeout(showNextStatus, 400);
      }, 1200);
    } else {
      // Final message
      if (statusOk) statusOk.style.opacity = 1;
      if (status) status.style.opacity = 0;
      if (barFill) barFill.style.width = "100%";
      setTimeout(goToLogin, 800);
    }
  }
  setTimeout(showNextStatus, 2200);
}

function updateLoadingProgress() {
  const barFill = document.getElementById("introBarFill");
  let progress = 0;
  function animate() {
    if (progress < 100) {
      progress += 1.5;
      if (barFill) barFill.style.width = progress + "%";
      requestAnimationFrame(animate);
    }
  }
  if (barFill) animate();
}

function goToLogin() {
  const root = document.getElementById("intro-root");
  const stars = document.getElementById("intro-stars");
  if (root) {
    root.style.transition = "opacity 1s";
    root.style.opacity = 0;
  }
  if (stars) {
    stars.style.transition = "opacity 1s";
    stars.style.opacity = 0;
  }
  setTimeout(() => {
    if (root) root.style.display = "none";
    if (stars) stars.style.display = "none";
    if (typeof renderLogin === 'function') renderLogin();
  }, 1000);
}

// Yıldız efekti fonksiyonu
function drawStars(canvas) {
  const ctx = canvas.getContext('2d');
  const w = window.innerWidth;
  const h = window.innerHeight;
  canvas.width = w;
  canvas.height = h;
  ctx.clearRect(0, 0, w, h);
  for (let i = 0; i < 120; i++) {
    const x = Math.random() * w;
    const y = Math.random() * h;
    const r = Math.random() * 1.2 + 0.5;
    ctx.beginPath();
    ctx.arc(x, y, r, 0, 2 * Math.PI);
    ctx.fillStyle = 'rgba(255,255,255,' + (Math.random() * 0.7 + 0.3) + ')';
    ctx.shadowColor = '#41bdf5';
    ctx.shadowBlur = 8;
    ctx.fill();
    ctx.shadowBlur = 0;
  }
}

// Safety starter: ensure star canvas animation runs if present (helps when CSS/js loading order changed)
window.addEventListener('DOMContentLoaded', () => {
  try {
    const canvas = document.getElementById('star-bg');
    if (!canvas) return;
    // Prefer premium star setup if available
    if (typeof setupPremiumStars === 'function') {
      try { setupPremiumStars(); } catch (e) { console.warn('setupPremiumStars failed', e); }
      return;
    }
    // Fallback to legacy setupCanvas
    if (typeof setupCanvas === 'function') {
      try { setupCanvas(); } catch (e) { console.warn('setupCanvas failed', e); }
      return;
    }
    // As a last resort, call animateStars if defined
    if (typeof animateStars === 'function') {
      try { animateStars(); } catch (e) { console.warn('animateStars failed', e); }
    }
  } catch (e) { console.error('Star safety starter error', e); }
});

/* ---------- LAYOUT ---------- */
function renderLayout(page, user) {
  // Remove any previous back button event
  if (window._settingsBackBtnHandler) {
    document.removeEventListener('click', window._settingsBackBtnHandler);
    window._settingsBackBtnHandler = null;
  }
  // Main layout
  document.getElementById('app').innerHTML = `
    <div class="topbar glass" id="topbar">
      <span id="topbar-title">SmartDisplay</span>
      <span id="topbar-right"></span>
    </div>
    <div class="layout">
      <div class="sidebar glass" id="sidebar">
        <div data-link="home">Home</div>
        <div data-link="lighting">Lighting</div>
        <div data-link="thermostats">Thermostats</div>
        <div data-link="settings">Settings</div>
        <div data-link="users">Users</div>
        <div data-link="logs">Logs</div>
        <div data-link="logout">Logout</div>
      </div>
      <div id="main-content" class="content"></div>
    </div>
  `;
  // Settings mode UI contract
  if (page === 'settings') {
    // Hide global sidebar, show settings sidebar
    const sidebar = document.getElementById('sidebar');
    if (sidebar) sidebar.style.display = 'none';
    // Mount settings view in main-content
    const main = document.getElementById('main-content');
    if (main) mountSettingsView(main);
    // Show settings sidebar
    const settingsSidebar = document.getElementById('settingsSidebar');
    if (settingsSidebar) settingsSidebar.style.display = '';
    // Topbar right: Back button
    const topbarRight = document.getElementById('topbar-right');
    if (topbarRight) {
      topbarRight.innerHTML = `<button id="settings-back-btn" style="padding:8px 18px;border-radius:8px;background:#232c36;color:#b0c4d4;font-size:1.08em;border:none;cursor:pointer;box-shadow:0 2px 8px #0002;">⬅️ Back</button>`;
      // Back button event
      window._settingsBackBtnHandler = function(e) {
        if (e.target && e.target.id === 'settings-back-btn') {
          location.hash = '#/home';
        }
      };
      document.addEventListener('click', window._settingsBackBtnHandler);
    }
  } else {
    // Restore global sidebar
    const sidebar = document.getElementById('sidebar');
    if (sidebar) sidebar.style.display = '';
    // Remove settings sidebar if present
    const settingsSidebar = document.getElementById('settingsSidebar');
    if (settingsSidebar) settingsSidebar.style.display = 'none';
    // Topbar right: username
    const topbarRight = document.getElementById('topbar-right');
    if (topbarRight) topbarRight.textContent = user.name;
  }
}
// --- LIGHTING PAGE (MAIN UI) ---
function renderLightingPage() {
  if (currentUser.role === 'guest') {
    document.getElementById('main-content').innerHTML = `<div class="glass"><h3>Lighting</h3><p>Bu sayfa misafirler için gizli.</p></div>`;
    return;
  }
  let lightingConfig = {};
  try { lightingConfig = JSON.parse(localStorage.getItem('lightingConfig')||'{}')||{}; } catch(e){}
  fetch('/api/devices?type=light').then(r=>r.json()).then(lights => {
    if (!Array.isArray(lights) || lights.length === 0) {
      document.getElementById('main-content').innerHTML = `<div class="glass"><h3>Lighting</h3><div style="color:#b0c4de">Hiç ışık bulunamadı. Home Assistant bağlantınızı kontrol edin.</div></div>`;
      return;
    }
    // Filter enabled lights
    const enabled = lights.filter(l => lightingConfig[l.id] && lightingConfig[l.id].enabled);
    if (enabled.length === 0) {
      document.getElementById('main-content').innerHTML = `<div class="glass"><h3>Lighting</h3><div style="color:#b0c4de">Hiçbir ışık etkin değil.<br>Ayarlardan ışık ekleyin.</div></div>`;
      return;
    }
    // Group by room
    const byRoom = {};
    for (const l of enabled) {
      const room = (lightingConfig[l.id] && lightingConfig[l.id].room) || 'Other';
      if (!byRoom[room]) byRoom[room] = [];
      byRoom[room].push(l);
    }
    let html = '<div class="lighting-grid-root">';
    for (const room of Object.keys(byRoom)) {
      html += `<div class="lighting-room-group"><div class="lighting-room-title">${room}</div><div class="lighting-grid">`;
      for (const light of byRoom[room]) {
        const cfg = lightingConfig[light.id] || {};
        const cardType = cfg.cardType || 'auto';
        const alias = cfg.alias || light.name || light.id;
        // Card rendering by type
        html += renderLightingCard(light, cardType, alias);
      }
      html += '</div></div>';
    }
    html += '</div>';
    document.getElementById('main-content').innerHTML = `<div class="glass"><h3>Lighting</h3>${html}</div>`;
  }).catch(()=>{
    document.getElementById('main-content').innerHTML = `<div class="glass"><h3>Lighting</h3><div style="color:#b0c4de">Home Assistant'a ulaşılamıyor. Lütfen bağlantınızı kontrol edin.</div></div>`;
  });
}
}

// Card rendering helper
function renderLightingCard(light, cardType, alias) {
  // For now, all cards are read-only, state is shown but not interactive
  // Simulate state: light.state = {on: true/false, brightness: 0-255, rgb: [r,g,b]}
  let state = light.state || {on: false};
  let unavailable = state.unavailable;
  let isOn = !!state.on;
  let brightness = typeof state.brightness === 'number' ? Math.round(state.brightness/2.55) : null;
  let rgb = Array.isArray(state.rgb) ? state.rgb : null;
  let cardClass = 'lighting-card glass';
  if (!isOn) cardClass += ' lighting-card-off';
  if (unavailable) cardClass += ' lighting-card-unavail';
  let icon = '<span class="lighting-icon">💡</span>';
  let stateText = unavailable ? 'Unavailable' : (isOn ? 'On' : 'Off');
  let extra = '';
  if (cardType === 'dimmer' && brightness !== null) {
    extra = `<div class="lighting-dimmer-bar"><div class="lighting-dimmer-fill" style="height:${brightness}%;"></div></div><div class="lighting-dimmer-label">${brightness}%</div>`;
  }
  if (cardType === 'rgb' && rgb) {
    extra = `<div class="lighting-rgb-circle" style="background:rgb(${rgb[0]},${rgb[1]},${rgb[2]})"></div>`;
    if (brightness !== null) extra += `<div class="lighting-dimmer-label">${brightness}%</div>`;
  }
  return `<div class="${cardClass}">
    ${icon}
    <div class="lighting-alias">${alias}</div>
    <div class="lighting-state">${stateText}</div>
    ${extra}
  </div>`;
}
