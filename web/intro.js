// --- LOGIN LOGIC ---
window.addEventListener('DOMContentLoaded', () => {
  const loginBtn = document.getElementById('login-btn');
  if (loginBtn) {
    loginBtn.addEventListener('click', handleLogin);
    document.getElementById('admin-password').addEventListener('keydown', function(e) {
      if (e.key === 'Enter') handleLogin();
    });
  }
});

function handleLogin() {
  const pw = document.getElementById('admin-password').value;
  const errorDiv = document.getElementById('login-error');
  if (pw === '1234') {
    window.location.href = 'home.html'; // veya window.location.hash = '#/home';
  } else {
    errorDiv.textContent = 'Hatalı şifre!';
    errorDiv.style.display = 'block';
  }
}

// --- SHOOTING STARS ---
let shootingStars = [];
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

function animateStars() {
  const canvas = document.getElementById('star-bg');
  starCtx.clearRect(0, 0, canvas.width, canvas.height);
  const t = performance.now() * 0.00025;
  for (let s of stars) {
    // Twinkle
    let alpha = s.baseAlpha + 0.25 * Math.sin(t * 2 + s.twinkle);
    // Yıldızları sağa ve aşağı kaydır (parallax)
    s.x += 0.08 + Math.sin(t + s.twinkle) * s.speed * 0.2;
    s.y += 0.04 + Math.cos(t + s.twinkle) * s.speed * 0.2;
    // Wrap
    if (s.x < 0) s.x += canvas.width;
    if (s.x > canvas.width) s.x -= canvas.width;
    if (s.y < 0) s.y += canvas.height;
    if (s.y > canvas.height) s.y -= canvas.height;
    // Draw
    starCtx.save();
    starCtx.globalAlpha = Math.max(0, Math.min(1, alpha));
    starCtx.beginPath();
    starCtx.arc(s.x, s.y, s.r, 0, 2 * Math.PI);
    starCtx.shadowColor = s.color;
    starCtx.shadowBlur = 8;
    starCtx.fillStyle = s.color;
    starCtx.fill();
    starCtx.restore();
  }
  // Shooting stars
  for (let i = shootingStars.length - 1; i >= 0; i--) {
    let s = shootingStars[i];
    starCtx.save();
    starCtx.globalAlpha = s.alpha;
    let dx = Math.cos(s.angle) * s.len;
    let dy = Math.sin(s.angle) * s.len;
    let grad = starCtx.createLinearGradient(s.x, s.y, s.x + dx, s.y + dy);
    grad.addColorStop(0, '#fff');
    grad.addColorStop(1, 'rgba(65,189,245,0)');
    starCtx.strokeStyle = grad;
    starCtx.lineWidth = 2.5;
    starCtx.beginPath();
    starCtx.moveTo(s.x, s.y);
    starCtx.lineTo(s.x + dx, s.y + dy);
    starCtx.stroke();
    starCtx.restore();
    s.x += Math.cos(s.angle) * s.speed;
    s.y += Math.sin(s.angle) * s.speed;
    s.alpha -= 0.012;
    if (s.alpha <= 0) shootingStars.splice(i, 1);
  }
  if (Math.random() < 0.012) spawnShootingStar();
  requestAnimationFrame(animateStars);
}
// ---- CONFIG ----
const STATUS_MESSAGES = [
  { text: "System Started...", class: "" },
  { text: "Donanım kontrol ediliyor...", class: "" },
  { text: "Network initializing...", class: "" },
  { text: "Connecting to Home Assistant...", class: "" },
  { text: "System Check OK", class: "ok" }
];
const LOADING_DURATION = 5200; // ms, total loading bar time
const STATUS_INTERVALS = [0, 1100, 2200, 3400, 4400]; // ms, when to show each message
const FADE_TIME = 700; // ms, fade in/out for status
const TITLE_FADEIN = 900;
const TITLE_FADEOUT = 900;
const LOGO_FADEIN = 1100;
const LOADINGBAR_FADEIN = 700;
const LOGIN_TRANSITION_DELAY = 500; // ms after 100%

// ---- INIT ----

window.addEventListener('DOMContentLoaded', initIntro);

function initIntro() {
  setupCanvas();
  animateStars();

  // Show title immediately (SPA'da intro yoksa hata olmasın)
  const title = document.getElementById('intro-title');
  if (title) title.style.opacity = 1;

  // Hide logo/bar/status initially (SPA'da intro yoksa hata olmasın)
  const logo = document.getElementById('intro-logo-container');
  if (logo) logo.style.opacity = 0;
  const barBg = document.getElementById('intro-loading-bar-bg');
  if (barBg) barBg.style.opacity = 0;
  const status = document.getElementById('intro-status');
  if (status) status.style.opacity = 0;

  // Show logo and bar after fade-in
  setTimeout(() => {
    showLogoAndBar();
  }, TITLE_FADEIN);
}


