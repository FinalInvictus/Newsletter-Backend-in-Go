const api = '/';
const grid = document.querySelector('#article-grid');
const adminList = document.querySelector('#admin-articles');
const articleForm = document.querySelector('#article-form');
const toast = document.querySelector('#toast');
let articles = [];
let publicArticles = [];
let categories = [];

// Request JSON from the same-origin API and turn non-success responses into readable errors.
async function request(path, options = {}) {
  const response = await fetch(api + path.replace(/^\//, ''), {
    ...options,
    headers: { ...(options.body ? { 'Content-Type': 'application/json' } : {}), ...options.headers },
  });
  if (response.status === 204) return null;
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error || `Request failed (${response.status})`);
  return data;
}

// Escape values before placing API-provided text into HTML templates.
function escapeHTML(value = '') {
  return String(value).replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
}

// Show a short-lived confirmation or error message without interrupting the reader.
function showToast(message) {
  toast.textContent = message;
  toast.classList.add('show');
  window.setTimeout(() => toast.classList.remove('show'), 2800);
}

// Convert the backend's JSON content field into readable text for story previews.
function contentText(contents) {
  if (typeof contents === 'string') {
    try { return contentText(JSON.parse(contents)); } catch { return contents; }
  }
  if (Array.isArray(contents)) return contents.map(contentText).join(' ');
  if (contents && typeof contents === 'object') return Object.values(contents).map(contentText).join(' ');
  return '';
}

// Format timestamps consistently for the publication and editorial list.
function formatDate(value) {
  if (!value) return 'JUST ADDED';
  return new Intl.DateTimeFormat('en', { month: 'short', day: 'numeric', year: 'numeric' }).format(new Date(value)).toUpperCase();
}

// Render published articles as the home page's latest stories.
function renderStories() {
  document.querySelector('#article-count').textContent = articles.length;
  if (!publicArticles.length) {
    grid.innerHTML = '';
    document.querySelector('#empty-state').hidden = false;
    return;
  }
  document.querySelector('#empty-state').hidden = true;
  grid.innerHTML = publicArticles.slice(0, 6).map((article, index) => {
    const summary = article.description || contentText(article.contents);
    const palette = ['#76836d,#c1a27b', '#9d8f79,#647a71', '#687d77,#c1aa80'];
    return `<article class="story-card" data-id="${escapeHTML(article.id)}" tabindex="0" role="button" aria-label="Read ${escapeHTML(article.title)}">
      <div class="story-image">${article.cover_image ? `<img src="${escapeHTML(article.cover_image)}" alt="" loading="lazy" onerror="this.remove()">` : `<div class="story-art" style="background:linear-gradient(135deg,${palette[index % palette.length]})">${escapeHTML((article.title || 'F').slice(0, 1))}</div>`}</div>
      <span class="story-meta">${escapeHTML(article.type || 'STORY')} &nbsp;·&nbsp; ${escapeHTML(formatDate(article.published_at || article.created_at))}</span>
      <h3>${escapeHTML(article.title)}</h3><p>${escapeHTML(summary)}</p>
      <div class="story-footer"><span>${escapeHTML(article.estimated_read_time || 1)} MIN READ</span><span>READ STORY&nbsp; ↗</span></div></article>`;
  }).join('');
}

// Open the selected article in a focused reading dialog.
function openStory(article) {
  const existing = document.querySelector('#story-dialog');
  existing?.remove();
  const dialog = document.createElement('dialog');
  dialog.id = 'story-dialog';
  const body = contentText(article.contents).split(/\n\s*\n/).filter(Boolean).map((paragraph) => `<p>${escapeHTML(paragraph)}</p>`).join('');
  dialog.innerHTML = `<button class="icon-button dialog-close" aria-label="Close story">×</button><span class="story-meta">${escapeHTML(article.type || 'STORY')} &nbsp;·&nbsp; ${escapeHTML(formatDate(article.published_at))}</span><h2>${escapeHTML(article.title)}</h2><p class="dialog-deck">${escapeHTML(article.description || '')}</p><div class="dialog-body">${body || `<p>${escapeHTML(contentText(article.contents))}</p>`}</div><form class="comment-form"><h3>Leave a thought</h3><input name="author_name" placeholder="Your name (optional)"><textarea name="content" placeholder="Add to the conversation…" required></textarea><button class="primary-btn">Post comment →</button></form><div class="comments"></div>`;
  document.body.append(dialog);
  dialog.showModal();
  dialog.querySelector('.dialog-close').addEventListener('click', () => dialog.close());
  dialog.addEventListener('click', (event) => { if (event.target === dialog) dialog.close(); });
  loadComments(article.id, dialog);
  dialog.querySelector('.comment-form').addEventListener('submit', async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const values = new FormData(form);
    try {
      await request(`articles/${article.id}/comments`, { method: 'POST', body: JSON.stringify({ author_name: values.get('author_name') || null, content: values.get('content') }) });
      form.reset();
      await loadComments(article.id, dialog);
      showToast('Your thought has been added.');
    } catch (error) { showToast(error.message); }
  });
}

