"use strict";

// This frontend is served by the SAME Go/Gin server as /api/v1.
const API_BASE = "/api/v1";
const TOKEN_KEY = "go_api_token";
const $ = (selector) => document.querySelector(selector);
const state = { token: sessionStorage.getItem(TOKEN_KEY) || "", profile: null, articlePage: 1, userPage: 1 };

const escapeHTML = (value) => String(value ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]);
const idNumber = (id) => Number.isSafeInteger(Number(id)) && Number(id) > 0 ? Number(id) : 0;
const canEditArticles = () => ["Admin", "Editor"].includes(state.profile?.role);
const isAdmin = () => state.profile?.role === "Admin";
const roleChip = (role) => `<span class="chip chip-${["Admin", "Editor", "Member"].includes(role) ? role.toLowerCase() : "member"}">${escapeHTML(role || "Member")}</span>`;

function alertMessage(message, error = false) {
  const notice = $("#notice");
  notice.textContent = message;
  notice.className = error ? "error" : "";
  notice.scrollIntoView({ behavior: "smooth", block: "nearest" });
}
function showError(error) { alertMessage(error instanceof Error ? error.message : String(error), true); }
function run(task) { return Promise.resolve().then(task).catch(showError); }
function empty(title, description = "") {
  return `<div class="empty-state"><div class="empty-symbol">▤</div><strong>${escapeHTML(title)}</strong>${escapeHTML(description)}</div>`;
}
function switchTab(tab) {
  document.querySelectorAll(".panel").forEach(el => { el.hidden = el.id !== `panel-${tab}`; });
  document.querySelectorAll("[data-tab]").forEach(el => el.classList.toggle("selected", el.dataset.tab === tab));
  $("#current-page").textContent = ({ articles: "Articles", categories: "Categories", account: "Account", dashboard: "Dashboard", users: "Users", tester: "API tester" })[tab] || "Workspace";
  $("#notice").textContent = "";
  window.scrollTo({ top: 0, behavior: "smooth" });
}
function updateSession() {
  const who = state.profile ? `${state.profile.name || state.profile.email} · ${state.profile.role}` : (state.token ? "Signed in" : "Browsing as guest");
  $("#session-status").textContent = who;
  $("#sign-out").hidden = !state.token;
  $("#go-sign-in").hidden = Boolean(state.token);
  $("#article-new").hidden = !canEditArticles();
  $("#category-new").hidden = !isAdmin();
  $("#user-new").hidden = !isAdmin();
  $("#account-auth-forms").hidden = Boolean(state.token);
  $("#account-profile-area").hidden = !state.token;
}
function setToken(token) {
  state.token = token;
  state.profile = null;
  if (token) sessionStorage.setItem(TOKEN_KEY, token);
  else sessionStorage.removeItem(TOKEN_KEY);
  updateSession();
}
async function api(path, options = {}) {
  if (!path.startsWith("/") || path.startsWith("//")) throw new Error("The API path must begin with /.");
  const headers = new Headers(options.headers || {});
  if (state.token) headers.set("Authorization", `Bearer ${state.token}`);
  const init = { method: options.method || "GET", headers };
  if (options.body instanceof FormData) init.body = options.body;
  else if (options.body !== undefined) {
    headers.set("Content-Type", "application/json");
    init.body = JSON.stringify(options.body);
  }
  let response;
  try { response = await fetch(`${API_BASE}${path}`, init); }
  catch { throw new Error("Could not connect to the Go API. Open this UI through Gin (for example http://localhost:5001/ui/), not the IDE preview server."); }
  const type = response.headers.get("content-type") || "";
  const result = response.status === 204 ? null : type.includes("json") ? await response.json() : await response.text();
  if (!response.ok) {
    const msg = result?.error || result?.message || (typeof result === "string" ? result.slice(0, 120) : "Request failed");
    if (response.status === 401 && state.token) setToken("");
    throw new Error(`${response.status}: ${msg}${response.status === 403 ? " (Your account does not have permission.)" : ""}`);
  }
  return result;
}
function params(values) {
  const query = new URLSearchParams();
  for (const [key, val] of Object.entries(values)) if (val !== "" && val != null) query.set(key, String(val));
  return query.size ? `?${query}` : "";
}
function readForm(form) { return Object.fromEntries(new FormData(form).entries()); }
function jsonObject(fields, keys) { return Object.fromEntries(keys.filter((key) => fields[key] !== "").map(key => [key, fields[key]])); }
function collectMultipart(form, keys) {
  const data = new FormData();
  for (const key of keys) {
    const control = form.elements.namedItem(key);
    if (!control) continue;
    if (control.type === "file") { if (control.files.length) data.append(key, control.files[0]); }
    else if (control.value !== "") data.append(key, control.value);
  }
  return data;
}
function table(headers, rows) {
  if (!rows.length) return empty("Nothing here yet", "Try another search or add a new item.");
  return `<table><thead><tr>${headers.map(h => `<th scope="col">${escapeHTML(h)}</th>`).join("")}</tr></thead><tbody>${rows.join("")}</tbody></table>`;
}
function button(action, id, text, kind = "") {
  return `<button type="button" class="btn tiny ${kind}" data-action="${escapeHTML(action)}" data-id="${idNumber(id)}">${escapeHTML(text)}</button>`;
}
function pager(selector, paging, prev, next) {
  const root = $(selector);
  root.replaceChildren();
  if (!paging) return;
  const info = document.createElement("span");
  info.textContent = `${paging.count ?? 0} total · Page ${paging.page ?? 1} of ${paging.totalPage || 0}`;
  const buttons = document.createElement("div");
  buttons.className = "pagination-buttons";
  for (const [label, disabled, handler] of [
    ["← Previous", paging.page <= 1, prev], ["Next →", !paging.totalPage || paging.page >= paging.totalPage, next]
  ]) {
    const b = document.createElement("button"); b.type = "button"; b.className = "btn btn-outline btn-sm";
    b.textContent = label; b.disabled = disabled; b.addEventListener("click", () => run(handler)); buttons.append(b);
  }
  root.append(info, buttons);
}
function openCard(id) { $("#" + id).hidden = false; $("#" + id).scrollIntoView({ behavior: "smooth", block: "start" }); }
function clearEditor(kind) { $("#" + kind + "-form").reset(); $("#" + kind + "-form-title").textContent = `New ${kind}`; }
function openNew(kind) { clearEditor(kind); openCard(kind + "-editor"); }
function fieldsDetail(values) {
  const list = document.createElement("dl"); list.className = "detail-fields";
  for (const [name, value] of Object.entries(values)) {
    const dt = document.createElement("dt"); dt.textContent = name;
    const dd = document.createElement("dd"); dd.textContent = String(value ?? "—");
    list.append(dt, dd);
  }
  return list;
}
function revealDetails(kind, child) {
  const holder = $("#" + kind + "-detail");
  holder.replaceChildren(child);
  openCard(kind + "-detail-card");
}
function imageNode(src, cls, alt) {
  const image = document.createElement("img"); image.src = src; image.alt = alt; image.className = cls; image.loading = "lazy"; return image;
}

