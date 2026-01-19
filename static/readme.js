// Loads README.md from GitHub and renders it as escaped text inside a code block.
// This keeps things simple and fast, while still showing the freshest README on every visit.

(function(){
  const repo = window.__PROJECT_REPO__;
  const code = document.getElementById('readmeCode');
  const status = document.getElementById('readmeStatus');
  if(!repo || !code || !status) return;

  const candidates = [
    `https://raw.githubusercontent.com/phillip-england/${repo}/main/README.md`,
    `https://raw.githubusercontent.com/phillip-england/${repo}/master/README.md`,
  ];

  const escapeHtml = (s) => s
    .replaceAll('&','&amp;')
    .replaceAll('<','&lt;')
    .replaceAll('>','&gt;');

  (async () => {
    for(const url of candidates) {
      try {
        const res = await fetch(url, { cache: 'no-store' });
        if(!res.ok) continue;
        const txt = await res.text();
        code.innerHTML = escapeHtml(txt);
        status.textContent = 'loaded';
        return;
      } catch (_) {}
    }

    status.textContent = 'failed to load';
    code.textContent = 'Could not load README.md from GitHub (network/CORS/offline).';
  })();
})();