// Fetch and display comments attached to the selected article.
async function loadComments(articleID, dialog) {
  try {
    const comments = await request(`articles/${articleID}/comments`);
    dialog.querySelector('.comments').innerHTML = comments.map((comment) => `<div class="comment"><strong>${escapeHTML(comment.author_name || 'A reader')}</strong><span>${escapeHTML(formatDate(comment.created_at))}</span><p>${escapeHTML(comment.content)}</p></div>`).join('');
  } catch { dialog.querySelector('.comments').innerHTML = '<p class="form-message">Comments are temporarily unavailable.</p>'; }
}

// Load the public article feed and render a useful failure state if the API is offline.
async function loadStories() {
  try {
    publicArticles = await request('articles');
    renderStories();
  } catch (error) {
    grid.innerHTML = `<p class="loading-card">We couldn't load the stories. ${escapeHTML(error.message)}</p>`;
  }
}

// Render the editorial list with status and edit/delete controls.
function renderAdminArticles() {
  if (!articles.length) {
    adminList.innerHTML = '<div class="loading-card">No stories yet. Start with a new story above.</div>';
    return;
  }
  adminList.innerHTML = articles.map((article) => `<div class="admin-row"><div class="admin-row-main"><strong>${escapeHTML(article.title)}</strong><small><span class="badge ${escapeHTML(article.status)}">${escapeHTML(article.status)}</span>${escapeHTML(article.slug)} &nbsp;·&nbsp; ${escapeHTML(formatDate(article.created_at))}</small></div><div class="row-actions"><button data-action="edit" data-id="${escapeHTML(article.id)}">Edit</button><button data-action="delete" data-id="${escapeHTML(article.id)}">Delete</button></div></div>`).join('');
}

// Retrieve all articles and categories required by the editor.
async function loadEditor() {
  try {
    [articles, categories] = await Promise.all([request('articles/all'), request('categories')]);
    renderAdminArticles();
    renderCategories();
    document.querySelector('#article-form [name="category_id"]').innerHTML = '<option value="">No category</option>' + categories.map((category) => `<option value="${escapeHTML(category.id)}">${escapeHTML(category.name)}</option>`).join('');
  } catch (error) { adminList.innerHTML = `<div class="loading-card">The editor could not load: ${escapeHTML(error.message)}</div>`; }
}

// Refresh the subscriber list and show the current audience size.
async function loadSubscribers() {
  const list = document.querySelector('#subscriber-list');
  try {
    const subscribers = await request('subscribers');
    document.querySelector('#subscriber-count').textContent = `${subscribers.length} readers`;
    list.innerHTML = subscribers.length ? subscribers.map((subscriber) => `<div class="admin-row"><span class="subscriber-email">${escapeHTML(subscriber.email)}</span><span class="subscriber-date">${escapeHTML(formatDate(subscriber.subscribed_at))} ${subscriber.is_verified ? '· VERIFIED' : ''}</span></div>`).join('') : '<div class="loading-card">No readers have joined yet.</div>';
  } catch (error) { list.innerHTML = `<div class="loading-card">Unable to load readers: ${escapeHTML(error.message)}</div>`; }
}

// Display existing categories with their article friendly slugs.
function renderCategories() {
  const list = document.querySelector('#category-list');
  list.innerHTML = categories.length ? categories.map((category) => `<div class="admin-row"><div class="admin-row-main"><strong>${escapeHTML(category.name)}</strong><small>${escapeHTML(category.slug || '')}</small></div></div>`).join('') : '<div class="loading-card">No categories yet. Add one above.</div>';
}

