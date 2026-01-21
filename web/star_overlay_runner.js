(function(){
  function info(el){
    if(!el) return null;
    const cs = getComputedStyle(el);
    return { tag: el.tagName, id: el.id||null, cls: el.className||null, display: cs.display, position: cs.position, zIndex: cs.zIndex, bg: cs.background || cs.backgroundColor || null, opacity: cs.opacity, visibility: cs.visibility };
  }
  const points = [
    {x: innerWidth/2, y: innerHeight/2, name:'center'},
    {x: innerWidth*0.1, y: innerHeight/2, name:'left-mid'},
    {x: innerWidth*0.9, y: innerHeight/2, name:'right-mid'},
    {x: innerWidth/2, y: innerHeight*0.1, name:'top-mid'},
    {x: innerWidth/2, y: innerHeight*0.9, name:'bottom-mid'}
  ];

  const report = [];
  const outlines = [];
  points.forEach(p=>{
    const el = document.elementFromPoint(p.x,p.y);
    const top = info(el);
    const ancestors = [];
    let a = el;
    while(a && a !== document.body){ ancestors.push(info(a)); a = a.parentElement; }
    const canvas = document.getElementById('star-bg') || document.getElementById('intro-stars') || document.querySelector('canvas');
    const canvasInfo = info(canvas);
    report.push({point:p.name, coords:[Math.round(p.x),Math.round(p.y)], topElement:top, ancestors:ancestors, starCanvas: canvasInfo});
    // create outline for top element
    if(el && el.getBoundingClientRect){
      const r = el.getBoundingClientRect();
      const o = document.createElement('div');
      o.style.position='fixed'; o.style.left=r.left+'px'; o.style.top=r.top+'px'; o.style.width=r.width+'px'; o.style.height=r.height+'px';
      o.style.border='2px solid rgba(255,80,80,0.95)'; o.style.zIndex=2147483647; o.style.pointerEvents='none'; o.style.boxSizing='border-box';
      o.dataset.__starOverlay='1';
      document.body.appendChild(o); outlines.push(o);
      // label
      const lbl = document.createElement('div'); lbl.textContent = p.name; lbl.style.position='fixed'; lbl.style.left=(r.left+4)+'px'; lbl.style.top=(r.top+4)+'px'; lbl.style.background='rgba(255,80,80,0.95)'; lbl.style.color='#fff'; lbl.style.padding='2px 6px'; lbl.style.fontSize='12px'; lbl.style.zIndex=2147483647; lbl.style.pointerEvents='none'; document.body.appendChild(lbl);
      outlines.push(lbl);
    }
  });

  // append report container
  const pre = document.createElement('pre'); pre.style.position='fixed'; pre.style.right='12px'; pre.style.bottom='12px'; pre.style.maxHeight='40vh'; pre.style.overflow='auto'; pre.style.zIndex=2147483647; pre.style.background='rgba(12,14,18,0.9)'; pre.style.color='#eaf6ff'; pre.style.padding='12px'; pre.style.borderRadius='8px'; pre.style.fontSize='12px'; pre.style.boxShadow='0 6px 24px rgba(0,0,0,0.6)'; pre.dataset.__starOverlay='1';
  pre.textContent = JSON.stringify(report, null, 2);
  document.body.appendChild(pre);

  // cleanup helper
  window.__clearStarOverlay = function(){ document.querySelectorAll('[data__staroverlay], [data="__starOverlay"]').forEach(e=>e.remove()); document.querySelectorAll('[data__staroverlay]').forEach(e=>e.remove()); document.querySelectorAll('[data="__starOverlay"]').forEach(e=>e.remove()); document.querySelectorAll('[data__starOverlay]').forEach(e=>e.remove()); document.querySelectorAll('[data="__starOverlay"]').forEach(e=>e.remove()); document.querySelectorAll('[data__starOverlay]').forEach(e=>e.remove()); /*best-effort*/ };
  console.log('Star overlay runner attached — visual outlines + report appended.');
  console.log('Call window.__clearStarOverlay() to remove overlays.');
})();