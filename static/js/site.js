const revealElements = [...document.querySelectorAll('[data-reveal]')];
const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

const navToggle = document.querySelector('.nav-toggle');
const primaryNavigation = document.querySelector('#primary-navigation');

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
