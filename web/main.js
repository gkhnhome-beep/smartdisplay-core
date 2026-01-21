// --- GLOBAL STATE tanımı ---
window.State = {
    currentUser: (typeof sessionStorage !== 'undefined' && sessionStorage.getItem('currentUser')) ? sessionStorage.getItem('currentUser') : 'Kullanıcı',
    role: (typeof sessionStorage !== 'undefined' && sessionStorage.getItem('role')) ? sessionStorage.getItem('role') : 'guest',
    pin: (typeof sessionStorage !== 'undefined' && sessionStorage.getItem('pin')) ? sessionStorage.getItem('pin') : '',
    clockInterval: null,
    introStarsFrame: null
};
// Note: emergency inline fallback removed — styles and staged startup code handle visibility now
// --- ALARM STATE tanımı ---
let alarmLastState = {};
const SETTINGS_PAGES = {
    // Sistem
    'Sistem::Ekran': {
        title: 'Ekran',
        description: 'Ekran parlaklığı ve yoğunluk ayarları burada yönetilecek.',
        type: 'placeholder'
    },
    'Sistem::Dil': {
        title: 'Dil',
        description: 'Sistem dili ve yerelleştirme ayarları.',
        type: 'placeholder'
    },
    'Sistem::Tarih & Saat': {
        title: 'Tarih & Saat',
        description: 'Cihazın tarih ve saat ayarları.',
        type: 'placeholder'
    },
    'Sistem::Ağ Durumu': {
        title: 'Ağ Durumu',
        description: 'Cihazın ağ bağlantı durumu ve bilgileri.',
        type: 'placeholder',
        render: function(content) {
            let netStatus = 'Unknown';
            if (window.navigator.onLine !== undefined) {
                netStatus = window.navigator.onLine ? 'Online' : 'Offline';
            }
            content.innerHTML = `<h2 style='color:var(--accent); font-size:1.2rem; margin-bottom:1.2rem;'>Ağ Durumu</h2><p style='color:var(--muted); margin-bottom:1.2rem;'>Cihazın ağ bağlantı durumu ve bilgileri.</p><div class='settings-placeholder'>Coming soon / Not available in this build</div><div style='margin-top:1.2rem; color:#7ec3e6;'>Durum: <b>${netStatus}</b></div>`;
        }
    },
    'Sistem::Cihaz Bilgisi': {
        title: 'Cihaz Bilgisi',
        description: 'Donanım modeli, yazılım sürümü ve diğer cihaz bilgileri.',
        type: 'placeholder',
        render: function(content) {
            let model = window.deviceModel || 'Unavailable';
            let build = window.deviceBuild || 'Unavailable';
            let version = window.deviceVersion || 'Unavailable';
            content.innerHTML = `<h2 style='color:var(--accent); font-size:1.2rem; margin-bottom:1.2rem;'>Cihaz Bilgisi</h2><p style='color:var(--muted); margin-bottom:1.2rem;'>Donanım modeli, yazılım sürümü ve diğer cihaz bilgileri.</p><div class='settings-placeholder'>Coming soon / Not available in this build</div><div style='margin-top:1.2rem; color:#7ec3e6;'>Model: <b>${model}</b> | Build: <b>${build}</b> | Sürüm: <b>${version}</b></div>`;
        }
    },
    // Home Assistant
    'Home Assistant::Connection': {
        title: 'Home Assistant Bağlantısı',
        description: 'Home Assistant bağlantı ve kimlik doğrulama ayarları.',
        type: 'custom',
        render: function(content) {
            content.innerHTML = `<div id='ha-connection-panel'></div>`;
            if (typeof setupHAForm === 'function') setupHAForm();
        }
    },
    'Home Assistant::Aydınlatma': {
        title: 'Aydınlatma',
        description: 'Home Assistant aydınlatma varlıkları eşlemesi.',
        type: 'placeholder'
    },
    'Home Assistant::Termostatlar': {
        title: 'Termostatlar',
        description: 'Home Assistant termostat varlıkları eşlemesi.',
        type: 'placeholder'
    },
    'Home Assistant::Alarmo': {
        title: 'Alarmo',
        description: 'Home Assistant Alarmo entegrasyonu.',
        type: 'placeholder'
    },
    'Home Assistant::Kameralar': {
        title: 'Kameralar',
        description: 'Home Assistant kamera varlıkları eşlemesi.',
        type: 'placeholder'
    },
    // Kullanıcı & Erişim
    'Kullanıcı & Erişim::Kullanıcılar': {
        title: 'Kullanıcılar',
        description: 'Kullanıcı ekle, düzenle, sil.',
        type: 'custom',
        render: renderUserManagement
    },
    'Kullanıcı & Erişim::Oturum Bilgisi': {
        title: 'Oturum Bilgisi',
        description: 'Aktif kullanıcı ve oturum bilgileri.',
        type: 'placeholder',
        render: function(content) {
            let username = State.currentUser || 'Bilinmiyor';
            let role = State.role || 'Bilinmiyor';
            let sessionStart = window.sessionStartTime ? new Date(window.sessionStartTime).toLocaleString() : 'Bilinmiyor';
            content.innerHTML = `<h2 style='color:var(--accent); font-size:1.2rem; margin-bottom:1.2rem;'>Oturum Bilgisi</h2><p style='color:var(--muted); margin-bottom:1.2rem;'>Aktif kullanıcı ve oturum bilgileri.</p><div style='margin-top:1.2rem; color:#7ec3e6;'>Kullanıcı: <b>${username}</b> | Rol: <b>${role}</b> | Başlangıç: <b>${sessionStart}</b></div>`;
        }
    },
    'Kullanıcı & Erişim::PIN Yönetimi': {
        title: 'PIN Yönetimi',
        description: 'Kullanıcı PIN yönetimi.',
        type: 'placeholder'
    },
    'Kullanıcı & Erişim::Misafir Erişimi': {
        title: 'Misafir Erişimi',
        description: 'Misafir kullanıcı erişim ayarları.',
        type: 'placeholder'
    },
    // Bakım
    'Bakım::OTA Güncelleme': {
        title: 'OTA Güncelleme',
        description: 'Cihaz yazılımı güncelleme işlemleri.',
        type: 'custom',
        render: function(content) {
            // Mevcut OTA paneli (varsa)
            if (typeof renderOTAPanel === 'function') {
                renderOTAPanel(content);
            } else {
                content.innerHTML = `<h2 style='color:var(--accent); font-size:1.2rem; margin-bottom:1.2rem;'>OTA Güncelleme</h2><p style='color:var(--muted); margin-bottom:1.2rem;'>Cihaz yazılımı güncelleme işlemleri.</p><div class='settings-placeholder'>Coming soon / Not available in this build</div>`;
            }
        }
    },
    'Bakım::Yedekleme & Geri Yükleme': {
        title: 'Yedekleme & Geri Yükleme',
        description: 'Cihaz yedekleme ve geri yükleme işlemleri.',
        type: 'placeholder'
    },
    'Bakım::Kayıtlar': {
        title: 'Kayıtlar',
        description: 'Sistem kayıtları ve loglar.',
        type: 'placeholder'
    },
    'Bakım::Yeniden Başlat / Kapat': {
        title: 'Yeniden Başlat / Kapat',
        description: 'Cihazı yeniden başlatma veya kapatma işlemleri.',
        type: 'placeholder'
    },
    // Hakkında
    'Hakkında::Yazılım Sürümü': {
        title: 'Yazılım Sürümü',
        description: 'Yüklü yazılım sürümü.',
        type: 'placeholder',
        render: function(content) {
            let version = window.deviceVersion || window.appVersion || 'Unavailable in this build';
            content.innerHTML = `<h2 style='color:var(--accent); font-size:1.2rem; margin-bottom:1.2rem;'>Yazılım Sürümü</h2><p style='color:var(--muted); margin-bottom:1.2rem;'>Yüklü yazılım sürümü.</p><div class='settings-placeholder'>Coming soon / Not available in this build</div><div style='margin-top:1.2rem; color:#7ec3e6;'>Sürüm: <b>${version}</b></div>`;
        }
    },
    'Hakkında::Donanım Modeli': {
        title: 'Donanım Modeli',
        description: 'Cihaz donanım modeli.',
        type: 'placeholder',
        render: function(content) {
            let model = window.deviceModel || 'Unavailable in this build';
            content.innerHTML = `<h2 style='color:var(--accent); font-size:1.2rem; margin-bottom:1.2rem;'>Donanım Modeli</h2><p style='color:var(--muted); margin-bottom:1.2rem;'>Cihaz donanım modeli.</p><div class='settings-placeholder'>Coming soon / Not available in this build</div><div style='margin-top:1.2rem; color:#7ec3e6;'>Model: <b>${model}</b></div>`;
        }
    },
    'Hakkında::Lisanslar': {
        title: 'Lisanslar',
        description: 'Yasal bilgiler ve açık kaynak lisansları.',
        type: 'placeholder',
        render: function(content) {
            content.innerHTML = `<h2 style='color:var(--accent); font-size:1.2rem; margin-bottom:1.2rem;'>Lisanslar</h2><p style='color:var(--muted); margin-bottom:1.2rem;'>Yasal bilgiler ve açık kaynak lisansları.</p><div class='settings-placeholder'>Coming soon / Not available in this build</div><div style='margin-top:1.2rem; color:#7ec3e6;'>MIT License, Go, Home Assistant, Alarmo, ...</div>`;
        }
    }
};
// --- Settings Menu Data (must be top-level for all functions) ---
const SETTINGS_MENU = [
    { section: 'Sistem', items: ['Ekran', 'Dil', 'Tarih & Saat', 'Ağ Durumu', 'Cihaz Bilgisi'] },
    { section: 'Home Assistant', items: ['Connection', 'Aydınlatma', 'Termostatlar', 'Alarmo', 'Kameralar'] },
    { section: 'Kullanıcı & Erişim', items: ['Kullanıcılar', 'Oturum Bilgisi', 'PIN Yönetimi', 'Misafir Erişimi'] },
    { section: 'Bakım', items: ['OTA Güncelleme', 'Yedekleme & Geri Yükleme', 'Kayıtlar', 'Yeniden Başlat / Kapat'] },
    { section: 'Hakkında', items: ['Yazılım Sürümü', 'Donanım Modeli', 'Lisanslar'] }
];

