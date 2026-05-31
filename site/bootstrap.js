'use strict';

(async function () {
  const go = new Go();

  let wasmSrc = 'main.wasm';
  // Allow overriding wasm path via query string for testing
  const params = new URLSearchParams(location.search);
  if (params.has('wasm')) wasmSrc = params.get('wasm');

  try {
    const result = await WebAssembly.instantiateStreaming(fetch(wasmSrc), go.importObject);
    go.run(result.instance);
    // Poll until Go has registered its globals
    await waitFor(() => typeof sdvedit_load === 'function', 5000);
    window.__sdvedit_ready = true;
    document.dispatchEvent(new CustomEvent('sdvedit:ready'));
  } catch (err) {
    document.getElementById('load-error').textContent =
      'Failed to load WASM: ' + err.message;
    document.getElementById('load-error').hidden = false;
  }
})();

function waitFor(fn, timeout) {
  return new Promise((resolve, reject) => {
    const start = Date.now();
    const check = () => {
      if (fn()) return resolve();
      if (Date.now() - start > timeout) return reject(new Error('timeout'));
      requestAnimationFrame(check);
    };
    check();
  });
}