// Reset the story editor and show a clean form for creating a story.
function startNewArticle() {
  articleForm.reset();
  articleForm.elements.id.value = '';
  document.querySelector('#article-form-title').textContent = 'New story';
  document.querySelector('#article-form-message').textContent = '';
  articleForm.hidden = false;
  articleForm.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

// Populate the story editor with a selected article's existing fields.
function editArticle(article) {
  const form = articleForm;
  form.elements.id.value = article.id;
  for (const field of ['title', 'slug', 'description', 'type', 'status', 'estimated_read_time']) {
    if (form.elements[field]) form.elements[field].value = article[field] ?? (field === 'estimated_read_time' ? 3 : '');
  }
  form.elements.contents.value = typeof article.contents === 'string' ? article.contents : JSON.stringify(article.contents, null, 2);
  form.elements.cover_image.value = article.cover_image || '';
  form.elements.category_id.value = article.category_id || '';
  form.elements.tags.value = (article.tags || []).join(', ');
  document.querySelector('#article-form-title').textContent = 'Edit story';
  document.querySelector('#article-form-message').textContent = '';
  form.hidden = false;
  form.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

// Build a safe slug from a title and only replace it when the slug field is still untouched.
function slugify(value) {
  return value.toLowerCase().trim().normalize('NFD').replace(/[\u0300-\u036f]/g, '').replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
}

// Open the active editor panel and load subscribers only when that panel is selected.
function selectPanel(panelName) {
  document.querySelectorAll('.side-tab').forEach((button) => button.classList.toggle('active', button.dataset.panel === panelName));
  document.querySelectorAll('.editor-panel').forEach((panel) => { panel.hidden = panel.id !== `panel-${panelName}`; });
  if (panelName === 'subscribers') loadSubscribers();
}

// Subscribe a reader through the newsletter API and report duplicate or invalid addresses clearly.
document.querySelector('#subscribe-form').addEventListener('submit', async (event) => {
  event.preventDefault();
  const input = event.currentTarget.elements.email;
  const message = document.querySelector('#subscribe-message');
  const button = event.currentTarget.querySelector('button');
  button.disabled = true;
  try {
    await request('subscribe', { method: 'POST', body: JSON.stringify({ email: input.value }) });
    message.textContent = 'You’re on the list. See you Sunday.';
    message.classList.remove('error');
    event.currentTarget.reset();
  } catch (error) { message.textContent = error.message; message.classList.add('error'); }
  finally { button.disabled = false; }
});

// Open an article from mouse or keyboard activation of a story card.
grid.addEventListener('click', (event) => {
  const card = event.target.closest('.story-card');
  if (card) openStory(publicArticles.find((article) => article.id === card.dataset.id));
});
// Support keyboard activation so story cards remain accessible without a mouse.
grid.addEventListener('keydown', (event) => {
  if ((event.key === 'Enter' || event.key === ' ') && event.target.classList.contains('story-card')) { event.preventDefault(); openStory(publicArticles.find((article) => article.id === event.target.dataset.id)); }
});

// Switch between story, category, and subscriber views in the editorial desk.
document.querySelectorAll('.side-tab').forEach((button) => button.addEventListener('click', () => selectPanel(button.dataset.panel)));
// Reveal a blank form when the editor requests a new story.
document.querySelector('#new-article').addEventListener('click', startNewArticle);
// Close the article form without saving its contents.
document.querySelector('#close-form').addEventListener('click', () => { articleForm.hidden = true; });
// Cancel the edit and clear its unsaved fields.
document.querySelector('#cancel-form').addEventListener('click', () => { articleForm.hidden = true; articleForm.reset(); });

// Save a new or updated article, normalizing fields to match the backend's JSON contract.
articleForm.addEventListener('submit', async (event) => {
  event.preventDefault();
  const form = event.currentTarget;
  const values = new FormData(form);
  const message = document.querySelector('#article-form-message');
  let contents;
  const source = values.get('contents').trim();
  try { contents = JSON.parse(source); } catch { contents = { text: source }; }
  const cover = values.get('cover_image').trim();
  const payload = {
    title: values.get('title').trim(), slug: values.get('slug').trim(), description: values.get('description').trim(), contents,
    cover_image: cover || null, tags: values.get('tags').split(',').map((tag) => tag.trim()).filter(Boolean),
    category_id: values.get('category_id') || null, status: values.get('status'), type: values.get('type'), estimated_read_time: 3,
  };
  try {
    const id = values.get('id');
    await request(id ? `articles/${id}` : 'articles', { method: id ? 'PUT' : 'POST', body: JSON.stringify(payload) });
    form.hidden = true;
    form.reset();
    await loadEditor();
    showToast(id ? 'Story updated.' : 'Story saved.');
  } catch (error) { message.textContent = error.message; message.classList.add('error'); }
});

// Keep a new story slug in step with the title until the editor customizes it.
articleForm.elements.title.addEventListener('input', (event) => {
  const slug = articleForm.elements.slug;
  if (!slug.dataset.touched) slug.value = slugify(event.target.value);
});
// Remember when the editor explicitly changes the generated slug.
articleForm.elements.slug.addEventListener('input', () => { articleForm.elements.slug.dataset.touched = 'true'; });
// Clear the custom-slug preference when the form starts over.
articleForm.addEventListener('reset', () => { delete articleForm.elements.slug.dataset.touched; });

// Load articles and categories when the page opens.
loadEditor();
loadStories();