// Account: login, signup, profile, logout.
async function loadProfile() {
  const { user } = await api("/auth/profile");
  state.profile = user;
  updateSession();
  const holder = $("#profile-detail");
  holder.replaceChildren(fieldsDetail({ Name: user.name, Email: user.email, Role: user.role, "User ID": user.id }));
  if (user.avatar) holder.append(imageNode(user.avatar, "detail-hero", "Your avatar"));
  return user;
}
$("#login-form").addEventListener("submit", event => {
  event.preventDefault();
  const form = event.currentTarget; // ✅ เก็บไว้ก่อน

  run(async () => {
    const { email, password } = readForm(form);
    const result = await api("/auth/sign-in", {
      method: "POST",
      body: { email, password }
    });

    const token = result?.access_token ?? result?.token;

    if (!token) {
      throw new Error("Signed in, but the API returned no token.");
    }

    setToken(token);
    await loadProfile();
    await Promise.all([loadArticles(), loadCategories()]);

    form.reset(); // ✅
    alertMessage("Signed in successfully.");
  });
});
$("#signup-form").addEventListener("submit", event => {
  event.preventDefault();

  const form = event.currentTarget;

  run(async () => {
    const { email, password } = readForm(form);

    const result = await api("/auth/signup", {
      method: "POST",
      body: { email, password }
    });

    form.reset();

    alertMessage(`Account created: ${result?.user?.email || email}. You can now sign in.`);
  });
});
$("#sign-out").addEventListener("click", () => {
  setToken("");
  for (const kind of ["article", "category", "user"]) {
    $("#" + kind + "-editor").hidden = true;
    $("#" + kind + "-detail-card").hidden = true;
  }
  $("#profile-detail").textContent = "Sign in to view your profile.";
  switchTab("articles");
  run(loadArticles);
  alertMessage("Signed out of this browser. Your existing JWT can remain valid until it expires.");
});
$("#go-sign-in").addEventListener("click", () => switchTab("account"));
$("#profile-reload").addEventListener("click", () => run(loadProfile));
$("#profile-form").addEventListener("submit", event => {
  event.preventDefault();

  const form = event.currentTarget; // ✅

  run(async () => {
    const { user } = await api("/auth/profile", {
      method: "PATCH",
      body: collectMultipart(form, ["email", "name", "avatar"])
    });

    form.reset();
    await loadProfile();

    alertMessage(`Profile updated for ${user.name || user.email}.`);
  });
});