// --- FAZ A6: Basit i18n sözlüğü ---
const STRINGS = {
    tr: {
        alarm: {
            disarmed: 'Devre Dışı',
            arming: 'Çıkış Gecikmesi',
            pending: 'Giriş Gecikmesi',
            armed_home: 'Evde Kurulu',
            armed_away: 'Dışarıda Kurulu',
            armed_night: 'Gece Kurulu',
            triggered: 'Tetiklendi!',
            exit_delay: 'Exit delay active.',
            entry_delay: 'Entry delay active.',
            armed_msg: 'Alarm is armed. Disarm required to continue.',
            triggered_msg: 'Alarm triggered.'
        },
        error: {
            unreachable: 'Alarm sistemine ulaşılamıyor.',
            invalid: 'Geçersiz istek.',
            triggered: 'Alarm tetiklendi, işlem engellendi.',
            unknown: 'Bilinmeyen hata.',
            network: 'Ağ hatası.'
        },
        ui: {
            waiting: 'Waiting for alarm state update…',
            welcome: 'Hoş geldiniz!',
            notfound: 'Sayfa bulunamadı',
            version: 'Sürüm: v',
            system_status: 'Sistem Durumu',
            backend_ok: 'Erişilebilir',
            backend_fail: 'Erişilemiyor',
            last_poll: 'Son başarılı poll:'
        }
    }
};
const LANG = 'tr';
function t(path) {
    return path.split('.').reduce((o, k) => (o||{})[k], STRINGS[LANG]) || path;
}
// index page functions (intro + login) moved to web/index.js and loaded dynamically below

