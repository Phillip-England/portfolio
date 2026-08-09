const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

if (!reducedMotion && 'IntersectionObserver' in window) {
  const observer = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
        observer.unobserve(entry.target);
      }
    }
  }, { threshold: 0.12 });
  document.querySelectorAll('[data-reveal]').forEach((element) => observer.observe(element));
} else {
  document.querySelectorAll('[data-reveal]').forEach((element) => element.classList.add('visible'));
}

const form = document.querySelector('#contact-form');
const email = document.querySelector('#contact-email');
const message = document.querySelector('#contact-message');
const counter = document.querySelector('#char-counter');
const feedback = document.querySelector('#form-feedback');
const submit = document.querySelector('#contact-submit');

message?.addEventListener('input', () => { counter.textContent = `${message.value.length} / 255`; });
form?.addEventListener('submit', async (event) => {
  event.preventDefault();
  feedback.textContent = '';
  submit.disabled = true;
  submit.textContent = 'Sending…';
  try {
    const response = await fetch('/api/contact', { method:'POST', headers:{ 'Content-Type':'application/json' }, body:JSON.stringify({ email:email.value.trim(), message:message.value.trim() }) });
    const result = await response.json();
    feedback.textContent = result.message || (response.ok ? 'Message sent.' : 'Unable to send message.');
    if (response.ok) { form.reset(); counter.textContent = '0 / 255'; }
  } catch {
    feedback.textContent = 'Something went wrong. Please try again.';
  } finally {
    submit.disabled = false;
    submit.textContent = 'Send message';
  }
});
