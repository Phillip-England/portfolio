// Contact form behavior.
// TODO: Decide where it posts. Then replace the "mock submit" with a fetch() call.

(function(){
  const form = document.getElementById('contactForm');
  const email = document.getElementById('email');
  const message = document.getElementById('message');
  const count = document.getElementById('count');
  const status = document.getElementById('status');
  if(!form || !email || !message || !count || !status) return;

  const updateCount = () => {
    count.textContent = String(message.value.length);
  };
  message.addEventListener('input', updateCount);
  updateCount();

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    status.textContent = '';

    const payload = {
      email: email.value.trim(),
      message: message.value.trim(),
      createdAt: new Date().toISOString(),
    };

    // Basic validation
    if(!payload.email || !payload.email.includes('@')) {
      status.textContent = 'Please enter a valid email.';
      return;
    }
    if(!payload.message) {
      status.textContent = 'Please enter a message.';
      return;
    }
    if(payload.message.length > 255) {
      status.textContent = 'Message must be 255 characters or less.';
      return;
    }

    try {
      const res = await fetch('/contact', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (res.ok) {
        status.textContent = 'Message sent successfully!';
        form.reset();
        updateCount();
      } else if (res.status === 429) {
        status.textContent = 'Daily message limit reached. Please try again tomorrow.';
      } else {
        status.textContent = 'Failed to send. Please try again.';
      }
    } catch (_) {
      status.textContent = 'Connection error. Please try again.';
    }
  });
})();
