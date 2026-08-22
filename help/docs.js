
(function applyPreferences(){try{const theme=localStorage.getItem('vigil-help-theme')||'system';if(theme==='light'||theme==='dark')document.documentElement.dataset.theme=theme;const size=localStorage.getItem('vigil-help-font-size')||'comfortable';if(['compact','comfortable','large','extra-large'].includes(size))document.documentElement.dataset.fontSize=size}catch(_){}})();
function ready(fn){if(document.readyState!=='loading')fn();else document.addEventListener('DOMContentLoaded',fn)}
ready(function(){const body=document.body,themeSelect=document.querySelector('[data-theme-select]'),fontSelect=document.querySelector('[data-font-select]'),navOpen=document.querySelector('[data-nav-open]'),navClose=document.querySelector('[data-nav-close]'),search=document.querySelector('[data-command-search]'),clear=document.querySelector('[data-search-clear]'),status=document.querySelector('[data-search-status]');if(themeSelect){themeSelect.value=localStorage.getItem('vigil-help-theme')||'system';themeSelect.addEventListener('change',function(){if(this.value==='system'){delete document.documentElement.dataset.theme;localStorage.removeItem('vigil-help-theme')}else{document.documentElement.dataset.theme=this.value;localStorage.setItem('vigil-help-theme',this.value)}})}if(fontSelect){fontSelect.value=localStorage.getItem('vigil-help-font-size')||'comfortable';fontSelect.addEventListener('change',function(){document.documentElement.dataset.fontSize=this.value;localStorage.setItem('vigil-help-font-size',this.value)})}if(navOpen)navOpen.addEventListener('click',function(){body.classList.add('nav-open');navOpen.setAttribute('aria-expanded','true')});function closeNav(){body.classList.remove('nav-open');if(navOpen)navOpen.setAttribute('aria-expanded','false')}if(navClose)navClose.addEventListener('click',closeNav);document.addEventListener('keydown',function(e){if(e.key==='Escape')closeNav();if(e.key==='/'&&search&&!/input|textarea|select/i.test(document.activeElement.tagName)){e.preventDefault();search.focus()}if((e.ctrlKey||e.metaKey)&&e.key.toLowerCase()==='k'&&search){e.preventDefault();search.focus()}});if(search){const groups=[...document.querySelectorAll('[data-nav-group]')],initialOpen=new Map(groups.map(g=>[g,g.open]));function run(){const q=search.value.trim().toLowerCase();let count=0;body.classList.toggle('searching',!!q);for(const g of groups){let visible=0;for(const item of g.querySelectorAll('li[data-command]')){const ok=!q||item.dataset.command.toLowerCase().includes(q);item.hidden=!ok;if(ok){visible++;count++}}g.hidden=visible===0;if(q&&visible)g.open=true;if(!q)g.open=initialOpen.get(g)}if(status)status.textContent=q?`${count} command${count===1?'':'s'} found`:''}search.addEventListener('input',run);if(clear)clear.addEventListener('click',function(){search.value='';run();search.focus()})}document.querySelectorAll('[data-copy]').forEach(btn=>btn.addEventListener('click',async function(){const text=this.dataset.copy;try{await navigator.clipboard.writeText(text);const old=this.textContent;this.textContent='Copied';setTimeout(()=>this.textContent=old,1200)}catch(_){this.textContent='Copy failed'}}));const active=document.querySelector('.nav-group a.active');if(active)active.scrollIntoView({block:'center'});const tocLinks=[...document.querySelectorAll('[data-toc-link]')];if(tocLinks.length&&'IntersectionObserver'in window){const byId=new Map(tocLinks.map(a=>[a.getAttribute('href').slice(1),a]));const obs=new IntersectionObserver(entries=>{entries.forEach(entry=>{if(entry.isIntersecting){tocLinks.forEach(a=>a.classList.remove('active'));const a=byId.get(entry.target.id);if(a)a.classList.add('active')}})},{rootMargin:'-20% 0px -65% 0px'});document.querySelectorAll('h2[id]').forEach(h=>obs.observe(h))}});


ready(function initImageLightbox(){
  const images = [...document.querySelectorAll('.docs-main img, .doc-article img')]
    .filter(img => !img.closest('a, button') && (img.currentSrc !== '' || img.getAttribute('src')));
  if (!images.length) return;

  let activeTrigger = null;
  const dialog = document.createElement('div');
  dialog.className = 'image-lightbox';
  dialog.hidden = true;
  dialog.setAttribute('role', 'dialog');
  dialog.setAttribute('aria-modal', 'true');
  dialog.setAttribute('aria-label', 'Image preview');
  dialog.innerHTML = '<button type="button" class="image-lightbox-close" aria-label="Close image preview">&times;</button><figure><img alt=""><figcaption></figcaption></figure>';
  document.body.appendChild(dialog);

  const closeButton = dialog.querySelector('.image-lightbox-close');
  const preview = dialog.querySelector('img');
  const caption = dialog.querySelector('figcaption');

  function closeLightbox(){
    dialog.hidden = true;
    document.body.classList.remove('lightbox-open');
    preview.removeAttribute('src');
    if (activeTrigger) activeTrigger.focus();
    activeTrigger = null;
  }

  function openLightbox(img, trigger){
    activeTrigger = trigger;
    preview.src = img.currentSrc || img.src;
    preview.alt = img.alt || 'Preview image';
    const text = img.closest('figure')?.querySelector('figcaption')?.textContent?.trim() || img.alt || '';
    caption.textContent = text;
    caption.hidden = !text;
    dialog.hidden = false;
    document.body.classList.add('lightbox-open');
    closeButton.focus();
  }

  images.forEach(img => {
    const trigger = document.createElement('button');
    trigger.type = 'button';
    trigger.className = 'lightbox-trigger';
    trigger.setAttribute('aria-label', 'Open image preview' + (img.alt ? ': ' + img.alt : ''));
    img.parentNode.insertBefore(trigger, img);
    trigger.appendChild(img);
    trigger.addEventListener('click', () => openLightbox(img, trigger));
  });

  closeButton.addEventListener('click', closeLightbox);
  dialog.addEventListener('click', event => {
    if (event.target === dialog) closeLightbox();
  });
  document.addEventListener('keydown', event => {
    if (event.key === 'Escape' && !dialog.hidden) closeLightbox();
  });
});