function showLogoAndBar() {
  const logo = document.getElementById('intro-logo-container');
  if (logo) logo.style.opacity = 1;
  setTimeout(() => {
    const barBg = document.getElementById('intro-loading-bar-bg');
    if (barBg) barBg.style.opacity = 1;
    if (typeof startLoadingSequenceSync === 'function') startLoadingSequenceSync();
  }, LOGO_FADEIN);
}

// Synchronize status messages and loading bar
function startLoadingSequenceSync() {
  const bar = document.getElementById('intro-loading-bar');
  const status = document.getElementById('intro-status');
  if (!bar || !status) return;
  let idx = 0;
  let startTime = performance.now();
  let lastStatusTime = 0;
  status.style.opacity = 0;

  function animate(now) {
    let elapsed = now - startTime;
    let progress = Math.min(elapsed / LOADING_DURATION, 1);
    if (bar) bar.style.width = (progress * 100).toFixed(1) + "%";

    // Show status messages at the right time
    if (idx < STATUS_MESSAGES.length && elapsed >= STATUS_INTERVALS[idx]) {
      // Fade out previous
      if (idx > 0 && status) status.style.opacity = 0;
      setTimeout(() => {
        if (idx < STATUS_MESSAGES.length) {
          status.textContent = STATUS_MESSAGES[idx].text;
          status.className = STATUS_MESSAGES[idx].class;
          status.style.opacity = 1;
        }
      }, idx === 0 ? 0 : FADE_TIME);
      idx++;
      lastStatusTime = elapsed;
    }

    if (progress < 1) {
      requestAnimationFrame(animate);
    } else {
      // Wait for last status message, then fade out title, then go to login
      setTimeout(() => {
        fadeOutTitleAndGoToLogin();
      }, LOGIN_TRANSITION_DELAY);
    }
  }
  requestAnimationFrame(animate);
}

function fadeOutTitleAndGoToLogin() {
  const title = document.getElementById('intro-title');
  title.style.opacity = 0;
  setTimeout(() => {
    goToLogin();
  }, TITLE_FADEOUT);
}

// ---- STARS ----
let starCtx, stars = [];
function setupCanvas() {
  const canvas = document.getElementById('star-bg');
  function resize() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
  }
  window.addEventListener('resize', resize);
  resize();
  starCtx = canvas.getContext('2d');
  // Generate stars
  stars = [];
  const STAR_COUNT = Math.floor(window.innerWidth * window.innerHeight / 1800);
  for (let i = 0; i < STAR_COUNT; i++) {
    stars.push({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      r: Math.random() * 0.8 + 0.5,
      baseAlpha: Math.random() * 0.5 + 0.3,
      twinkle: Math.random() * Math.PI * 2,
      speed: Math.random() * 0.08 + 0.02,
      color: Math.random() < 0.7 ? "#eaf6ff" : "#7fd6ff"
    });
  }
}

function animateStars() {
  const canvas = document.getElementById('star-bg');
  starCtx.clearRect(0, 0, canvas.width, canvas.height);
  const t = performance.now() * 0.00025;
  for (let s of stars) {
    // Twinkle
    let alpha = s.baseAlpha + 0.25 * Math.sin(t * 2 + s.twinkle);
    // Yıldızları sağa ve aşağı kaydır (parallax)
    s.x += 0.08 + Math.sin(t + s.twinkle) * s.speed * 0.2;
    s.y += 0.04 + Math.cos(t + s.twinkle) * s.speed * 0.2;
    // Wrap
    if (s.x < 0) s.x += canvas.width;
    if (s.x > canvas.width) s.x -= canvas.width;
    if (s.y < 0) s.y += canvas.height;
    if (s.y > canvas.height) s.y -= canvas.height;
    // Draw
    starCtx.save();
    starCtx.globalAlpha = Math.max(0, Math.min(1, alpha));
    starCtx.beginPath();
    starCtx.arc(s.x, s.y, s.r, 0, 2 * Math.PI);
    starCtx.shadowColor = s.color;
    starCtx.shadowBlur = 8;
    starCtx.fillStyle = s.color;
    starCtx.fill();
    starCtx.restore();
  }
  // Shooting stars
  for (let i = shootingStars.length - 1; i >= 0; i--) {
    let s = shootingStars[i];
    starCtx.save();
    starCtx.globalAlpha = s.alpha;
    let dx = Math.cos(s.angle) * s.len;
    let dy = Math.sin(s.angle) * s.len;
    let grad = starCtx.createLinearGradient(s.x, s.y, s.x + dx, s.y + dy);
    grad.addColorStop(0, '#fff');
    grad.addColorStop(1, 'rgba(65,189,245,0)');
    starCtx.strokeStyle = grad;
    starCtx.lineWidth = 2.5;
    starCtx.beginPath();
    starCtx.moveTo(s.x, s.y);
    starCtx.lineTo(s.x + dx, s.y + dy);
    starCtx.stroke();
    starCtx.restore();
    s.x += Math.cos(s.angle) * s.speed;
    s.y += Math.sin(s.angle) * s.speed;
    s.alpha -= 0.012;
    if (s.alpha <= 0) shootingStars.splice(i, 1);
  }
  if (Math.random() < 0.012) spawnShootingStar();
  requestAnimationFrame(animateStars);
}

