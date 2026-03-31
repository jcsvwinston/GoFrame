(function () {
  "use strict";

  function escapeHtml(value) {
    const div = document.createElement("div");
    div.textContent = value === null || value === undefined ? "" : String(value);
    return div.innerHTML;
  }

  function sectionHead(title, subtitle, badge) {
    const badgeHTML = badge ? `<span class="status-chip">${escapeHtml(badge)}</span>` : "";
    return `
      <section class="section-head">
        <div>
          <h2 class="section-title">${escapeHtml(title || "")}</h2>
          <p class="section-subtitle">${escapeHtml(subtitle || "")}</p>
        </div>
        ${badgeHTML}
      </section>
    `;
  }

  function loading() {
    return `
      <div class="loading-lines">
        <div class="loading-line"></div>
        <div class="loading-line"></div>
        <div class="loading-line"></div>
      </div>
    `;
  }

  function empty(message) {
    return `<div class="table-empty">${escapeHtml(message || "No data")}</div>`;
  }

  function error(title, message, actionLabel, actionID) {
    const btn = actionLabel
      ? `<button class="btn btn-primary" type="button" id="${escapeHtml(actionID || "error-retry")}">${escapeHtml(actionLabel)}</button>`
      : "";
    return `
      <section class="error-state" role="alert">
        <h3>${escapeHtml(title || "Request failed")}</h3>
        <p>${escapeHtml(message || "An unexpected error occurred.")}</p>
        ${btn}
      </section>
    `;
  }

  function kv(label, value) {
    return `
      <article class="detail-card">
        <p class="detail-label">${escapeHtml(label || "")}</p>
        <p class="detail-value">${escapeHtml(value || "-")}</p>
      </article>
    `;
  }

  window.AdminUI = {
    escapeHtml: escapeHtml,
    sectionHead: sectionHead,
    loading: loading,
    empty: empty,
    error: error,
    kv: kv,
  };
})();