// Articles: list/filter/view/edit/create/delete.
async function loadArticles() {
  const f = readForm($("#article-search"));
  const result = await api(`/articles${params({ term: f.term, categoryId: f.categoryId, page: state.articlePage, limit: f.limit || 10 })}`);
  const items = result?.articles?.items || [];
  $("#article-list").innerHTML = table(["Article", "Category", "Author", "Actions"], items.map(a => `<tr>
    <td><div class="table-name-cell">${a.image ? `<img class="table-thumb" alt="" src="${escapeHTML(a.image)}" loading="lazy">` : ""}<div><span class="cell-primary">${escapeHTML(a.title)}</span><span class="cell-secondary">#${idNumber(a.id)} · ${escapeHTML(a.excerpt || "No summary")}</span></div></div></td>
    <td><span class="chip">${escapeHTML(a.category?.name || a.category_id || "—")}</span></td><td>${escapeHTML(a.user?.name || "—")}</td>
    <td><div class="actions">${button("article-view", a.id, "View", "primary")}${canEditArticles() ? button("article-edit", a.id, "Edit") + button("article-delete", a.id, "Delete", "danger") : ""}</div></td></tr>`));
  pager("#article-pager", result?.articles?.paging, () => { state.articlePage--; return loadArticles(); }, () => { state.articlePage++; return loadArticles(); });
}
async function showArticle(id, forEdit = false) {
  const { article } = await api(`/articles/${id}`);
  const content = document.createElement("div");
  const title = document.createElement("h3"); title.className = "detail-article-title"; title.textContent = article.title;
  const meta = document.createElement("div"); meta.className = "detail-meta";
  for (const label of [`Category: ${article.category?.name || article.category_id || "—"}`, `Author: ${article.user?.name || "—"}`, `Article #${article.id}`]) {
    const span = document.createElement("span"); span.textContent = label; meta.append(span);
  }
  content.append(title, meta);
  if (article.image) content.append(imageNode(article.image, "detail-hero", article.title));
  if (article.excerpt) { const intro = document.createElement("p"); intro.textContent = article.excerpt; intro.style.marginTop = "15px"; content.append(intro); }
  const body = document.createElement("div"); body.className = "detail-body"; body.textContent = article.body || ""; content.append(body);
  if (forEdit) {
    const form = $("#article-form");
    for (const key of ["id", "title", "excerpt", "body"]) form.elements.namedItem(key).value = article[key] ?? "";
    form.elements.namedItem("category_id").value = article.category_id ?? "";
    form.elements.namedItem("image").value = "";
    $("#article-form-title").textContent = `Edit article #${id}`;
    openCard("article-editor");
  } else revealDetails("article", content);
}
$("#article-search").addEventListener("submit", event => { event.preventDefault(); state.articlePage = 1; run(loadArticles); });
$("#article-new").addEventListener("click", () => openNew("article"));
$("#article-reset").addEventListener("click", () => clearEditor("article"));
$("#article-form").addEventListener("submit", event => {
  event.preventDefault();

  const form = event.currentTarget;

  run(async () => {
    const id = form.elements.namedItem("id").value.trim();

    const body = collectMultipart(form, [
      "title",
      "body",
      "excerpt",
      "category_id",
      "image"
    ]);

    if (!id && ["title", "body", "excerpt", "category_id", "image"].some(key => !body.has(key))) {
      throw new Error("New articles require a title, summary, body, category and image.");
    }

    const result = await api(
        id ? `/articles/${id}` : "/articles",
        {
          method: id ? "PATCH" : "POST",
          body
        }
    );

    clearEditor("article");
    $("#article-editor").hidden = true;

    state.articlePage = 1;
    await loadArticles();

    alertMessage(`Article ${id ? "updated" : "created"} (ID ${result?.article?.id || result?.articles?.id || id}).`);
  });
});