// ---- LOGIN TRANSITION ----
function goToLogin() {
  // Fade out intro
  const intro = document.getElementById('intro-root');
  intro.style.transition = "opacity 1.1s";
  intro.style.opacity = 0;
  setTimeout(() => {
    intro.style.display = "none";
    // Show pin pad login
    showPinPadLogin();
  }, 1100);
}

// ---- PIN PAD LOGIN ----
function showPinPadLogin() {
  // Show the star background again for premium effect
  const starBg = document.getElementById('star-bg');
  starBg.style.opacity = 1;

  // Add pinpad-active class to body for style/overflow
  document.body.classList.add('pinpad-active');

  // Create pin pad container (ultra-premium glassmorphism + neon + embossed)
  let pinPadHtml = `
    <div class="pin-container glass-glow ultra-premium-glass">
      <div class="wow-particles"></div>
      <div class="pin-display glass-glow">
        <div id="pinStatus" class="pin-status">••••</div>
      </div>
      <div class="pin-pad">
        <div class="pin-button glass-btn" data-value="1"><span class="embossed">1</span></div>
        <div class="pin-button glass-btn" data-value="2"><span class="embossed">2</span></div>
        <div class="pin-button glass-btn" data-value="3"><span class="embossed">3</span></div>
        <div class="pin-button glass-btn" data-value="4"><span class="embossed">4</span></div>
        <div class="pin-button glass-btn" data-value="5"><span class="embossed">5</span></div>
        <div class="pin-button glass-btn" data-value="6"><span class="embossed">6</span></div>
        <div class="pin-button glass-btn" data-value="7"><span class="embossed">7</span></div>
        <div class="pin-button glass-btn" data-value="8"><span class="embossed">8</span></div>
        <div class="pin-button glass-btn" data-value="9"><span class="embossed">9</span></div>
        <div class="pin-button glass-btn special theme-sky" id="clearButton" data-value="C"><span class="embossed">C</span></div>
        <div class="pin-button glass-btn" data-value="0"><span class="embossed">0</span></div>
        <div class="pin-button glass-btn special theme-sky" id="enterButton" data-value="✔"><span class="embossed">✔</span></div>
      </div>
      <button id="guest-access-btn" class="guest-access-btn">Misafir Erişim</button>
    </div>
  `;
  // Add to body
  let pinPadDiv = document.createElement('div');
  pinPadDiv.id = 'pinpad-root';
  pinPadDiv.innerHTML = pinPadHtml;
  document.body.appendChild(pinPadDiv);

  // Add floating particles for wow effect
  const particles = pinPadDiv.querySelector('.wow-particles');
  for (let i = 0; i < 8; i++) {
    const p = document.createElement('div');
    p.className = 'wow-particle';
    p.style.width = `${18 + Math.random() * 22}px`;
    p.style.height = p.style.width;
    p.style.left = `${10 + Math.random() * 80}%`;
    p.style.top = `${10 + Math.random() * 70}%`;
    p.style.animationDelay = `${Math.random() * 6}s`;
    particles.appendChild(p);
  }

  // Animate in
  setTimeout(() => {
    pinPadDiv.style.opacity = 1;
    pinPadDiv.style.transition = 'opacity 1s';
  }, 50);

  // Pin pad logic (premium version)
  let pinCode = '';
  const maxPinLength = 4;
  const pinDisplay = document.getElementById('pinStatus');
  const pinButtons = document.querySelectorAll('.pin-button');
  pinButtons.forEach(button => {
    button.addEventListener('click', () => {
      const value = button.dataset.value;
      if (value === 'C') {
        pinCode = '';
      } else if (value === '✔') {
        if (pinCode.length === maxPinLength) {
          checkPin(pinCode);
        } else {
          pinDisplay.classList.add('pulse-fail');
          setTimeout(() => pinDisplay.classList.remove('pulse-fail'), 800);
        }
      } else {
        if (pinCode.length < maxPinLength) {
          pinCode += value;
          if (pinCode.length === maxPinLength) {
            checkPin(pinCode);
          }
        }
      }
      // Girilen rakam kadar nokta göster
      pinDisplay.textContent = '•'.repeat(pinCode.length);
    });
  });

  function checkPin(pin) {
    if (pin === '1234') {
      pinDisplay.classList.add('pulse-success');
      setTimeout(() => {
        pinDisplay.classList.remove('pulse-success');
          window.location.href = 'home.html';
      }, 500);
    } else {
      pinDisplay.classList.add('pulse-fail');
      setTimeout(() => pinDisplay.classList.remove('pulse-fail'), 800);
      pinCode = '';
      pinDisplay.textContent = '';
    }
  }
}