// Load index page script (provides window.renderIntro and window.renderLogin)
// Ensure index.js is loaded; callback invoked after load (or immediately if already present)
function ensureIndexLoaded(cb) {
    if (window.renderIntro && window.renderLogin) {
        if (typeof cb === 'function') cb();
        return;
    }
    if (window.__indexScriptInjected) {
        // script already injected but functions not ready yet; poll briefly
        const start = Date.now();
        const iv = setInterval(() => {
            if (window.renderIntro && window.renderLogin) {
                clearInterval(iv);
                if (typeof cb === 'function') cb();
            } else if (Date.now() - start > 3000) { // timeout 3s
                clearInterval(iv);
                console.error('[Loader] index.js did not expose expected functions in time');
                if (typeof cb === 'function') cb();
            }
        }, 50);
        return;
    }
    window.__indexScriptInjected = true;
    try {
        const s = document.createElement('script');
        s.src = './index.js';
        s.async = false;
        s.defer = false;
        s.onload = () => {
            if (window.renderIntro && window.renderLogin) {
                if (typeof cb === 'function') cb();
            } else {
                // Give a short grace period for the script to initialize
                setTimeout(() => { if (typeof cb === 'function') cb(); }, 50);
            }
        };
        s.onerror = () => { console.error('[Loader] Failed to load index.js'); if (typeof cb === 'function') cb(); };
        document.head.appendChild(s);
    } catch (e) {
        console.error('[Loader] Exception while injecting index.js', e);
        if (typeof cb === 'function') cb();
    }
}
// Ensure settings.js is loaded; call cb when ready
function ensureSettingsLoaded(cb) {
    if (window.mountSettingsView && window.renderSettingsSidebar && window.openSettingsItem) {
        if (typeof cb === 'function') cb();
        return;
    }
    if (window.__settingsScriptInjected) {
        const start = Date.now();
        const iv = setInterval(() => {
            if (window.mountSettingsView && window.renderSettingsSidebar) {
                clearInterval(iv);
                if (typeof cb === 'function') cb();
            } else if (Date.now() - start > 3000) {
                clearInterval(iv);
                console.error('[Loader] settings.js did not expose expected functions in time');
                if (typeof cb === 'function') cb();
            }
        }, 50);
        return;
    }
    window.__settingsScriptInjected = true;
    try {
        const s = document.createElement('script');
        s.src = './settings.js';
        s.async = false;
        s.defer = false;
        s.onload = () => { if (typeof cb === 'function') cb(); };
        s.onerror = () => { console.error('[Loader] Failed to load settings.js'); if (typeof cb === 'function') cb(); };
        document.head.appendChild(s);
    } catch (e) {
        console.error('[Loader] Exception while injecting settings.js', e);
        if (typeof cb === 'function') cb();
    }
}
// --- 1. GLOBAL DURUM (STATE) ---
// Restore State from sessionStorage if available
const State = {
    currentUser: (typeof sessionStorage !== 'undefined' && sessionStorage.getItem('currentUser')) ? sessionStorage.getItem('currentUser') : 'Kullanıcı',
    role: (typeof sessionStorage !== 'undefined' && sessionStorage.getItem('role')) ? sessionStorage.getItem('role') : 'guest',
    clockInterval: null,
    introStarsFrame: null
};

// Temizlik fonksiyonu
const clearAllIntervals = () => {
    if (State.clockInterval) clearInterval(State.clockInterval);
    if (State.introStarsFrame) cancelAnimationFrame(State.introStarsFrame);
};

// --- 2. MERKEZİ YÖNLENDİRİCİ (ROUTER) ---
// Sayfa yüklendiğinde State'i sessionStorage'dan tekrar yükle
function restoreStateFromSession() {
    if (typeof sessionStorage !== 'undefined') {
        if (sessionStorage.getItem('role')) State.role = sessionStorage.getItem('role');
        if (sessionStorage.getItem('currentUser')) State.currentUser = sessionStorage.getItem('currentUser');
    }
}

restoreStateFromSession();