// Categories: full CRUD.
async function loadCategories() {
  const result = await api("/categories");
  const items = result?.categories || [];
  $("#category-list").innerHTML = table(["Category", "Description", "Actions"], items.map(c => `<tr>
    <td><span class="cell-primary">${escapeHTML(c.name)}</span><span class="cell-secondary">Category #${idNumber(c.id)}</span></td>
    <td>${escapeHTML(c.description)}</td><td><div class="actions">${button("category-view", c.id, "View", "primary")}${isAdmin() ? button("category-edit", c.id, "Edit") + button("category-delete", c.id, "Delete", "danger") : ""}</div></td></tr>`));
  for (const selector of ["#article-filter-category", "#article-edit-category"]) {
    const select = $(selector), old = select.value;
    select.replaceChildren(new Option(selector.includes("filter") ? "All categories" : "Choose category", ""));
    for (const item of items) select.add(new Option(item.name, item.id));
    select.value = old;
  }
}
async function showCategory(id, forEdit = false) {
  const { category } = await api(`/categories/${id}`);
  if (forEdit) {
    const form = $("#category-form");
    for (const key of ["id", "name", "description"]) form.elements.namedItem(key).value = category[key] ?? "";
    $("#category-form-title").textContent = `Edit category #${id}`;
    openCard("category-editor");
  } else revealDetails("category", fieldsDetail({ Name: category.name, Description: category.description, "Category ID": category.id, "Article count": category.articles?.length ?? 0 }));
}
$("#categories-reload").addEventListener("click", () => run(loadCategories));
$("#category-new").addEventListener("click", () => openNew("category"));
$("#category-reset").addEventListener("click", () => clearEditor("category"));
$("#category-form").addEventListener("submit", event => {
  event.preventDefault();
  const form = event.currentTarget;
  run(async () => {
    const fields = readForm(form);
    if (!fields.id && (!fields.name || !fields.description)) throw new Error("Enter a name and description for the new category.");
    await api(fields.id ? `/categories/${fields.id}` : "/categories", { method: fields.id ? "PATCH" : "POST", body: jsonObject(fields, ["name", "description"]) });
    clearEditor("category"); $("#category-editor").hidden = true;
    await loadCategories();
    alertMessage(`Category ${fields.id ? "updated" : "created"}.`);
  });
});

