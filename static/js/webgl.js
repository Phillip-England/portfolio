/**
 * Your WebGL entry point.
 *
 * No Three.js, framework, or build step is involved. The page loads this file as
 * an ES module. Start in render(), replace the shaders, or split helpers into
 * neighboring modules and import them here.
 */

const canvas = document.querySelector('#webgl-canvas');
const gl = canvas?.getContext('webgl2', {
  alpha: true,
  antialias: true,
  depth: true,
  powerPreference: 'high-performance',
});

if (!gl) {
  canvas?.remove();
} else {
  const state = {
    time: 0,
    delta: 0,
    scroll: 0,
    pointer: { x: 0, y: 0 },
    viewport: { width: 0, height: 0, dpr: 1 },
  };

  // Lifecycle hook: allocate shaders, buffers, textures, and programs here.
  function setup() {
    gl.clearColor(0, 0, 0, 0);
  }

  // Lifecycle hook: update canvas size, viewport, projection matrices, or FBOs.
  function resize() {
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    const width = Math.max(1, Math.floor(window.innerWidth * dpr));
    const height = Math.max(1, Math.floor(window.innerHeight * dpr));

    if (canvas.width !== width || canvas.height !== height) {
      canvas.width = width;
      canvas.height = height;
    }

    state.viewport = { width, height, dpr };
    gl.viewport(0, 0, width, height);
  }

  // Your render logic starts here. This intentionally only clears the canvas.
  function render() {
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

    // Example inputs available to your scene:
    // state.time                 seconds since start
    // state.delta                seconds since last frame
    // state.scroll               normalized document scroll (0..1)
    // state.pointer.x/.y         normalized pointer position (-1..1)
    // state.viewport             physical width, height, and DPR
  }

  function updatePageInputs(event) {
    if (event?.type === 'pointermove') {
      state.pointer.x = (event.clientX / window.innerWidth) * 2 - 1;
      state.pointer.y = 1 - (event.clientY / window.innerHeight) * 2;
    }
    const range = document.documentElement.scrollHeight - window.innerHeight;
    state.scroll = range > 0 ? Math.min(1, Math.max(0, window.scrollY / range)) : 0;
  }

  let previous = performance.now();
  function frame(now) {
    state.delta = Math.min((now - previous) / 1000, 0.1);
    state.time = now / 1000;
    previous = now;
    render();
    requestAnimationFrame(frame);
  }

  setup();
  resize();
  updatePageInputs();
  window.addEventListener('resize', resize, { passive: true });
  window.addEventListener('scroll', updatePageInputs, { passive: true });
  window.addEventListener('pointermove', updatePageInputs, { passive: true });
  requestAnimationFrame(frame);
}