function router() {
    const hash = window.location.hash;
    const app = document.getElementById('app');
    if (!app) {
        console.error("Hata: 'app' ID'li element bulunamadı. Lütfen index.html dosyanızı kontrol edin.");
        return;
    }

    // --- ALARM LOCKDOWN: If alarm is triggered, forcibly show only alarm screen, hide layout ---
    if (alarmLastState && typeof alarmLastState === 'object' && (alarmLastState.triggered || alarmLastState.state === 'triggered')) {
        // Remove any main layout DOM to prevent duplication
        const app = document.getElementById('app');
        if (app) {
            // Remove sidebar and topbar if present
            const sidebar = document.getElementById('sidebar');
            if (sidebar) sidebar.remove();
            const topbar = document.getElementById('topbar-status');
            if (topbar) topbar.style.display = 'none';
            // Remove guest topbar if present
            const guestTopbar = document.getElementById('guest-topbar');
            if (guestTopbar) guestTopbar.remove();
            // Only mount alarm screen if not already present
            let mainContent = document.getElementById('main-content');
            if (!mainContent) {
                app.innerHTML = `<main id="main-content" style="flex:1; padding:2rem; overflow-y:auto;"></main>`;
                mainContent = document.getElementById('main-content');
            }
            renderAlarmScreen(mainContent);
        }
        return;
    }

    // İlk girişte veya hash yoksa intro göster, sonra login'e yönlendir
    if (!hash || hash === '#/' || hash === '') {
        // Remove sidebar/topbar/guest-topbar to avoid layout duplication
        const sidebar = document.getElementById('sidebar');
        if (sidebar) sidebar.remove();
        const topbar = document.getElementById('topbar-status');
        if (topbar) topbar.style.display = 'none';
        const guestTopbar = document.getElementById('guest-topbar');
        if (guestTopbar) guestTopbar.remove();
        ensureIndexLoaded(() => { if (window.renderIntro) window.renderIntro(); });
        setTimeout(() => {
            // Alarm tetiklendiyse introdan sonra da alarm ekranına yönlendir
            if (alarmLastState && typeof alarmLastState === 'object' && (alarmLastState.triggered || alarmLastState.state === 'triggered')) {
                window.location.hash = '#/alarm';
                // Alarm lockdown will handle layout
            } else {
                window.location.hash = '#/login';
            }
        }, 1800); // 1.8 saniye intro göster
        return;
    }
    // --- HOTFIX: First boot, HA not configured, force route to Settings
    // Detect HA not configured (first boot)
    if (window.haState && window.haState.isConfigured === false) {
        // Only reroute if not already on settings
        if (hash !== '#/settings') {
            // Remove sidebar/topbar/guest-topbar to avoid layout duplication
            const sidebar = document.getElementById('sidebar');
            if (sidebar) sidebar.remove();
            const topbar = document.getElementById('topbar-status');
            if (topbar) topbar.style.display = 'none';
            const guestTopbar = document.getElementById('guest-topbar');
            if (guestTopbar) guestTopbar.remove();
            console.log('[ViewManager] HA not configured, routing to Settings');
            window.location.hash = '#/settings';
            return;
        }
    }

    // --- FAZ A6: KIOSK LONG-RUN ERGONOMICS ---
    let idleTimeout = null;
    let idleStart = null;
    let driftInterval = null;
    let driftX = 0, driftY = 0;
    const DRIFT_PIXELS = 2;
    const DRIFT_INTERVAL = 90 * 1000; // 90s
    const IDLE_DIM_TIMEOUT = 180 * 1000; // 3dk

    function resetIdle() {
        idleStart = Date.now();
        document.body.classList.remove('idle-dim');
        if (idleTimeout) clearTimeout(idleTimeout);
        idleTimeout = setTimeout(() => {
            document.body.classList.add('idle-dim');
        }, IDLE_DIM_TIMEOUT);
    }

    function startUIDrift() {
        if (driftInterval) clearInterval(driftInterval);
        driftInterval = setInterval(() => {
            // Alarm ekranı ve triggered durumda drift yapma
            const alarmRoot = document.getElementById('alarm-root');
            if (alarmRoot && (alarmLastState?.triggered || alarmLastState?.state === 'triggered')) return;
            driftX = (driftX + 1) % (DRIFT_PIXELS + 1);
            driftY = (driftY + 1) % (DRIFT_PIXELS + 1);
            document.body.style.transform = `translate(${driftX}px,${driftY}px)`;
        }, DRIFT_INTERVAL);
    }

    function stopUIDrift() {
        if (driftInterval) clearInterval(driftInterval);
        document.body.style.transform = '';
    }

    ['mousemove','keydown','mousedown','touchstart'].forEach(evt => {
        window.addEventListener(evt, () => {
            resetIdle();
            stopUIDrift();
            startUIDrift();
        });
    });

    resetIdle();
    startUIDrift();
    if (hash === '#/login') {
        // Eski PIN modalı/giriş formunu göster
        ensureIndexLoaded(() => { if (window.renderLogin) window.renderLogin(); });
        return;
    }
    // --- Settings view mount logic ---
    if (hash === '#/settings') {
        // Dynamically load settings view implementation
        ensureSettingsLoaded(() => {
            if (window.mountSettingsView) {
                window.mountSettingsView(app);
                if (!window.settingsSidebarState || !window.settingsSidebarState.expanded) {
                    window.settingsSidebarState = { expanded: SETTINGS_MENU[0].section };
                }
                if (window.renderSettingsSidebar) window.renderSettingsSidebar();
                setTimeout(() => {
                    let firstSettingsItem;
                    try {
                        firstSettingsItem = document.querySelector('.settings-submenu[style*="block"] .settings-item');
                        if (firstSettingsItem) firstSettingsItem.click();
                    } catch (e) {
                        // ignore
                    }
                }, 0);
            } else {
                console.error('[Loader] settings.js did not provide mountSettingsView');
            }
        });
        return;
    }
    // ...other view logic for other hashes (e.g. #/home, #/alarm, etc.)...
    // ...eski settings-menu-item ve renderSettingsPage kodları kaldırıldı
}

// Layout mode setter (dummy implementation)
function setLayoutMode(mode) {
    console.log("Layout mode set:", mode);
}

function renderSettingsSection(section) {
    const content = document.getElementById('settings-content');
    if (!content) return;
    if (section === 'genel') {
        content.innerHTML = `<h3>Genel Ayarlar</h3><p>Genel sistem ayarları burada olacak.</p>`;
    } else if (section === 'kullanici') {
        content.innerHTML = `<h3>Kullanıcı Ayarları</h3><p>Kullanıcı ile ilgili ayarlar burada olacak.</p>`;
    } else if (section === 'sistem') {
        content.innerHTML = `<h3>Sistem Ayarları</h3><p>Sistem ile ilgili ayarlar burada olacak.</p>`;
    }
}

