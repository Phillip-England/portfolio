// Loads README.md from GitHub and renders it as plain HTML.

(function(){
  const repo = window.__PROJECT_REPO__;
  const container = document.getElementById('readmeContent');
  const status = document.getElementById('readmeStatus');
  if(!repo || !container || !status) return;

  const candidates = [
    `https://raw.githubusercontent.com/phillip-england/${repo}/main/README.md`,
    `https://raw.githubusercontent.com/phillip-england/${repo}/master/README.md`,
  ];

  const escapeHtml = (s) => s
    .replaceAll('&','&amp;')
    .replaceAll('<','&lt;')
    .replaceAll('>','&gt;');

  const parseMarkdown = (text) => {
    let html = text;

    html = html.replace(/^### (.*$)/gim, '<h3>$1</h3>');
    html = html.replace(/^## (.*$)/gim, '<h2>$1</h2>');
    html = html.replace(/^# (.*$)/gim, '<h1>$1</h1>');

    html = html.replace(/\*\*(.*)\*\*/gim, '<strong>$1</strong>');
    html = html.replace(/\*(.*)\*/gim, '<em>$1</em>');
    html = html.replace(/`([^`]+)`/gim, '<code>$1</code>');

    html = html.replace(/```(\w+)?\n([\s\S]*?)```/gim, (match, _lang, code) => {
      const escaped = escapeHtml(code.trim());
      return `<pre><code>${escaped}</code></pre>`;
    });

    html = html.replace(/```(\w+)?([\s\S]*?)```/gim, (match, _lang, code) => {
      const escaped = escapeHtml(code.trim());
      return `<pre><code>${escaped}</code></pre>`;
    });

    html = html.replace(/^\- (.*$)/gim, '<li>$1</li>');
    html = html.replace(/^\d+\. (.*$)/gim, '<li>$1</li>');

    html = html.replace(/\[([^\]]+)\]\(([^\)]+)\)/gim, '<a href="$2" target="_blank">$1</a>');

    html = html.replace(/\n\n/g, '</p><p>');
    html = html.replace(/\n/g, '<br>');

    return `<p>${html}</p>`;
  };

  (async () => {
    for(const url of candidates) {
      try {
        const res = await fetch(url, { cache: 'no-store' });
        if(!res.ok) continue;
        const txt = await res.text();
        container.innerHTML = parseMarkdown(txt);
        status.textContent = 'loaded';

        return;
      } catch (_) {}
    }

    status.textContent = 'failed to load';
    container.innerHTML = '<p class="text-[#555]">Could not load README.md from GitHub (network/CORS/offline).</p>';
  })();
})();
