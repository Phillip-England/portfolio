const revealElements = [...document.querySelectorAll('[data-reveal]')];
const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

const navToggle = document.querySelector('.nav-toggle');
const primaryNavigation = document.querySelector('#primary-navigation');

const copyIcon = `
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
    <rect width="14" height="14" x="8" y="8" rx="2"></rect>
    <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"></path>
  </svg>
`;
const checkIcon = `
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
    <path d="M20 6 9 17l-5-5"></path>
  </svg>
`;

const copyText = async (text) => {
  if (navigator.clipboard?.writeText && window.isSecureContext) {
    await navigator.clipboard.writeText(text);
    return;
  }

  const textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.setAttribute('readonly', '');
  textarea.style.position = 'fixed';
  textarea.style.top = '-9999px';
  document.body.append(textarea);
  textarea.select();

  try {
    document.execCommand('copy');
  } finally {
    textarea.remove();
  }
};

const setCopyState = (button, options = {}) => {
  const {
    copiedLabel = 'Copied',
    defaultLabel = 'Copy',
    copiedHTML = checkIcon,
    defaultHTML = copyIcon,
    status,
    statusText = '',
    resetStatusText = '',
    resetDelay = 1600,
  } = options;

  button.classList.add('is-copied');
  button.setAttribute('aria-label', copiedLabel);
  button.title = copiedLabel;
  button.innerHTML = copiedHTML;
  if (status) status.textContent = statusText;

  clearTimeout(button.copyResetTimer);
  button.copyResetTimer = setTimeout(() => {
    button.classList.remove('is-copied');
    button.setAttribute('aria-label', defaultLabel);
    button.title = defaultLabel;
    button.innerHTML = defaultHTML;
    if (status) status.textContent = resetStatusText;
  }, resetDelay);
};

const markdownButton = document.querySelector('[data-copy-markdown]');
const markdownSource = document.querySelector('#markdown-source');
const markdownStatus = document.querySelector('#copy-markdown-status');

markdownButton?.addEventListener('click', async () => {
  try {
    await copyText(markdownSource?.value ?? '');
    setCopyState(markdownButton, {
      copiedLabel: 'Markdown copied',
      defaultLabel: 'Copy Markdown',
      copiedHTML: `${checkIcon}<span>Markdown copied</span>`,
      defaultHTML: `${copyIcon}<span>Copy Markdown</span>`,
      status: markdownStatus,
      statusText: 'Markdown copied',
      resetDelay: 2200,
    });
  } catch {
    markdownButton.setAttribute('aria-label', 'Unable to copy Markdown');
    markdownButton.title = 'Unable to copy Markdown';
    if (markdownStatus) markdownStatus.textContent = 'Unable to copy Markdown';
  }
});

document.querySelectorAll('.content pre').forEach((pre) => {
  if (pre.closest('.code-copy-wrapper')) return;

  const wrapper = document.createElement('div');
  wrapper.className = 'code-copy-wrapper';
  pre.before(wrapper);
  wrapper.append(pre);

  const button = document.createElement('button');
  button.className = 'code-copy-button';
  button.type = 'button';
  button.setAttribute('aria-label', 'Copy code');
  button.title = 'Copy code';
  button.innerHTML = copyIcon;
  wrapper.append(button);

  button.addEventListener('click', async () => {
    const code = pre.querySelector('code')?.innerText ?? pre.innerText;

    try {
      await copyText(code.replace(/\n$/, ''));
      setCopyState(button, {
        copiedLabel: 'Copied',
        defaultLabel: 'Copy code',
      });
    } catch {
      button.setAttribute('aria-label', 'Unable to copy');
      button.title = 'Unable to copy';
    }
  });
});

const closeNavigation = () => {
  if (!navToggle || !primaryNavigation) return;
  navToggle.setAttribute('aria-expanded', 'false');
  navToggle.setAttribute('aria-label', 'Open navigation menu');
  primaryNavigation.classList.remove('is-open');
};

navToggle?.addEventListener('click', () => {
  const isOpen = navToggle.getAttribute('aria-expanded') === 'true';
  navToggle.setAttribute('aria-expanded', String(!isOpen));
  navToggle.setAttribute('aria-label', isOpen ? 'Open navigation menu' : 'Close navigation menu');
  primaryNavigation?.classList.toggle('is-open', !isOpen);
});

primaryNavigation?.addEventListener('click', (event) => {
  if (event.target.closest('a')) closeNavigation();
});

document.addEventListener('click', (event) => {
  if (!event.target.closest('.nav')) closeNavigation();
});

document.addEventListener('keydown', (event) => {
  if (event.key !== 'Escape' || navToggle?.getAttribute('aria-expanded') !== 'true') return;
  closeNavigation();
  navToggle.focus();
});

window.addEventListener('resize', () => {
  if (window.innerWidth > 620) closeNavigation();
});

if (!reducedMotion && 'IntersectionObserver' in window) {
  const reveal = (element) => element.classList.add('visible');
  const isInitiallyVisible = (element) => {
    const rect = element.getBoundingClientRect();
    return rect.top < window.innerHeight * 0.92;
  };

  const revealObserver = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (!entry.isIntersecting) return;
      reveal(entry.target);
      revealObserver.unobserve(entry.target);
    });
  }, { rootMargin: '0px 0px -12% 0px', threshold: 0.16 });

  requestAnimationFrame(() => {
    revealElements.forEach((element) => {
      if (isInitiallyVisible(element)) {
        reveal(element);
      } else {
        revealObserver.observe(element);
      }
    });
  });
} else {
  revealElements.forEach((element) => element.classList.add('visible'));
}

const navLinks = [...document.querySelectorAll('.nav-links a[href^="#"]')];
const sections = navLinks
  .map((link) => document.querySelector(link.getAttribute('href')))
  .filter(Boolean);

if ('IntersectionObserver' in window && sections.length > 0) {
  const navObserver = new IntersectionObserver((entries) => {
    const visible = entries
      .filter((entry) => entry.isIntersecting)
      .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];

    if (!visible) return;

    navLinks.forEach((link) => {
      const active = link.getAttribute('href') === `#${visible.target.id}`;
      if (active) {
        link.setAttribute('aria-current', 'true');
      } else {
        link.removeAttribute('aria-current');
      }
    });
  }, { rootMargin: '-20% 0px -65% 0px', threshold: [0.1, 0.4, 0.7] });

  sections.forEach((section) => navObserver.observe(section));
}

const form = document.querySelector('#contact-form');
const name = document.querySelector('#contact-name');
const email = document.querySelector('#contact-email');
const message = document.querySelector('#contact-message');
const counter = document.querySelector('#char-counter');
const feedback = document.querySelector('#form-feedback');
const submit = document.querySelector('#contact-submit');

message?.addEventListener('input', () => { counter.textContent = `${message.value.length} / 1200`; });
form?.addEventListener('submit', async (event) => {
  event.preventDefault();
  feedback.textContent = '';
  submit.disabled = true;
  submit.textContent = 'Sending...';
  try {
    const response = await fetch('/api/contact', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: name.value.trim(),
        email: email.value.trim(),
        message: message.value.trim(),
      }),
    });
    const result = await response.json();
    feedback.textContent = result.message || (response.ok ? 'Message sent.' : 'Unable to send message.');
    if (response.ok) { form.reset(); counter.textContent = '0 / 1200'; }
  } catch {
    feedback.textContent = 'Something went wrong. Please try again.';
  } finally {
    submit.disabled = false;
    submit.textContent = 'Send message';
  }
});