// --- KULLANICI & ERİŞİM: Kullanıcı Yönetimi Paneli ---
function renderUserManagement(content) {
        console.log("renderUserManagement çalıştı");
    content.innerHTML = `<h2 style='color:var(--accent); font-size:1.2rem; margin-bottom:1.2rem;'>Kullanıcı Yönetimi</h2>
        <div id='userList'></div>
        <button id='addUserBtn' style='margin:1rem 0 2rem 0;'>+ Yeni Kullanıcı Ekle</button>
        <div id='userFormPanel' style='display:none;'></div>
        <div id='userFormSuccess' style='color:#2ecc71; margin-top:0.7rem; font-weight:bold; display:none;'></div>`;
    loadUserList();
    document.getElementById('addUserBtn').onclick = () => showUserForm();
    // Event delegation for user row clicks
    document.getElementById('userList').onclick = function(e) {
        const row = e.target.closest('tr[data-username]');
        if (row) {
            const username = row.getAttribute('data-username');
            editUser(username);
        }
    };
}

function loadUserList() {
    fetch('/api/users/list', {
        method: 'GET',
        headers: {
            'X-User-Role': (window.State && State.role) ? State.role : '',
            'X-User-Pin': (window.State && State.pin) ? State.pin : ''
        },
        credentials: 'same-origin'
    })
    .then(res => {
        if (!res.ok) {
            document.getElementById('userList').innerHTML = '<div style="color:#f44;">Kullanıcı listesi alınamadı (404 veya sunucu hatası).</div>';
            return Promise.reject('Kullanıcı listesi alınamadı');
        }
        return res.json();
    })
    .then(data => {
        if (!data.success) return;
        const userList = document.getElementById('userList');
        userList.innerHTML = `<table style='width:100%;margin-bottom:1.2rem;'><tr><th>Kullanıcı Adı</th><th>Rol</th><th>PIN</th><th>İşlem</th></tr>` +
            data.users.map(u => `<tr data-username="${u.username}"><td class="user-clickable" style="cursor:pointer;color:#41bdf5;">${u.username}</td><td>${u.role}</td><td>${u.pin}</td><td>
            <button onclick='deleteUser("${u.username}")' style='color:#f44;'>Sil</button></td></tr>`).join('') + '</table>';
    });
}

function showUserForm(user) {
    const panel = document.getElementById('userFormPanel');
    panel.style.display = 'block';
    panel.innerHTML = `<div style='margin-bottom:0.5rem;'>
        <input id='userFormUsername' placeholder='Kullanıcı Adı' value='${user ? user.Username : ''}' ${user ? 'readonly' : ''} />
        <input id='userFormPIN' placeholder='PIN' value='${user ? user.PIN : ''}' />
        <select id='userFormRole'>
            <option value='admin' ${user && user.Role==='admin' ? 'selected' : ''}>admin</option>
            <option value='user' ${user && user.Role==='user' ? 'selected' : ''}>user</option>
            <option value='guest' ${user && user.Role==='guest' ? 'selected' : ''}>guest</option>
        </select>
        <button id='saveUserBtn'>Kaydet</button>
        <button id='cancelUserBtn'>İptal</button>
    </div><div id='userFormError' style='color:#f44;'></div>`;
    document.getElementById('saveUserBtn').onclick = () => {
        const username = document.getElementById('userFormUsername').value.trim();
        const pin = document.getElementById('userFormPIN').value.trim();
        const role = document.getElementById('userFormRole').value;
        if (!username || !pin) {
            document.getElementById('userFormError').textContent = 'Kullanıcı adı ve PIN zorunlu!';
            return;
        }
        const payload = { Username: username, PIN: pin, Role: role };
        fetch(user ? '/api/users/update' : '/api/users/add', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-Role': (window.State && State.role) ? State.role : '',
                'X-User-Pin': (window.State && State.pin) ? State.pin : ''
            },
            body: JSON.stringify(payload),
            credentials: 'same-origin'
        })
        .then(res => {
            if (!res.ok) {
                document.getElementById('userFormError').textContent = 'Kullanıcı eklenemedi (404 veya sunucu hatası).';
                return Promise.reject('Kullanıcı eklenemedi');
            }
            return res.json();
        })
        .then(data => {
            if (data.success) {
                panel.style.display = 'none';
                loadUserList();
                // Show success message
                const succ = document.getElementById('userFormSuccess');
                if (succ) {
                    succ.textContent = user ? 'Güncelleme işlemi başarılı.' : 'Kullanıcı eklendi.';
                    succ.style.display = 'block';
                    setTimeout(() => { succ.style.display = 'none'; }, 2000);
                }
            } else {
                document.getElementById('userFormError').textContent = data.message || 'Hata oluştu';
            }
        });
    };
    document.getElementById('cancelUserBtn').onclick = () => {
        panel.style.display = 'none';
    };
}

function editUser(username) {
    fetch('/api/users/list', {
        method: 'GET',
        headers: { 'X-User-Role': (window.State && State.role) ? State.role : '' },
        credentials: 'same-origin'
    })
    .then(res => res.json())
    .then(data => {
        const user = data.users.find(u => u.Username === username);
        if (user) showUserForm(user);
    });
}

function deleteUser(username) {
    if (!confirm('Kullanıcı silinsin mi?')) return;
    fetch('/api/users/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username }),
        credentials: 'same-origin'
    })
    .then(res => res.json())
    .then(data => {
        if (data.success) loadUserList();
        else alert(data.message || 'Silme hatası');
    });
}

