package server

import "net/http"

const uiHTML = `<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Muster — Stockyard</title>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:wght@400;700&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rust-light:#e8753a;--rust-dark:#8b3d1a;--leather:#a0845c;--leather-light:#c4a87a;--cream:#f0e6d3;--cream-dim:#bfb5a3;--cream-muted:#7a7060;--gold:#d4a843;--green:#5ba86e;--red:#c0392b;--font-serif:'Libre Baskerville',Georgia,serif;--font-mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--font-serif);min-height:100vh}a{color:var(--rust-light);text-decoration:none}
.hdr{background:var(--bg2);border-bottom:2px solid var(--rust-dark);padding:.9rem 1.8rem;display:flex;align-items:center;justify-content:space-between}.hdr-left{display:flex;align-items:center;gap:1rem}.hdr-brand{font-family:var(--font-mono);font-size:.75rem;color:var(--leather);letter-spacing:3px;text-transform:uppercase}.hdr-title{font-family:var(--font-mono);font-size:1.1rem;color:var(--cream)}.badge{font-family:var(--font-mono);font-size:.6rem;padding:.2rem .6rem;letter-spacing:1px;text-transform:uppercase;border:1px solid;color:var(--green);border-color:var(--green)}
.main{max-width:1000px;margin:0 auto;padding:2rem 1.5rem}.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:1rem;margin-bottom:2rem}.card{background:var(--bg2);border:1px solid var(--bg3);padding:1rem 1.2rem}.card-val{font-family:var(--font-mono);font-size:1.6rem;font-weight:700;color:var(--cream);display:block}.card-lbl{font-family:var(--font-mono);font-size:.58rem;letter-spacing:2px;text-transform:uppercase;color:var(--leather);margin-top:.2rem}
.section{margin-bottom:2rem}.section-title{font-family:var(--font-mono);font-size:.68rem;letter-spacing:3px;text-transform:uppercase;color:var(--rust-light);margin-bottom:.8rem;padding-bottom:.5rem;border-bottom:1px solid var(--bg3)}table{width:100%;border-collapse:collapse;font-family:var(--font-mono);font-size:.75rem}th{background:var(--bg3);padding:.4rem .8rem;text-align:left;color:var(--leather-light);font-weight:400;font-size:.62rem;letter-spacing:1px;text-transform:uppercase}td{padding:.4rem .8rem;border-bottom:1px solid var(--bg3);color:var(--cream-dim)}tr:hover td{background:var(--bg2)}.empty{color:var(--cream-muted);text-align:center;padding:2rem;font-style:italic}
.pill{display:inline-block;font-family:var(--font-mono);font-size:.55rem;padding:.1rem .4rem;border-radius:2px;text-transform:uppercase}.pill-open{background:#2a1a1a;color:var(--red)}.pill-resolved{background:#1a3a2a;color:var(--green)}.pill-high{background:#2a1a1a;color:var(--red)}.pill-medium{background:#2a2a1a;color:var(--gold)}.pill-low{background:var(--bg3);color:var(--cream-muted)}
.oncall-banner{background:var(--bg2);border:2px solid var(--green);padding:1.2rem 1.5rem;margin-bottom:2rem;display:flex;justify-content:space-between;align-items:center;font-family:var(--font-mono)}.oncall-name{font-size:1.1rem;color:var(--cream)}.oncall-label{font-size:.6rem;letter-spacing:2px;text-transform:uppercase;color:var(--green)}
</style></head><body>
<div class="hdr"><div class="hdr-left">
<svg viewBox="0 0 64 64" width="22" height="22" fill="none"><rect x="8" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="28" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="48" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="8" y="27" width="48" height="7" rx="2.5" fill="#c4a87a"/></svg>
<span class="hdr-brand">Stockyard</span><span class="hdr-title">Muster</span></div>
<div style="display:flex;gap:.8rem;align-items:center"><span class="badge">Free</span></div></div>
<div class="main">
<div id="oncall-banner" class="oncall-banner" style="display:none">
  <div><div class="oncall-label">Currently On Call</div><div class="oncall-name" id="oncall-name">—</div></div>
  <div style="font-size:.7rem;color:var(--cream-muted)" id="oncall-dates"></div>
</div>
<div class="cards">
  <div class="card"><span class="card-val" id="s-members">—</span><span class="card-lbl">Members</span></div>
  <div class="card"><span class="card-val" id="s-open">—</span><span class="card-lbl">Open</span></div>
  <div class="card"><span class="card-val" id="s-resolved">—</span><span class="card-lbl">Resolved</span></div>
</div>
<div class="section"><div class="section-title">Open Incidents</div>
<table><thead><tr><th>Title</th><th>Severity</th><th>Status</th><th>Started</th></tr></thead>
<tbody id="inc-body"></tbody></table></div>
<div class="section"><div class="section-title">Schedule</div>
<table><thead><tr><th>Member</th><th>Start</th><th>End</th><th>Notes</th></tr></thead>
<tbody id="sched-body"></tbody></table></div>
</div>
<script>
async function refresh(){
  try{const s=await(await fetch('/api/status')).json();document.getElementById('s-members').textContent=s.members||0;document.getElementById('s-open').textContent=s.open_incidents||0;document.getElementById('s-resolved').textContent=s.resolved_incidents||0;}catch(e){}
  try{const o=await(await fetch('/api/oncall')).json();if(o.on_call){document.getElementById('oncall-banner').style.display='flex';document.getElementById('oncall-name').textContent=o.on_call.member_name;document.getElementById('oncall-dates').textContent=o.on_call.start_date+' → '+o.on_call.end_date;}else{document.getElementById('oncall-banner').style.display='none';}}catch(e){}
  try{const d=await(await fetch('/api/incidents?status=open')).json();const is=d.incidents||[];const tb=document.getElementById('inc-body');
  if(!is.length){tb.innerHTML='<tr><td colspan="4" class="empty">No open incidents</td></tr>';}
  else{tb.innerHTML=is.map(i=>'<tr><td style="color:var(--cream)">'+esc(i.title)+'</td><td><span class="pill pill-'+i.severity+'">'+i.severity+'</span></td><td><span class="pill pill-'+i.status+'">'+i.status+'</span></td><td style="font-size:.65rem;color:var(--cream-muted)">'+timeAgo(i.started_at)+'</td></tr>').join('');}}catch(e){}
  try{const d=await(await fetch('/api/schedules')).json();const ss=d.schedules||[];const tb=document.getElementById('sched-body');
  if(!ss.length){tb.innerHTML='<tr><td colspan="4" class="empty">No schedules</td></tr>';}
  else{tb.innerHTML=ss.map(s=>'<tr><td style="color:var(--cream)">'+esc(s.member_name)+'</td><td>'+s.start_date+'</td><td>'+s.end_date+'</td><td style="font-size:.65rem">'+esc(s.notes)+'</td></tr>').join('');}}catch(e){}
}
function esc(s){const d=document.createElement('div');d.textContent=s||'';return d.innerHTML;}
function timeAgo(s){if(!s)return'—';const d=new Date(s);const diff=Date.now()-d.getTime();if(diff<60000)return'now';if(diff<3600000)return Math.floor(diff/60000)+'m';if(diff<86400000)return Math.floor(diff/3600000)+'h';return Math.floor(diff/86400000)+'d';}
refresh();setInterval(refresh,8000);
</script></body></html>`

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(uiHTML))
}
