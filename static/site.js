// Tiny site JS: keeps things lightweight.
// - Marks current nav link
// - Adds view transitions for internal navigation
// - Respects reduced-motion preferences

// View transition styles
const vt = document.createElement('style');
vt.textContent = `
  ::view-transition-old(root), ::view-transition-new(root) { animation-duration: 260ms; }
  ::view-transition-old(root) { animation-timing-function: ease-in; }
  ::view-transition-new(root) { animation-timing-function: ease-out; }
  html { scrollbar-gutter: stable; }
`;
document.head.appendChild(vt);

(function () {
  try {
    const path = location.pathname;
    document.querySelectorAll('nav a[href]').forEach((a) => {
      const href = a.getAttribute('href');
      if (!href) return;
      if (href === path || (href !== '/' && path.endsWith(href))) {
        a.classList.add('text-ink', 'font-semibold');
      }
    });
  } catch (_) {}
})();

// View transitions on internal navigation.
(function(){
  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  if (!document.startViewTransition || reducedMotion) return;

  document.addEventListener('click', (e) => {
    const a = e.target.closest('a');
    if (!a) return;
    if (a.hasAttribute('download') || a.target) return;
    if (a.origin !== window.location.origin) return;
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;

    e.preventDefault();
    document.startViewTransition(() => {
      window.location.href = a.href;
    });
  });
})();

// Mobile nav toggle
(function(){
  const menu = document.querySelector('[data-mobile-menu]');
  const openBtn = document.querySelector('[data-mobile-menu-open]');
  if (!menu || !openBtn) return;

  const closeTargets = menu.querySelectorAll('[data-mobile-menu-close]');
  const open = () => {
    menu.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
  };
  const close = () => {
    menu.classList.add('hidden');
    document.body.style.overflow = '';
  };

  openBtn.addEventListener('click', open);
  closeTargets.forEach((el) => el.addEventListener('click', close));
  menu.addEventListener('click', (e) => {
    const link = e.target.closest('a[href]');
    if (link) close();
  });
})();

// Home-only ASCII halo
(function(){
  if(document.getElementById('asciiHalo')){
    const s = document.createElement('script');
    s.src = '/static/ascii.js';
    s.defer = true;
    document.head.appendChild(s);
  }
})();