// Kullanıcı & Erişim menüsüne paneli bağla
// Kullanıcılar menüsü için render fonksiyonu zaten yukarıda eklendi.
// --- FOOTER/DIAGNOSTIC OVERLAY ---
function renderFooterVersion() {
    let footer = document.getElementById('footer-version');
    if (!footer) {
        footer = document.createElement('div');
        footer.id = 'footer-version';
        footer.style = 'position:fixed;bottom:0;right:0;padding:6px 16px;background:#121417cc;color:#b0c4d4;font-size:0.98em;z-index:9999;user-select:none;pointer-events:none;border-top-left-radius:8px;';
        document.body.appendChild(footer);
    }
    fetch('http://localhost:8090/api/admin/update/status', { credentials: 'same-origin' })
        .then(r => r.json())
        .then(data => {
            if (data && data.ok && data.data && data.data.current_version) {
                footer.innerHTML = 'v' + data.data.current_version +
                  ' <span id="footer-diagnostic" style="margin-left:1.2rem;text-decoration:underline;cursor:pointer;color:#41bdf5;">Sistem</span>';
                document.getElementById('footer-diagnostic').onclick = renderDiagnosticOverlay;
            }
        });
}

// Diagnostic overlay (health/status)
function renderDiagnosticOverlay() {
    let diag = document.getElementById('diagnostic-overlay');
    if (!diag) {
        diag = document.createElement('div');
        diag.id = 'diagnostic-overlay';
        diag.style = 'position:fixed;top:0;left:0;width:100vw;height:100vh;background:rgba(18,20,23,0.92);color:#fff;z-index:10000;display:flex;align-items:center;justify-content:center;flex-direction:column;';
        diag.innerHTML = '<div style="font-size:2.2rem;margin-bottom:1.2rem;">Sistem Durumu</div><div id="diag-content">Yükleniyor…</div><button id="diag-close" style="margin-top:2rem;padding:0.7rem 2.2rem;font-size:1.1rem;border-radius:8px;border:none;background:#41bdf5;color:#121417;cursor:pointer;">Kapat</button>';
        document.body.appendChild(diag);
        document.getElementById('diag-close').onclick = () => { diag.remove(); };
    }
    // Backend ve alarmo durumu için fetch
    fetch('http://localhost:8090/api/ui/alarm/state')
        .then(r => r.json())
        .then(data => {
            let alarmOk = (data && data.ok && data.result);
            let lastPoll = new Date().toLocaleString();
            document.getElementById('diag-content').innerHTML = `
                <div>Backend: <b style="color:${alarmOk?'#2ecc71':'#e74c3c'}">${alarmOk?'Erişilebilir':'Erişilemiyor'}</b></div>
                <div>Alarmo: <b style="color:${alarmOk?'#2ecc71':'#e74c3c'}">${alarmOk?'Erişilebilir':'Erişilemiyor'}</b></div>
                <div>Son başarılı poll: <b>${lastPoll}</b></div>
            `;
        })
        .catch(() => {
            document.getElementById('diag-content').innerHTML = '<div>Backend: <b style="color:#e74c3c">Erişilemiyor</b></div><div>Alarmo: <b style="color:#e74c3c">Erişilemiyor</b></div><div>Son başarılı poll: <b>-</b></div>';
        });
}

// --- 3. ANA YAPI (LAYOUT) ---
function renderMainLayout() {
    const app = document.getElementById('app');
    if (!app) return;
    // Role-based sidebar visibility
    if (window.__layoutMode === 'guest' || State.role === 'guest') {
        // Hide sidebar entirely in guest mode
        app.innerHTML = `
            <main id="main-content" style="flex:1; padding:2rem; overflow-y:auto;"></main>
        `;
        setLayoutMode('guest');
        return;
    }
    // Sidebar items by role
    const isAdmin = State.role === 'admin';
    const isUser = State.role === 'user';
    // Settings only for admin
    let settingsMenu = isAdmin ? '<li class="menu-item" data-link="#/settings" data-view="#/settings">⚙️ Sistem</li>' : '';
    // Sidebar HTML
    app.innerHTML = `
        <div style="display:flex; min-height:100vh; background:#121417; color:white; font-family:sans-serif;">
            <nav id="sidebar" style="width:250px; background:#1c1f26; border-right:1px solid #333; display:flex; flex-direction:column;">
                <div style="padding:2rem; font-size:1.5rem; font-weight:bold; color:#41bdf5;">Smart Display</div>
                <ul id="menu-list" style="list-style:none; padding:0; margin:0; flex:1;">
                    <li class="menu-item" data-link="#/home" data-view="#/home">🏠 Ana Sayfa</li>
                    <li class="menu-item" data-link="#/alarm" data-view="#/alarm">🔔 Alarm</li>
                    <li class="menu-item" data-link="#/climate" data-view="#/climate">🌡️ İklim</li>
                    <li class="menu-item" data-link="#/lights" data-view="#/lights">💡 Aydınlatma</li>
                    <li class="menu-item" data-link="#/energy" data-view="#/energy">⚡ Enerji</li>
                    ${settingsMenu}
                </ul>
                <div style="border-top:1.5px solid #333; margin:0 0 0 0; padding:0;"></div>
                <div class="menu-item sidebar-logout" data-link="#/login" data-view="#/login" style="color:#ff4444; margin-top:auto; padding:1.2rem 2rem; font-weight:600; font-size:1.08rem; display:block; cursor:pointer;">🚪 Çıkış Yap</div>
            </nav>
            <main id="main-content" class="home-surface" style="flex:1; padding:2rem; overflow-y:auto;"></main>
        </div>
    `;
    setupSidebarEvents();
    setLayoutMode(window.__layoutMode || 'normal');
}

function setupSidebarEvents() {
    // Sadece event delegation ile click yönetimi
    const menuList = document.getElementById('menu-list');
    if (menuList) {
        menuList.onclick = function(e) {
            const item = e.target.closest('.menu-item');
            if (item && item.getAttribute('data-link')) {
                window.location.hash = item.getAttribute('data-link');
            }
        };
    }
}

