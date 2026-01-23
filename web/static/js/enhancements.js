(function() {
  function setupAutoRefresh() {
    const cfg = (window.dashboardUI) || {};
    const enabled = !!cfg.enableAutoRefresh;
    const seconds = Number(cfg.autoRefreshSeconds) > 0 ? Number(cfg.autoRefreshSeconds) : 0;
    if (enabled && seconds > 0) {
      setTimeout(() => window.location.reload(), seconds * 1000);
    }
  }

  function enhanceTable(table) {
    if (!table || table.dataset.enhanced === 'true') return;
    
    // Skip tables in hidden tab panes
    const tabPane = table.closest('.tab-pane');
    if (tabPane) {
      console.log('Table in tab pane:', tabPane.id, 'active:', tabPane.classList.contains('active'));
    }
    if (tabPane && !tabPane.classList.contains('active')) {
      console.log('Skipping table in inactive pane:', tabPane.id);
      return;
    }
    
    console.log('Enhancing table in pane:', tabPane?.id || 'no pane');
    const tbody = table.querySelector('tbody');
    if (!tbody) return;
    const rows = Array.from(tbody.querySelectorAll('tr'));
    const perPage = parseInt(table.dataset.perPage || '25', 10) || 25;
    const enableTagFilter = table.dataset.tagFilter === 'true';
    let filtered = rows.slice();
    let sortIndex = null;
    let sortDir = 1;
    let currentPage = 1;
    let activeTag = 'all';
    const tagSet = new Set();
    if (enableTagFilter) {
      rows.forEach(r => {
        const tagsAttr = r.dataset.tags || '';
        tagsAttr.split(',').map(t => t.trim()).filter(Boolean).forEach(t => tagSet.add(t));
      });
    }

    // Controls wrapper
    const controls = document.createElement('div');
    controls.className = 'table-enhance-controls d-flex align-items-center justify-content-between gap-2 mb-2 flex-wrap';
    const search = document.createElement('input');
    search.type = 'search';
    search.className = 'form-control form-control-sm';
    search.placeholder = 'Search table...';
    search.style.maxWidth = '240px';
    const tagFilters = document.createElement('div');
    tagFilters.className = 'd-flex flex-wrap gap-2';
    const info = document.createElement('div');
    info.className = 'small text-muted';
    controls.appendChild(search);
    if (enableTagFilter && tagSet.size > 0) {
      const allBtn = document.createElement('button');
      allBtn.type = 'button';
      allBtn.className = 'btn btn-outline-secondary btn-sm active';
      allBtn.textContent = 'All';
      allBtn.addEventListener('click', () => {
        activeTag = 'all';
        updateTagButtons(allBtn);
        currentPage = 1;
        render();
      });
      tagFilters.appendChild(allBtn);
      Array.from(tagSet).sort().forEach(tag => {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'btn btn-outline-secondary btn-sm';
        btn.textContent = tag;
        btn.dataset.tag = tag;
        btn.addEventListener('click', () => {
          activeTag = tag;
          updateTagButtons(btn);
          currentPage = 1;
          render();
        });
        tagFilters.appendChild(btn);
      });
      controls.appendChild(tagFilters);
    }
    controls.appendChild(info);
    function updateTagButtons(activeBtn) {
      tagFilters.querySelectorAll('button').forEach(btn => {
        btn.classList.toggle('active', btn === activeBtn);
      });
    }
    table.parentNode.insertBefore(controls, table);

    // Pagination
    const pager = document.createElement('div');
    pager.className = 'd-flex justify-content-between align-items-center mt-2 flex-wrap gap-2';
    const pageLabel = document.createElement('div');
    pageLabel.className = 'small text-muted';
    const btnGroup = document.createElement('div');
    btnGroup.className = 'btn-group btn-group-sm';
    const prev = document.createElement('button');
    prev.type = 'button';
    prev.className = 'btn btn-outline-secondary';
    prev.textContent = 'Prev';
    const next = document.createElement('button');
    next.type = 'button';
    next.className = 'btn btn-outline-secondary';
    next.textContent = 'Next';
    btnGroup.append(prev, next);
    pager.append(pageLabel, btnGroup);
    table.parentNode.appendChild(pager);

    // Sorting
    const ths = Array.from(table.querySelectorAll('thead th'));
    ths.forEach((th, idx) => {
      th.style.cursor = 'pointer';
      th.addEventListener('click', () => {
        if (sortIndex === idx) {
          sortDir *= -1;
        } else {
          sortIndex = idx;
          sortDir = 1;
        }
        render();
      });
    });

    function render() {
      const term = search.value.trim().toLowerCase();
      filtered = rows.filter(row => {
        const matchesSearch = !term || row.textContent.toLowerCase().includes(term);
        if (!matchesSearch) return false;
        if (enableTagFilter && activeTag !== 'all') {
          const tagsAttr = row.dataset.tags || '';
          const tags = tagsAttr.split(',').map(t => t.trim()).filter(Boolean);
          return tags.includes(activeTag);
        }
        return true;
      });

      if (sortIndex !== null) {
        filtered.sort((a, b) => {
          const aText = (a.children[sortIndex]?.textContent || '').trim();
          const bText = (b.children[sortIndex]?.textContent || '').trim();
          const aNum = parseFloat(aText.replace(/[^0-9.-]+/g, ''));
          const bNum = parseFloat(bText.replace(/[^0-9.-]+/g, ''));
          const aIsNum = !isNaN(aNum);
          const bIsNum = !isNaN(bNum);
          if (aIsNum && bIsNum) {
            return (aNum - bNum) * sortDir;
          }
          return aText.localeCompare(bText) * sortDir;
        });
      }

      const totalPages = Math.max(1, Math.ceil(filtered.length / perPage));
      if (currentPage > totalPages) currentPage = totalPages;
      const start = (currentPage - 1) * perPage;
      const end = start + perPage;

      rows.forEach(r => (r.style.display = 'none'));
      filtered.slice(start, end).forEach(r => (r.style.display = ''));

      info.textContent = `${filtered.length} items`;
      pageLabel.textContent = `Page ${currentPage} / ${totalPages}`;
      prev.disabled = currentPage <= 1;
      next.disabled = currentPage >= totalPages;
    }

    search.addEventListener('input', () => {
      currentPage = 1;
      render();
    });
    prev.addEventListener('click', () => {
      if (currentPage > 1) {
        currentPage -= 1;
        render();
      }
    });
    next.addEventListener('click', () => {
      const totalPages = Math.max(1, Math.ceil(filtered.length / perPage));
      if (currentPage < totalPages) {
        currentPage += 1;
        render();
      }
    });

    table.dataset.enhanced = 'true';
    render();
  }

  function enhanceAllTables() {
    document.querySelectorAll('table.table-enhance').forEach(table => {
      // Only enhance tables in visible tab panes or no tab pane
      const tabPane = table.closest('.tab-pane');
      if (!tabPane || tabPane.classList.contains('active')) {
        enhanceTable(table);
      }
    });
  }

  function enhanceCardGrid(grid) {
    if (!grid || grid.dataset.enhanced === 'true') return;
    const items = Array.from(grid.querySelectorAll('[data-card-item]'));
    if (items.length === 0) return;

    const perPage = parseInt(grid.dataset.perPage || '8', 10) || 8;
    const compactThreshold = parseInt(grid.dataset.compactThreshold || '6', 10) || 6;
    let currentPage = 1;

    const pagerHost = document.querySelector(`[data-card-grid-pager="${grid.dataset.cardGrid}"]`) || grid.nextElementSibling;
    const pager = document.createElement('div');
    pager.className = 'd-flex justify-content-between align-items-center mt-2 flex-wrap gap-2 card-grid-controls';
    const info = document.createElement('div');
    info.className = 'small text-muted';
    const btnGroup = document.createElement('div');
    btnGroup.className = 'btn-group btn-group-sm';
    const prev = document.createElement('button');
    prev.type = 'button';
    prev.className = 'btn btn-outline-secondary';
    prev.textContent = 'Prev';
    const next = document.createElement('button');
    next.type = 'button';
    next.className = 'btn btn-outline-secondary';
    next.textContent = 'Next';
    btnGroup.append(prev, next);
    pager.append(info, btnGroup);

    (pagerHost || grid).appendChild(pager);

    function render() {
      const totalPages = Math.max(1, Math.ceil(items.length / perPage));
      if (currentPage > totalPages) currentPage = totalPages;
      const start = (currentPage - 1) * perPage;
      const end = start + perPage;

      items.forEach((item, idx) => {
        item.style.display = (idx >= start && idx < end) ? '' : 'none';
      });

      if (items.length > compactThreshold) {
        grid.classList.add('overview-compact');
      } else {
        grid.classList.remove('overview-compact');
      }

      const visibleCount = Math.max(0, Math.min(items.length, end) - start);
      info.textContent = `${visibleCount} shown of ${items.length}`;
      prev.disabled = currentPage <= 1;
      next.disabled = currentPage >= totalPages;
      pager.style.display = totalPages > 1 ? 'flex' : 'none';
    }

    prev.addEventListener('click', () => {
      if (currentPage > 1) {
        currentPage -= 1;
        render();
      }
    });
    next.addEventListener('click', () => {
      const totalPages = Math.max(1, Math.ceil(items.length / perPage));
      if (currentPage < totalPages) {
        currentPage += 1;
        render();
      }
    });

    grid.dataset.enhanced = 'true';
    render();
  }

  function enhanceAllCardGrids() {
    document.querySelectorAll('[data-card-grid]').forEach(enhanceCardGrid);
  }

  document.addEventListener('DOMContentLoaded', () => {
    console.log('=== Enhancements.js initializing ===');
    setupAutoRefresh();
    enhanceAllTables();
    enhanceAllCardGrids();

    // Enhance tables in newly shown tabs
    document.querySelectorAll('[data-bs-toggle="tab"]').forEach(btn => {
      btn.addEventListener('shown.bs.tab', (event) => {
        console.log('Tab shown event:', event.target.textContent.trim());
        const targetId = event.target.getAttribute('data-bs-target');
        console.log('Target pane:', targetId);
        if (targetId) {
          const targetPane = document.querySelector(targetId);
          if (targetPane) {
            console.log('Found pane, enhancing tables...');
            targetPane.querySelectorAll('table.table-enhance').forEach(table => {
              if (table.dataset.enhanced !== 'true') {
                console.log('Enhancing table in newly shown tab');
                enhanceTable(table);
              } else {
                console.log('Table already enhanced');
              }
            });
          }
        }
      });
    });
  });
})();
