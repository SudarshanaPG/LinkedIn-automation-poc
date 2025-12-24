package app

const demoHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>LinkedIn POC - Dry Run Demo</title>
    <style>
      :root { color-scheme: light; }
      body { font-family: system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif; margin: 0; background: #f6f7fb; }
      header { padding: 18px 22px; background: white; border-bottom: 1px solid #e7e8ee; position: sticky; top: 0; }
      h1 { font-size: 16px; margin: 0 0 6px 0; }
      p { margin: 0; color: #555; }
      .wrap { padding: 22px; max-width: 920px; margin: 0 auto; }
      .card { background: white; border: 1px solid #e7e8ee; border-radius: 12px; padding: 16px; margin-bottom: 14px; }
      label { display: block; font-weight: 600; margin-bottom: 8px; }
      input, textarea { width: 100%; font-size: 14px; padding: 10px 12px; border: 1px solid #cfd3e1; border-radius: 10px; outline: none; }
      input:focus, textarea:focus { border-color: #4c7dff; box-shadow: 0 0 0 3px rgba(76,125,255,0.15); }
      button { padding: 10px 12px; border: 0; border-radius: 10px; background: #1a73e8; color: white; font-weight: 600; cursor: pointer; }
      button:hover { filter: brightness(0.98); }
      .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
      .pill { display: inline-block; padding: 6px 10px; background: #eef2ff; border: 1px solid #dbe3ff; border-radius: 999px; font-size: 12px; color: #2b47a7; margin-right: 6px; }
      .spacer { height: 1200px; background: linear-gradient(#ffffff00, #ffffff), repeating-linear-gradient(90deg, #f6f7fb, #f6f7fb 20px, #f1f2f8 20px, #f1f2f8 40px); border-radius: 12px; border: 1px dashed #dfe2ef; }
      .hint { font-size: 12px; color: #666; margin-top: 8px; }
    </style>
  </head>
  <body>
    <header>
      <h1>Dry Run Demo Page</h1>
      <p>This page is loaded locally via Rod. In <strong>--dry-run</strong>, the tool does not log into LinkedIn and does not click Connect/Message.</p>
    </header>
    <div class="wrap">
      <div class="card">
        <span class="pill">mouse: bezier + overshoot</span>
        <span class="pill">typing: delays + typos</span>
        <span class="pill">scroll: variable steps</span>
      </div>

      <div class="card">
        <label for="demo-input">Type target</label>
        <textarea id="demo-input" rows="4" placeholder="The automation will type here in dry-run mode."></textarea>
        <div class="hint">Tip: watch the cursor movement, hover pauses, and irregular key timing.</div>
      </div>

      <div class="card grid">
        <div>
          <label>Hover target</label>
          <button id="demo-button" type="button" onclick="document.getElementById('demo-status').textContent='Clicked at '+new Date().toLocaleTimeString()">Click me</button>
          <div class="hint" id="demo-status">Waiting…</div>
        </div>
        <div>
          <label>Random content</label>
          <div class="hint">Scroll down to trigger more human-like scrolling behavior.</div>
        </div>
      </div>

      <div class="card spacer"></div>
    </div>
  </body>
</html>`
