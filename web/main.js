(function(){
  'use strict';

  // Clean, defensive frontend entry (drop-in candidate for main.js)
  // - single canonical State
  // - small loader helpers (index/settings)
  // - safe router that does not assume functions exist
  // - defensive DOM access and graceful fallbacks

  // ---- State ----
  const State = {
    currentUser: typeof sessionStorage !== 'undefined' ? sessionStorage.getItem('currentUser') || 'Kullanıcı' : 'Kullanıcı',
    role: typeof sessionStorage !== 'undefined' ? sessionStorage.getItem('role') || 'guest' : 'guest',
    clockInterval: null,
    introStarsFrame: null
  };
  // expose for legacy code
  window.State = State;

  // ---- Utilities ----
  function safeGet(id){ return document.getElementById(id) || null; }
  function safeCall(fn, ...args){ if(typeof fn === 'function'){ try{ return fn.apply(null, args); }catch(e){ console.error('safeCall error', e); } } }

  // ---- Loader helpers (synchronous injection to preserve script order) ----
  function injectScriptOnce(src, flagName, onReady){
    if(window[flagName]) return setTimeout(onReady, 0);
    window[flagName] = true;
    const s = document.createElement('script');
    s.src = src;
    s.async = false;
    s.defer = false;
    s.onload = () => setTimeout(onReady, 0);
    s.onerror = () => { console.error('[Loader] failed to load', src); onReady && onReady(); };
    document.head.appendChild(s);
  }

  function ensureIndexLoaded(cb){
    if(window.renderIntro && window.renderLogin) return cb && cb();
    injectScriptOnce('./index.js', '__indexScriptInjected', cb);
  }

  function ensureSettingsLoaded(cb){
    if(window.mountSettingsView && window.renderSettingsSidebar) return cb && cb();
    injectScriptOnce('./settings.js', '__settingsScriptInjected', cb);
  }

  // ---- State restore ----
  function restoreStateFromSession(){
    if(typeof sessionStorage === 'undefined') return;
    try{
      const role = sessionStorage.getItem('role');
      const user = sessionStorage.getItem('currentUser');
      if(role) State.role = role;
      if(user) State.currentUser = user;
    }catch(e){ console.warn('restoreStateFromSession failed', e); }
  }
  restoreStateFromSession();

  // ---- Router (minimal, defensive) ----
  function showIntro(){ ensureIndexLoaded(()=> safeCall(window.renderIntro)); }
  function showLogin(){ ensureIndexLoaded(()=> safeCall(window.renderLogin)); }
  function showSettings(){ ensureSettingsLoaded(()=> safeCall(window.mountSettingsView)); }

  function router(){
    const hash = (window.location.hash||'').replace(/^#/, '');
    const app = safeGet('app');
    if(!app){ console.error("Router: can't find #app"); return; }

    // simple routing table
    switch(hash){
      case '':
      case '/':
      case '/intro':
      case 'intro':
        showIntro();
        return;
      case '/login':
      case 'login':
        showLogin();
        return;
      case '/settings':
      case 'settings':
        // keep layout intact - settings module mounts into #app
        showSettings();
        return;
      default:
        // default: try to render home if available, otherwise show login
        if(typeof window.renderHome === 'function'){ safeCall(window.renderHome); }
        else showLogin();
    }
  }

  // ---- Safe init ----
  function init(){
    // attach router to hash changes
    window.addEventListener('hashchange', router, { passive: true });

    // start router once DOM ready
    if(document.readyState === 'loading'){
      document.addEventListener('DOMContentLoaded', router);
    } else router();

    // gentle safeguard: ensure star canvas startup functions are called if present
    try{
      if(typeof window.setupPremiumStars === 'function') safeCall(window.setupPremiumStars);
      if(typeof window.setupCanvas === 'function') safeCall(window.setupCanvas);
      if(typeof window.animateStars === 'function') safeCall(window.animateStars);
    }catch(e){ console.warn('star startup failed', e); }
  }

  // make init visible for manual calls/tests
  window.appInit = init;

  // auto-run but allow override
  if(!window.__appInitSuppressed) init();

})();