// --- Sidebar Active Highlight ---
// Usage: setActiveSidebarItem(routeOrViewName)
function setActiveSidebarItem(routeOrViewName) {
    // Accepts either hash (e.g. '#/home') or view name
    const items = document.querySelectorAll('#sidebar .menu-item');
    items.forEach(item => {
        const view = item.getAttribute('data-view') || item.getAttribute('data-link');
        if (view === routeOrViewName) {
            item.classList.add('active');
            // Subtle highlight: muted background, no neon, no strong accent
            item.style.background = 'rgba(77,179,250,0.08)';
            item.style.color = '#4db3fa';
            item.style.borderLeft = '4px solid #7fcfa0';
        } else {
            item.classList.remove('active');
            item.style.background = '';
            item.style.color = item.classList.contains('sidebar-logout') ? '#ff4444' : '#8fa1b3';
            item.style.borderLeft = '4px solid transparent';
        }
        // Settings item: hide or disable for user
        if (item.getAttribute('data-view') === '#/settings' && State.role === 'user') {
            item.style.pointerEvents = 'none';
            item.style.opacity = '0.5';
        } else {
            item.style.pointerEvents = '';
            item.style.opacity = '';
        }
    });
}

// --- 4. İÇERİK YÖNETİCİSİ ---
function routeContent(hash) {
    const container = document.getElementById('main-content');
    if (!container) return;
    // Remove home-surface class for all views, add only for Home
    container.classList.remove('home-surface');
    if (hash === '#/home') {
        container.classList.add('home-surface');
        // --- HOME SYSTEM STATUS HERO PANEL (event-driven, read-only) ---
        // Remove any previous hero panel
        let hero = document.getElementById('homeHero');
        if (hero) hero.remove();
        hero = document.createElement('section');
        hero.className = 'home-hero';
        hero.id = 'homeHero';
        container.prepend(hero);
        // Initial render (store-driven updates will follow)
        if (window.__store && typeof window.setupHomeHeroEventDriven === 'function') {
            window.setupHomeHeroEventDriven(window.__store);
        }
        // Home panel content (below hero)
        let homePanel = document.getElementById('homePanel');
        if (homePanel) homePanel.remove();
        homePanel = document.createElement('div');
        homePanel.className = 'home-panel';
        homePanel.id = 'homePanel';
        homePanel.style = 'max-width:520px;margin:3.5rem auto 0 auto;';
        homePanel.innerHTML = `<h2 style="margin-bottom:0.7em;">Ana Sayfa</h2><p style="font-size:1.13rem; color:var(--muted,#8fa1b3);">${t('ui.welcome')}</p>`;
        container.appendChild(homePanel);
    } else if (hash === '#/alarm') {
        // --- Home Hero State Adapter ---
        function deriveHeroModel(state) {
            // Robustly pick values, fallback to unknown/"—"
            let net = 'unknown';
            if (state?.network?.isOnline !== undefined) net = state.network.isOnline ? 'ok' : 'bad';
            else if (state?.net?.isOnline !== undefined) net = state.net.isOnline ? 'ok' : 'bad';
            else if (state?.connection?.online !== undefined) net = state.connection.online ? 'ok' : 'bad';
            // Home Assistant
            let ha = 'unknown', haConfigured = undefined;
            if (state?.haState) {
                haConfigured = state.haState.isConfigured;
                if (haConfigured === false) ha = 'not_configured';
                else if (state.haState.isConnected === true) ha = 'ok';
                else if (state.haState.isConnected === false) ha = 'bad';
            } else if (state?.settings?.haState) {
                haConfigured = state.settings.haState.isConfigured;
                if (haConfigured === false) ha = 'not_configured';
                else if (state.settings.haState.isConnected === true) ha = 'ok';
                else if (state.settings.haState.isConnected === false) ha = 'bad';
            } else if (state?.homeAssistant) {
                haConfigured = state.homeAssistant.isConfigured;
                if (haConfigured === false) ha = 'not_configured';
                else if (state.homeAssistant.isConnected === true) ha = 'ok';
                else if (state.homeAssistant.isConnected === false) ha = 'bad';
            }
            // System Health
            let sys = 'unknown';
            if (state?.systemHealth?.level) sys = state.systemHealth.level;
            else if (state?.health?.status) sys = state.health.status;
            else if (state?.device?.health) sys = state.device.health;
            // Last Sync
            let lastSync = '—';
            if (state?.haState?.lastSyncAt) lastSync = state.haState.lastSyncAt;
            else if (state?.sync?.lastAt) lastSync = state.sync.lastAt;
            else if (state?.settings?.lastSyncAt) lastSync = state.settings.lastSyncAt;
            // Uptime
            let uptime = '—';
            if (state?.system?.uptime) uptime = state.system.uptime;
            else if (state?.device?.uptimeSeconds) uptime = state.device.uptimeSeconds + 's';
            // Calm message
            let calmMsg = 'Durum bilgisi bekleniyor.';
            if (sys === 'ok') calmMsg = 'Sistem stabil çalışıyor.';
            else if (sys === 'warn') calmMsg = 'Bazı servislerde geçici sorun var.';
            else if (sys === 'bad') calmMsg = 'Kritik durum algılandı.';
            // Map for display
            function label(val, type) {
                if (type === 'sys') {
                    if (val === 'ok') return 'Healthy';
                    if (val === 'warn') return 'Degraded';
                    if (val === 'bad') return 'Critical';
                    return 'Unknown';
                }
                if (type === 'net') {
                    if (val === 'ok') return 'Online';
                    if (val === 'bad') return 'Offline';
                    return 'Unknown';
                }
                if (type === 'ha') {
                    if (val === 'ok') return 'Connected';
                    if (val === 'bad') return 'Disconnected';
                    if (val === 'not_configured') return 'Not configured';
                    return 'Unknown';
                }
                return val;
            }
            function statusClass(val, type) {
                if (type === 'ha' && val === 'not_configured') return 'hero-status-bad';
                if (val === 'ok') return 'hero-status-ok';
                if (val === 'warn') return 'hero-status-warn';
                if (val === 'bad') return 'hero-status-bad';
                return 'hero-status-unknown';
            }
            return {
                sys, net, ha, lastSync, uptime, calmMsg,
                sysLabel: label(sys, 'sys'),
                netLabel: label(net, 'net'),
                haLabel: label(ha, 'ha'),
                sysClass: statusClass(sys, 'sys'),
                netClass: statusClass(net, 'net'),
                haClass: statusClass(ha, 'ha'),
            };
        }

        // --- Home Hero Render ---
        function renderHomeHero(model) {
            const hero = document.getElementById('homeHero');
            if (!hero) return;
            hero.innerHTML = `
              <div class="hero-row">
                <div class="hero-chip ${model.sysClass}"><span class="icon">🖥️</span> <span>Sistem: <b>${model.sysLabel}</b></span></div>
                <div class="hero-chip ${model.netClass}"><span class="icon">🌐</span> <span>Ağ: <b>${model.netLabel}</b></span></div>
                <div class="hero-chip ${model.haClass}"><span class="icon">🏠</span> <span>HA: <b>${model.haLabel}</b></span></div>
                <div class="hero-chip"><span class="icon">⏱️</span> <span>Son Sync: <b>${model.lastSync}</b></span></div>
              </div>
              <div class="hero-row">
                <div class="hero-chip"><span class="icon">⏳</span> <span>Uptime: <b>${model.uptime}</b></span></div>
                <div class="hero-chip" style="grid-column: span 3; background:none; box-shadow:none; border:none;"></div>
              </div>
              <div class="hero-calm-msg">${model.calmMsg}</div>
            `;
        }

        // --- Home Hero Event-driven Wiring ---
        function setupHomeHeroEventDriven(store) {
            if (window.__homeHeroUnsub) window.__homeHeroUnsub();
            let prevKey = '';
            function update() {
                const state = store.getState ? store.getState() : store.state;
                const model = deriveHeroModel(state);
                const key = JSON.stringify([model.sys, model.net, model.ha, model.lastSync, model.uptime, model.calmMsg]);
                if (key !== prevKey) {
                    renderHomeHero(model);
                    prevKey = key;
                }
            }
            const unsub = store.subscribe(update);
            window.__homeHeroUnsub = unsub;
            // Initial render
            update();
        }

        // --- ALARM STATE POLLING (A5.2) ---
        function pollAlarmState() {
            fetch('/api/alarm/state')
                .then(response => response.json())
                .then(data => {
                    // ... kodlar ...
                    const err = document.getElementById('alarm-error');
                    if (err && !alarmActionPending) err.textContent = '';
                })
                .catch(() => {
                    const conn = document.getElementById('alarm-conn');
                    if (conn) conn.textContent = t('error.unreachable');
                    console.log('[SmartDisplay] Poll temporarily failed');
                });
        }

        // --- INITIAL ROUTING: Giriş ekranı veya önceki duruma göre yönlendir ---
        // Eğer kullanıcı daha önce giriş yaptıysa, doğrudan ana sayfaya yönlendir
        if (State.currentUser && State.role) {
            // Giriş bilgilerini sessionStorage'dan yükle
            try {
                sessionStorage.setItem('currentUser', State.currentUser);
                sessionStorage.setItem('role', State.role);
            } catch (e) {}
            // Ana sayfaya yönlendir
            setTimeout(() => {
                if (typeof renderMainLayout === 'function') {
                    renderMainLayout();
                } else if (typeof routeContent === 'function') {
                    routeContent('#/home');
                } else {
                    window.location.hash = '#/home';
                }
            }, 100);
        } else {
            // İlk kez giriş yapıyorsa, doğrudan giriş ekranına yönlendir
            setTimeout(() => {
                ensureIndexLoaded(() => { if (window.renderLogin) window.renderLogin(); });
            }, 100);
        }
    } else {
        const sidebar = document.getElementById('sidebar');
        if (sidebar) {
            sidebar.style.pointerEvents = '';
            sidebar.style.opacity = '';
        }
        window.onhashchange = router;
    }
}

