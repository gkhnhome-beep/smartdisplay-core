// index.js — extracted index page functions from main.js
// Provides `window.renderIntro()` and `window.renderLogin()` used by the SPA.

// INTRO SCREEN
window.renderIntro = function() {
    if (alarmLastState && typeof alarmLastState === 'object' && (alarmLastState.triggered || alarmLastState.state === 'triggered')) {
        return;
    }
    if (!window.__presenceSoundPlayed) {
        if (typeof playPresenceSound === 'function') playPresenceSound();
        else if (window.playPresenceSound) window.playPresenceSound();
    }
    const app = document.getElementById('app');
    if (!app) return;
    app.innerHTML = `
        <div style="height:100vh; display:flex; flex-direction:column; align-items:center; justify-content:center; background:#121417; color:#41bdf5;">
            <div style="font-size:3rem; font-weight:bold; margin-bottom:2rem;">Smart Display</div>
            <div style="width:120px; height:120px; border-radius:50%; background:radial-gradient(circle,#41bdf5 60%,#0a1a2f 100%); display:flex; align-items:center; justify-content:center; margin-bottom:2rem;">
                <img src="ha_logo.png" alt="Home Assistant Logo" style="width:80px; height:80px;">
            </div>
            <div style="width:220px; height:18px; background:#1a2a3d; border-radius:8px; margin-bottom:2rem; overflow:hidden;">
                <div id="introBar" style="height:100%; width:0%; background:linear-gradient(90deg,#41bdf5 0%,#65e3ff 100%); border-radius:8px; transition:width 1.5s;"></div>
            </div>
            <div style="font-size:1.2rem; opacity:0.7;">Sistem başlatılıyor...</div>
        </div>`;
    setTimeout(() => {
        const bar = document.getElementById('introBar');
        if (bar) bar.style.width = '100%';
    }, 100);
};

// LOGIN SCREEN
window.renderLogin = function() {
    const app = document.getElementById('app');
    app.innerHTML = `
        <div style="height:100vh; display:flex; align-items:center; justify-content:center; background:#121417;">
            <div style="background:#23272b; border-radius:18px; box-shadow:0 0 32px #41bdf5cc; padding:2.5rem 2.5rem; display:flex; flex-direction:column; align-items:center;">
                <div style="font-size:2rem; color:#41bdf5; margin-bottom:1.2rem;">PIN ile Giriş</div>
                <div id="pinDisplay" style="letter-spacing:12px; font-size:2.2rem; color:#41bdf5; background:rgba(65,189,245,0.08); border-radius:14px; width:180px; height:48px; display:flex; align-items:center; justify-content:center; margin-bottom:1.5rem; border:1px solid #41bdf533;">••••</div>
                <div id="pinPad" style="display:grid; grid-template-columns:repeat(3,1fr); gap:18px; width:240px; margin-bottom:1.2rem;">
                    ${[1,2,3,4,5,6,7,8,9,'C',0,'✔'].map(val => {
                        let style = val==='C' ? 'background:#f44;' : (val==='✔' ? 'background:#2ecc71;' : 'background:#23272b;');
                        return `<button class="pin-btn" data-value="${val}" style="height:60px; border-radius:12px; font-size:1.5rem; color:#fff; border:none; cursor:pointer; ${style}">${val}</button>`;
                    }).join('')}
                </div>
                <div id="loginError" style="min-height:1.5rem; color:#f44; font-weight:bold;"></div>
            </div>
        </div>`;

    let pin = "";
    const pinDisplay = document.getElementById('pinDisplay');
    const errorDiv = document.getElementById('loginError');
    document.querySelectorAll('.pin-btn').forEach(btn => {
        btn.onclick = () => {
            const val = btn.getAttribute('data-value');
            if (val === 'C') {
                pin = "";
                errorDiv.textContent = "";
            } else if (val === '✔') {
                if (pin.length !== 4) {
                    errorDiv.style.color = "#f44";
                    errorDiv.textContent = "PIN 4 haneli olmalı!";
                    return;
                }
                fetch('/api/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ pin })
                })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        State.currentUser = data.username || 'Kullanıcı';
                        State.role = data.role || 'guest';
                        State.pin = pin;
                        try {
                            sessionStorage.setItem('currentUser', State.currentUser);
                            sessionStorage.setItem('role', State.role);
                            sessionStorage.setItem('pin', State.pin);
                        } catch (e) {}
                        errorDiv.style.color = "#2ecc71";
                        errorDiv.textContent = "Giriş Başarılı!";
                        setTimeout(() => {
                            if (typeof renderMainLayout === 'function') {
                                renderMainLayout();
                            } else if (typeof routeContent === 'function') {
                                routeContent('#/home');
                            } else {
                                window.location.hash = '#/home';
                            }
                        }, 600);
                    } else {
                        errorDiv.style.color = "#f44";
                        errorDiv.textContent = data.message || "Hatalı PIN, lütfen tekrar deneyin.";
                        pin = "";
                        pinDisplay.textContent = pin.padEnd(4, '•');
                    }
                })
                .catch(() => {
                    errorDiv.style.color = "#f44";
                    errorDiv.textContent = "Sunucu hatası. Lütfen tekrar deneyin.";
                    pin = "";
                    pinDisplay.textContent = pin.padEnd(4, '•');
                });
            } else if (pin.length < 4 && !isNaN(val)) {
                pin += val;
                errorDiv.textContent = "";
            }
            pinDisplay.textContent = pin.padEnd(4, '•');
        };
    });
};

// End of index.js
