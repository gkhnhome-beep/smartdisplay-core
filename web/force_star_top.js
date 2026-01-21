(function(){
  function tryFix(){
    const canvas = document.getElementById('star-bg') || document.getElementById('intro-stars') || document.querySelector('canvas');
    if(!canvas){ console.log('force_star_top: no canvas found yet'); return; }
    canvas.style.position = 'fixed';
    canvas.style.left = '0';
    canvas.style.top = '0';
    canvas.style.right = '0';
    canvas.style.bottom = '0';
    canvas.style.zIndex = '2147483646';
    canvas.style.pointerEvents = 'none';
    // move to document.body end so it sits above siblings
    try{ document.body.appendChild(canvas); }catch(e){ console.warn('force_star_top: append failed', e); }
    console.log('force_star_top applied', canvas);
  }
  if(document.readyState === 'complete' || document.readyState === 'interactive') tryFix();
  else document.addEventListener('DOMContentLoaded', tryFix);
  window.__forceStarTop = tryFix;
})();