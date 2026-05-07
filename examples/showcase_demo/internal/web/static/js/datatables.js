console.log('Script loaded');

let articlesCurrentPage = 0;
let authorsCurrentPage = 0;
let commentsCurrentPage = 0;
let pageSize = 25;

async function loadStats() {
  try {
    const response = await fetch('/api/stats');
    const data = await response.json();
    document.getElementById('stat-articles').textContent = data.articles || '-';
    document.getElementById('stat-authors').textContent = data.authors || '-';
    document.getElementById('stat-comments').textContent = data.comments || '-';
  } catch (error) {
    console.error('Error loading stats:', error);
  }
}

async function loadArticles(page = 0) {
  try {
    const response = await fetch('/api/datatables/articles', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ draw: 1, start: page * pageSize, length: pageSize })
    });
    const data = await response.json();
    console.log('Articles data:', data);
    
    const tbody = document.getElementById('articles-tbody');
    tbody.innerHTML = '';
    
    if (data.data && data.data.length > 0) {
      data.data.forEach(row => {
        tbody.innerHTML += `
          <tr>
            <td>${row.id}</td>
            <td>${row.title}</td>
            <td>${row.author}</td>
            <td>${row.category}</td>
            <td>${row.published ? 'Sí' : 'No'}</td>
            <td>${row.view_count}</td>
            <td>${row.published_at}</td>
          </tr>
        `;
      });
      
      // Update pagination info
      articlesCurrentPage = page;
      updatePagination('articles', data.recordsTotal, page);
    } else {
      tbody.innerHTML = '<tr><td colspan="7">No hay datos</td></tr>';
    }
  } catch (error) {
    console.error('Error loading articles:', error);
    document.getElementById('articles-tbody').innerHTML = '<tr><td colspan="7" class="error">Error: ' + error.message + '</td></tr>';
  }
}

async function loadAuthors(page = 0) {
  try {
    const response = await fetch('/api/datatables/authors', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ draw: 1, start: page * pageSize, length: pageSize })
    });
    const data = await response.json();
    console.log('Authors data:', data);
    
    const tbody = document.getElementById('authors-tbody');
    tbody.innerHTML = '';
    
    if (data.data && data.data.length > 0) {
      data.data.forEach(row => {
        tbody.innerHTML += `
          <tr>
            <td>${row.id}</td>
            <td>${row.name}</td>
            <td>${row.email}</td>
            <td>${row.position}</td>
            <td>${row.article_count}</td>
            <td>${row.social_github}</td>
            <td>${row.created_at}</td>
          </tr>
        `;
      });
      
      authorsCurrentPage = page;
      updatePagination('authors', data.recordsTotal, page);
    } else {
      tbody.innerHTML = '<tr><td colspan="7">No hay datos</td></tr>';
    }
  } catch (error) {
    console.error('Error loading authors:', error);
    document.getElementById('authors-tbody').innerHTML = '<tr><td colspan="7" class="error">Error: ' + error.message + '</td></tr>';
  }
}

async function loadComments(page = 0) {
  try {
    const response = await fetch('/api/datatables/comments', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ draw: 1, start: page * pageSize, length: pageSize })
    });
    const data = await response.json();
    console.log('Comments data:', data);
    
    const tbody = document.getElementById('comments-tbody');
    tbody.innerHTML = '';
    
    if (data.data && data.data.length > 0) {
      data.data.forEach(row => {
        tbody.innerHTML += `
          <tr>
            <td>${row.id}</td>
            <td>${row.article_id}</td>
            <td>${row.author_name}</td>
            <td>${row.author_email}</td>
            <td>${row.content.substring(0, 50)}...</td>
            <td>${row.approved ? 'Sí' : 'No'}</td>
            <td>${row.created_at}</td>
          </tr>
        `;
      });
      
      commentsCurrentPage = page;
      updatePagination('comments', data.recordsTotal, page);
    } else {
      tbody.innerHTML = '<tr><td colspan="7">No hay datos</td></tr>';
    }
  } catch (error) {
    console.error('Error loading comments:', error);
    document.getElementById('comments-tbody').innerHTML = '<tr><td colspan="7" class="error">Error: ' + error.message + '</td></tr>';
  }
}

