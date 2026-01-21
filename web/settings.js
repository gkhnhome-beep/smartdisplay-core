// settings.js — extracted settings view functions from main.js
// Exposes: window.renderSettingsSidebar, window.settingsSidebarClickHandler,
// window.openSettingsItem, window.mountSettingsView

(function(){
    window.renderSettingsSidebar = function() {
        const sidebar = document.getElementById('settingsSidebar');
        if (!sidebar) return;
        let html = '';
        if (!window.settingsSidebarState) window.settingsSidebarState = { expanded: null };
        SETTINGS_MENU.forEach((section, idx) => {
            const expanded = window.settingsSidebarState.expanded === section.section;
            html += `<div class="settings-section-title" data-section="${section.section}" tabindex="0" style="cursor:pointer;user-select:none;display:flex;align-items:center;justify-content:space-between;">
                    <span>${section.section}</span>
                    <span style="font-size:1.1em; color:var(--muted);">${expanded ? '▾' : '▸'}</span>
                </div>`;
            html += `<div class="settings-submenu" data-section="${section.section}" style="display:${expanded ? 'block' : 'none'};">`;
            section.items.forEach(item => {
                html += `<div class="settings-item" tabindex="0" data-section="${section.section}" data-item="${item}">${item}</div>`;
            });
            html += `</div>`;
        });
        sidebar.innerHTML = html;
    };

    window.settingsSidebarClickHandler = function(e) {
        const secTitle = e.target.closest('.settings-section-title');
        if (secTitle) {
            const sec = secTitle.getAttribute('data-section');
            if (window.settingsSidebarState.expanded === sec) {
                window.settingsSidebarState.expanded = null;
            } else {
                window.settingsSidebarState.expanded = sec;
            }
            window.renderSettingsSidebar();
            return;
        }
        const item = e.target.closest('.settings-item');
        if (item) {
            window.openSettingsItem(item.getAttribute('data-section'), item.getAttribute('data-item'), item);
        }
    };

    window.openSettingsItem = function(section, item, el) {
        document.querySelectorAll('.settings-item').forEach(e => e.classList.remove('active'));
        if (el) el.classList.add('active');
        let content = document.getElementById('settingsContent');
        if (!content) return;
        const key = section + '::' + item;
        const page = SETTINGS_PAGES[key];
        if (page) {
            let header = `<div class="settings-page-header">
                    <div class="settings-page-title">${page.title || item}</div>
                    ${page.description ? `<div class="settings-page-desc">${page.description}</div>` : ''}
                    <hr class="settings-divider" />
                </div>`;
            if (page.type === 'custom' && typeof page.render === 'function') {
                content.innerHTML = header;
                page.render(content);
            } else if (typeof page.render === 'function') {
                content.innerHTML = header;
                page.render(content);
            } else {
                content.innerHTML = header + `<div class='settings-placeholder-card'><span class='icon'>⚙️</span>Bu özellik bu sürümde mevcut değil veya yakında eklenecek.</div>`;
            }
        } else {
            content.innerHTML = `<div class="settings-page-header"><div class="settings-page-title">${item}</div><hr class="settings-divider" /></div><div class='settings-placeholder-card'><span class='icon'>⚙️</span>Bu özellik bu sürümde mevcut değil veya yakında eklenecek.</div>`;
        }
    };

    window.mountSettingsView = function(rootEl) {
        if (window.SmartDisplay && window.SmartDisplay.debug) {
            if (!window.__settingsMountLogged) {
                console.log('[DEV] Settings mount');
                window.__settingsMountLogged = true;
            }
        }
        rootEl.innerHTML = `
            <div class="settings-layout" id="settingsLayoutRoot">
                <aside class="settings-sidebar" id="settingsSidebar"></aside>
                <section class="settings-content" id="settingsContent"></section>
            </div>
        `;
        // Ensure root is above background canvas
        try { rootEl.style.position = rootEl.style.position || 'relative'; rootEl.style.zIndex = '2'; } catch(e) {}
        // Render sidebar items
        window.renderSettingsSidebar();
        const sidebar = document.getElementById('settingsSidebar');
        if (sidebar) {
            // Attach handler
            sidebar.addEventListener('click', window.settingsSidebarClickHandler);
            // Apply robust inline fallback styles to guard against missing/overridden CSS
            sidebar.style.width = sidebar.style.width || '290px';
            sidebar.style.background = sidebar.style.background || '#1a1d22';
            sidebar.style.color = sidebar.style.color || '#eaf6ff';
            sidebar.style.padding = sidebar.style.padding || '24px 0 0 0';
            sidebar.style.boxSizing = 'border-box';
        }
        const content = document.getElementById('settingsContent');
        if (content) {
            content.style.flex = '1';
            content.style.padding = content.style.padding || '40px 48px';
            content.style.background = content.style.background || 'transparent';
            content.style.color = content.style.color || '#eaf6ff';
            content.style.boxSizing = 'border-box';
        }
        // Make settings items readable even if CSS variables are missing
        document.querySelectorAll('.settings-item').forEach(it => {
            if (!it.style.color) it.style.color = '#bcd7ef';
            it.style.cursor = 'pointer';
        });
        // Ensure the settings layout container itself sits above the animated background
        const layoutRoot = document.getElementById('settingsLayoutRoot');
        if (layoutRoot) {
            layoutRoot.style.position = layoutRoot.style.position || 'relative';
            layoutRoot.style.zIndex = '3';
            layoutRoot.style.minHeight = layoutRoot.style.minHeight || '100vh';
        }
    };

})();

// end of settings.js