// Dashboard: Admin only.
async function loadDashboard() {
  if (!isAdmin()) { $("#dashboard-data").innerHTML = `<div class="card empty-state"><div class="empty-symbol">♙</div><strong>Admin access required</strong>Sign in as an Admin to view application statistics.</div>`; return; }
  const { dashboard } = await api("/dashboard");
  const root = $("#dashboard-data"); root.replaceChildren();
  const stats = document.createElement("div"); stats.className = "stat-grid";
  for (const [label, value] of [["Articles", dashboard.articlesCount], ["Categories", dashboard.categoriesCount], ["Total users", (dashboard.usersCount || []).reduce((sum, u) => sum + Number(u.count || 0), 0)]]) {
    const card = document.createElement("div"); card.className = "stat";
    const t = document.createElement("span"); t.textContent = label;
    const v = document.createElement("strong"); v.textContent = value ?? 0;
    card.append(t, v); stats.append(card);
  }
  root.append(stats);
  for (const [heading, items, renderer] of [
    ["Users by role", dashboard.usersCount || [], u => `${u.role}: ${u.count}`],
    ["Latest articles", dashboard.latestArticles || [], a => `${a.title} (#${a.id})`]
  ]) {
    const card = document.createElement("div"); card.className = "card dashboard-block";
    const h = document.createElement("h3"); h.textContent = heading; card.append(h);
    for (const item of items) { const line = document.createElement("div"); line.className = "dashboard-line"; line.textContent = renderer(item); card.append(line); }
    if (!items.length) { const span = document.createElement("p"); span.textContent = "No data yet."; card.append(span); }
    root.append(card);
  }
}
$("#dashboard-reload").addEventListener("click", () => run(loadDashboard));

// Users: Admin-only full CRUD, promote/demote.
async function loadUsers() {
  if (!isAdmin()) { $("#user-list").innerHTML = empty("Admin access required", "Sign in as an Admin to manage users."); $("#user-pager").replaceChildren(); return; }
  const f = readForm($("#user-search"));
  const result = await api(`/users${params({ term: f.term, page: state.userPage, limit: f.limit || 10 })}`);
  const items = result?.users?.items || [];
  $("#user-list").innerHTML = table(["User", "Email", "Role", "Actions"], items.map(u => `<tr>
    <td><span class="cell-primary">${escapeHTML(u.name || "Unnamed user")}</span><span class="cell-secondary">User #${idNumber(u.id)}</span></td>
    <td>${escapeHTML(u.email)}</td><td>${roleChip(u.role)}</td>
    <td><div class="actions">${button("user-view", u.id, "View", "primary")}${button("user-edit", u.id, "Edit")}${u.role === "Member" ? button("user-promote", u.id, "Promote") : ""}${u.role === "Editor" ? button("user-demote", u.id, "Demote") : ""}${button("user-delete", u.id, "Delete", "danger")}</div></td></tr>`));
  pager("#user-pager", result?.users?.paging, () => { state.userPage--; return loadUsers(); }, () => { state.userPage++; return loadUsers(); });
}
async function showUser(id, forEdit = false) {
  const { user } = await api(`/users/${id}`);
  if (forEdit) {
    const form = $("#user-form");
    for (const key of ["id", "name", "email"]) form.elements.namedItem(key).value = user[key] ?? "";
    form.elements.namedItem("password").value = "";
    $("#user-form-title").textContent = `Edit user #${id}`;
    openCard("user-editor");
  } else {
    const content = document.createElement("div");
    content.append(fieldsDetail({ Name: user.name, Email: user.email, Role: user.role, "User ID": user.id }));
    if (user.avatar) content.append(imageNode(user.avatar, "detail-hero", `${user.name} avatar`));
    revealDetails("user", content);
  }
}
$("#user-search").addEventListener("submit", event => { event.preventDefault(); state.userPage = 1; run(loadUsers); });
$("#user-new").addEventListener("click", () => openNew("user"));
$("#user-reset").addEventListener("click", () => clearEditor("user"));
$("#user-form").addEventListener("submit", event => {
  event.preventDefault();

  const form = event.currentTarget;

  run(async () => {
    const fields = readForm(form);

    if (!fields.id && (!fields.name || !fields.email || !fields.password)) {
      throw new Error("New users require a name, email and password.");
    }

    await api(
        fields.id ? `/users/${fields.id}` : "/users",
        {
          method: fields.id ? "PATCH" : "POST",
          body: jsonObject(fields, ["name", "email", "password"])
        }
    );

    clearEditor("user");
    $("#user-editor").hidden = true;

    state.userPage = 1;
    await loadUsers();

    alertMessage(`User ${fields.id ? "updated" : "created"}.`);
  });
});