function updatePagination(table, totalRecords, currentPage) {
  const totalPages = Math.ceil(totalRecords / pageSize);
  const start = currentPage * pageSize + 1;
  const end = Math.min((currentPage + 1) * pageSize, totalRecords);
  
  let paginationHtml = `
    <div style="padding: 15px; background: #f5f5f5; border-top: 1px solid #ddd;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">
        <span>Mostrando ${start} a ${end} de ${totalRecords} registros</span>
        <div>
          <label style="margin-right: 10px;">Registros por página:</label>
          <select id="pagesize-${table}" onchange="changePageSize('${table}', this.value)" style="padding: 5px;">
            <option value="10" ${pageSize === 10 ? 'selected' : ''}>10</option>
            <option value="25" ${pageSize === 25 ? 'selected' : ''}>25</option>
            <option value="50" ${pageSize === 50 ? 'selected' : ''}>50</option>
            <option value="100" ${pageSize === 100 ? 'selected' : ''}>100</option>
          </select>
        </div>
      </div>
      <div style="display: flex; justify-content: center; align-items: center; gap: 10px;">
        <button id="first-${table}" ${currentPage === 0 ? 'disabled' : ''} style="padding: 5px 10px;">⏮ Primero</button>
        <button id="prev-${table}" ${currentPage === 0 ? 'disabled' : ''} style="padding: 5px 10px;">◀ Anterior</button>
        <span>Página ${currentPage + 1} de ${totalPages}</span>
        <button id="next-${table}" ${currentPage >= totalPages - 1 ? 'disabled' : ''} style="padding: 5px 10px;">Siguiente ▶</button>
        <button id="last-${table}" ${currentPage >= totalPages - 1 ? 'disabled' : ''} style="padding: 5px 10px;">Último ⏭</button>
      </div>
    </div>
  `;
  
  const tableElement = document.getElementById(table + '-table');
  const existingPagination = tableElement.nextElementSibling;
  if (existingPagination && existingPagination.classList.contains('pagination-container')) {
    existingPagination.remove();
  }
  
  const paginationDiv = document.createElement('div');
  paginationDiv.className = 'pagination-container';
  paginationDiv.innerHTML = paginationHtml;
  tableElement.parentNode.insertBefore(paginationDiv, tableElement.nextSibling);
  
  // Add event listeners
  const firstBtn = document.getElementById('first-' + table);
  const prevBtn = document.getElementById('prev-' + table);
  const nextBtn = document.getElementById('next-' + table);
  const lastBtn = document.getElementById('last-' + table);
  
  if (firstBtn && !firstBtn.disabled) {
    firstBtn.addEventListener('click', () => loadTablePage(table, 0));
  }
  if (prevBtn && !prevBtn.disabled) {
    prevBtn.addEventListener('click', () => loadTablePage(table, currentPage - 1));
  }
  if (nextBtn && !nextBtn.disabled) {
    nextBtn.addEventListener('click', () => loadTablePage(table, currentPage + 1));
  }
  if (lastBtn && !lastBtn.disabled) {
    lastBtn.addEventListener('click', () => loadTablePage(table, totalPages - 1));
  }
}

window.loadTablePage = function(table, page) {
  console.log('loadTablePage called:', table, page);
  if (page < 0) return;
  
  if (table === 'articles') loadArticles(page);
  if (table === 'authors') loadAuthors(page);
  if (table === 'comments') loadComments(page);
}

window.changePageSize = function(table, newSize) {
  console.log('changePageSize called:', table, newSize);
  pageSize = parseInt(newSize);
  
  // Reset to first page when changing page size
  if (table === 'articles') loadArticles(0);
  if (table === 'authors') loadAuthors(0);
  if (table === 'comments') loadComments(0);
}

// Tab switching
document.querySelectorAll('.tab-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
    btn.classList.add('active');
    document.getElementById('tab-' + btn.dataset.tab).classList.add('active');
    
    // Load data when tab changes
    if (btn.dataset.tab === 'authors') loadAuthors(0);
    if (btn.dataset.tab === 'comments') loadComments(0);
  });
});

// Load data on page load
document.addEventListener('DOMContentLoaded', () => {
  console.log('DOM Content Loaded');
  loadStats();
  loadArticles(0);
});