function sendAlarmAction(action) {
    if (alarmActionPending) return;
    alarmActionPending = true;
    // Tüm butonları disable et (A5.2)
    ['btn-arm-home','btn-arm-away','btn-arm-night','btn-disarm'].forEach(id => {
        const btn = document.getElementById(id);
        if (btn) btn.disabled = true;
    });
    const err = document.getElementById('alarm-error');
    if (err) err.textContent = t('ui.waiting');
    // Console log (A5.5)
    console.log('[SmartDisplay] Alarm action requested:', action);
    fetch('http://localhost:8090/api/ui/alarm/action', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action })
    })
    .then(async r => {
        if (r.status === 200) {
            // Başarılı, polling ile bekle
            if (err) err.textContent = t('ui.waiting');
        } else if (r.status === 400) {
            if (err) err.textContent = t('error.invalid');
            alarmActionPending = false;
        } else if (r.status === 409) {
            if (err) err.textContent = t('error.triggered');
            alarmActionPending = false;
        } else if (r.status === 503) {
            if (err) err.textContent = t('error.unreachable');
            alarmActionPending = false;
        } else {
            if (err) err.textContent = t('error.unknown');
            alarmActionPending = false;
        }
    })
    .catch(() => {
        if (err) err.textContent = t('error.network');
        alarmActionPending = false;
    });
}

// Login UI moved to web/index.js; loader above injects it and exposes window.renderLogin
// Home surface & hero styles moved to web/main.css

// --- 6. BAŞLAT ---
window.addEventListener('hashchange', router);
window.addEventListener('DOMContentLoaded', router);