// Delegated action buttons for all tables.
document.addEventListener("click", event => {
  const clicked = event.target.closest("button[data-action][data-id]");
  if (!clicked) return;
  const id = idNumber(clicked.dataset.id), action = clicked.dataset.action;
  if (!id) return;
  run(async () => {
    if (action === "article-view") return showArticle(id);
    if (action === "article-edit" && canEditArticles()) return showArticle(id, true);
    if (action === "category-view") return showCategory(id);
    if (action === "category-edit" && isAdmin()) return showCategory(id, true);
    if (action === "user-view" && isAdmin()) return showUser(id);
    if (action === "user-edit" && isAdmin()) return showUser(id, true);
    if (action.endsWith("-delete")) {
      const name = action.split("-")[0];
      if ((name === "article" && !canEditArticles()) || (name !== "article" && !isAdmin())) throw new Error("You don't have permission for this action.");
      if (!confirm(`Permanently delete ${name} #${id}? This cannot be undone.`)) return;
      await api(`/${name === "category" ? "categories" : name + "s"}/${id}`, { method: "DELETE" });
      if (name === "article") await loadArticles();
      else if (name === "category") { await loadCategories(); await loadArticles(); }
      else await loadUsers();
      return alertMessage(`${name} #${id} deleted.`);
    }
    if ((action === "user-promote" || action === "user-demote") && isAdmin()) {
      const mode = action === "user-promote" ? "promote" : "demote";
      if (!confirm(`${mode === "promote" ? "Promote" : "Demote"} user #${id}?`)) return;
      await api(`/users/${id}/${mode}`, { method: "PATCH" });
      await loadUsers();
      return alertMessage(`User #${id} ${mode === "promote" ? "promoted" : "demoted"}.`);
    }
  });
});
for (const button of document.querySelectorAll("[data-close]")) button.addEventListener("click", () => { $("#" + button.dataset.close).hidden = true; });

// Generic JSON API tester. Form-specific multipart uploads live in the corresponding page.
$("#test-form").addEventListener("submit", event => {
  event.preventDefault();

  const form = event.currentTarget;

  run(async () => {
    const { method, path: entered, body: raw } = readForm(form);

    const path = entered.trim();

    if (!/^\/[a-zA-Z0-9/_?=&%+.-]*$/.test(path)) {
      throw new Error("Enter a path like /articles?limit=5");
    }

    let body;

    if (raw.trim()) {
      try {
        body = JSON.parse(raw);
      } catch {
        throw new Error("The JSON body is invalid.");
      }
    }

    const result = await api(path, {
      method,
      ...(body === undefined ? {} : { body })
    });

    $("#test-result").textContent =
        result === null
            ? "204 No Content"
            : JSON.stringify(result, null, 2);

    alertMessage(`${method} ${path} succeeded.`);
  });
});

for (const el of document.querySelectorAll("[data-tab]")) el.addEventListener("click", () => {
  const tab = el.dataset.tab;
  switchTab(tab);
  if (tab === "articles") run(loadArticles);
  if (tab === "categories") run(loadCategories);
  if (tab === "dashboard") run(loadDashboard);
  if (tab === "users") run(loadUsers);
  if (tab === "account" && state.token) run(loadProfile);
});

updateSession();
run(async () => {
  await loadCategories();
  await loadArticles();
  if (state.token) { try { await loadProfile(); await loadArticles(); } catch (error) { showError(error); } }
});
