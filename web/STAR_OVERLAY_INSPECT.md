(function inspectStarsOverlay(){
  function info(el){
    if(!el) return null;
    const cs = getComputedStyle(el);
    return { tag: el.tagName, id: el.id||null, cls: el.className||null, display: cs.display, position: cs.position, zIndex: cs.zIndex, bg: cs.backgroundColor||cs.background, opacity: cs.opacity, visibility: cs.visibility };
  }
  const points = [
    {x: innerWidth/2, y: innerHeight/2, name:'center'},
    {x: innerWidth*0.1, y: innerHeight/2, name:'left-mid'},
    {x: innerWidth*0.9, y: innerHeight/2, name:'right-mid'},
    {x: innerWidth/2, y: innerHeight*0.1, name:'top-mid'},
    {x: innerWidth/2, y: innerHeight*0.9, name:'bottom-mid'}
  ];
  console.group('Star-overlay-inspect');
  points.forEach(p => {
    const el = document.elementFromPoint(p.x, p.y);
    console.group(p.name + ' ('+Math.round(p.x)+','+Math.round(p.y)+')');
    console.log('topElement:', el);
    console.log('topElement info:', info(el));
    // climb ancestors to body
    let a = el;
    while(a && a !== document.body){
      console.log('ancestor:', info(a));
      a = a.parentElement;
    }
    // show star canvas info
    const canvas = document.getElementById('star-bg') || document.getElementById('intro-stars') || document.querySelector('canvas');
    console.log('star canvas:', canvas, info(canvas));
    console.groupEnd();
  });
  console.groupEnd();
})();

---
Quick loader: to run the visual inspector from this repo, paste this single line into the DevTools Console and press Enter:

```js
var s=document.createElement('script'); s.src='./star_overlay_runner.js'; document.head.appendChild(s);
```

After running, a small report box and red outlines will appear on the page. When finished, run `window.__clearStarOverlay()` in the console to remove them.