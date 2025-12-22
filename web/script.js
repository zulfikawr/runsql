let currentData = null;
let currentFormat = "table";
let currentFile = null;

// DOM Elements
const fileInput = document.getElementById("fileInput");
const queryInput = document.getElementById("queryInput");
const runQueryBtn = document.getElementById("runQueryBtn");
const schemaList = document.getElementById("schemaList");
const clearFileBtn = document.getElementById("clearFileBtn");
const resultsContainer = document.getElementById("resultsContainer");
const formatBtns = document.querySelectorAll(".format-btn");
const themeToggle = document.getElementById("themeToggle");
const clearBtn = document.getElementById("clearBtn");
const formatBtn = document.getElementById("formatBtn");

// Theme Toggle (Header)
themeToggle.addEventListener("click", () => {
  const html = document.documentElement;
  const isDark = html.classList.contains("dark");

  if (isDark) {
    html.classList.remove("dark");
    themeToggle.querySelector("span").textContent = "dark_mode";
  } else {
    html.classList.add("dark");
    themeToggle.querySelector("span").textContent = "light_mode";
  }
});

// Initialize theme icon
if (document.documentElement.classList.contains("dark")) {
  themeToggle.querySelector("span").textContent = "light_mode";
}

// File upload
const revertSchemaBtn = document.getElementById("revertSchemaBtn");
let originalSchemaData = null; // Store original schema for revert

// File upload
// File upload
fileInput.addEventListener("change", async (e) => {
  if (e.target.files.length > 0) {
    const files = Array.from(e.target.files);
    currentFile = files; // Store as array

    // Reset original schema data
    originalSchemaData = null;

    // Show skeleton loading
    renderSkeletonSchema();
    document.getElementById("colCount").textContent = "...";

    // Fetch schema
    try {
      const formData = new FormData();
      files.forEach((file) => {
        formData.append("file", file);
      });

      const response = await fetch("/schema", {
        method: "POST",
        body: formData,
      });

      const data = await response.json();

      if (response.ok && data.status === "success") {
        originalSchemaData = data; // Cache original schema
        renderSchemas(data);
      } else {
        schemaList.innerHTML =
          '<div class="schema-empty">Error loading schema</div>';
      }
    } catch (err) {
      console.error(err);
      schemaList.innerHTML =
        '<div class="schema-empty">Error loading schema</div>';
    }
  }
});

function renderSkeletonSchema() {
  let html = "";
  // Show 3 skeleton groups
  for (let i = 0; i < 3; i++) {
    html += `
        <div class="schema-group" style="padding: 10px;">
            <div class="skeleton skeleton-header"></div>
            <div class="skeleton skeleton-item"></div>
            <div class="skeleton skeleton-item"></div>
        </div>`;
  }
  schemaList.innerHTML = html;
}

function sanitizeTableName(name) {
  let res = "";
  for (let i = 0; i < name.length; i++) {
    const r = name[i];
    if (
      (r >= "a" && r <= "z") ||
      (r >= "A" && r <= "Z") ||
      (r >= "0" && r <= "9") ||
      r === "_"
    ) {
      res += r;
    } else {
      res += "_";
    }
  }
  return res;
}

function getFileFormat(tableName) {
  if (!currentFile) return "";
  for (const file of currentFile) {
    // Replicate logic: base -> ext -> trim -> sanitize
    const base = file.name;
    const lastDot = base.lastIndexOf(".");
    let nameWithoutExt = base;
    let ext = "";

    if (lastDot !== -1) {
      nameWithoutExt = base.substring(0, lastDot);
      ext = base.substring(lastDot + 1).toUpperCase();
    }

    const sanitized = sanitizeTableName(nameWithoutExt);
    if (sanitized === tableName) {
      return ext;
    }
  }
  return "";
}

function sanitizeTableName(name) {
  let res = "";
  for (let i = 0; i < name.length; i++) {
    const r = name[i];
    if (
      (r >= "a" && r <= "z") ||
      (r >= "A" && r <= "Z") ||
      (r >= "0" && r <= "9") ||
      r === "_"
    ) {
      res += r;
    } else {
      res += "_";
    }
  }
  return res;
}

function getFileFormat(tableName) {
  if (!currentFile) return "";
  for (const file of currentFile) {
    // Replicate logic: base -> ext -> trim -> sanitize
    const base = file.name;
    const lastDot = base.lastIndexOf(".");
    let nameWithoutExt = base;
    let ext = "";

    if (lastDot !== -1) {
      nameWithoutExt = base.substring(0, lastDot);
      ext = base.substring(lastDot + 1).toUpperCase();
    }

    const sanitized = sanitizeTableName(nameWithoutExt);
    if (sanitized === tableName) {
      return ext;
    }
  }
  return "";
}

function renderSchemas(data) {
  if (!data.schemas && !data.columns) return;

  // Handle query result schema (single list of columns, no table names usually provided in current backend response for query)
  // If it's the specific format from /schema endpoint: data.schemas map
  // If it's from /query: data.columns array

  let html = "";
  let totalCols = 0;

  if (data.schemas) {
    // Multi-table schema
    for (const [tableName, columns] of Object.entries(data.schemas)) {
      totalCols += columns.length;
      const fmt = getFileFormat(tableName);

      html += `
            <div class="schema-group">
                <div class="schema-table-header" onclick="toggleSchema(this)">
                    <div class="schema-header-left">
                        <span class="material-symbols-outlined schema-arrow">keyboard_arrow_down</span>
                        <span>${escapeHtml(tableName)}</span>
                    </div>
                    ${fmt ? `<span class="schema-file-badge">${fmt}</span>` : ""}
                </div>
                <div class="schema-items-container">
                    ${columns
                      .map(
                        (col) => `
                        <div class="schema-item">
                            <span class="schema-col-name">${escapeHtml(col)}</span>
                            <span class="schema-col-type">TEXT</span>
                        </div>
                    `,
                      )
                      .join("")}
                </div>
            </div>`;
    }
  } else if (data.columns) {
    // Flat list (e.g. from query result)
    totalCols = data.columns.length;
    html += `
        <div class="schema-group">
            <div class="schema-table-header" onclick="toggleSchema(this)">
                <div class="schema-header-left">
                    <span class="material-symbols-outlined schema-arrow">keyboard_arrow_down</span>
                    <span>Result Columns</span>
                </div>
            </div>
            <div class="schema-items-container">
                ${data.columns
                  .map(
                    (col) => `
                    <div class="schema-item">
                        <span class="schema-col-name">${escapeHtml(col)}</span>
                        <span class="schema-col-type">TEXT</span>
                    </div>
                `,
                  )
                  .join("")}
            </div>
        </div>`;
  }

  schemaList.innerHTML = html;
  document.getElementById("colCount").textContent = `${totalCols} cols`;
}

// Toggle collapse function (attached to window for onclick access)
window.toggleSchema = function (header) {
  header.classList.toggle("collapsed");
  const container = header.nextElementSibling;
  container.classList.toggle("collapsed");
};

// Revert schema button
revertSchemaBtn.addEventListener("click", () => {
  if (originalSchemaData) {
    renderSchemas(originalSchemaData);
    // Optional: clear result/query if desired, but user just asked to revert schema view
  }
});

// Clear file button (if kept, or we can just remove this logic since we hid the section)
if (clearFileBtn) {
  clearFileBtn.addEventListener("click", () => {
    currentFile = null;
    fileInput.value = "";
    schemaList.innerHTML =
      '<div class="schema-empty">Upload a file to see schema</div>';
    document.getElementById("colCount").textContent = "0 cols";
    originalSchemaData = null;
    resultsContainer.innerHTML =
      '<div class="results-empty"><div class="empty-state"><div class="empty-icon"><span class="material-symbols-outlined">query_stats</span></div><p class="empty-text">Upload a file and run a query to see results</p></div></div>';
  });
}

// Clear query button
clearBtn.addEventListener("click", () => {
  queryInput.value = "";
  updateSQLHighlight();
});

// SQL Syntax Highlighting
const highlightLayer = document.getElementById("highlightLayer");

function highlightSQL(text) {
  // SQL keywords
  const keywords =
    /\b(SELECT|FROM|WHERE|AND|OR|NOT|IN|LIKE|LIMIT|ORDER|BY|GROUP|HAVING|JOIN|LEFT|RIGHT|INNER|OUTER|ON|AS|DISTINCT|COUNT|SUM|AVG|MIN|MAX|INSERT|UPDATE|DELETE|CREATE|DROP|ALTER|TABLE|DATABASE|VIEW|INDEX|PRIMARY|KEY|UNIQUE|FOREIGN|CHECK|DEFAULT|NULL|TRUE|FALSE)\b/gi;

  // SQL strings (both single and double quotes)
  const strings = /('(?:[^'\\]|\\.)*'|"(?:[^"\\]|\\.)*")/g;

  // SQL comments
  const comments = /(--[^\n]*|\/\*[\s\S]*?\*\/)/g;

  // SQL numbers
  const numbers = /\b(\d+\.?\d*)\b/g;

  let highlighted = text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");

  // Replace in order: comments, strings, keywords, numbers
  highlighted = highlighted.replace(
    comments,
    '<span class="sql-comment">$1</span>',
  );
  highlighted = highlighted.replace(
    strings,
    '<span class="sql-string">$1</span>',
  );
  highlighted = highlighted.replace(
    keywords,
    '<span class="sql-keyword">$1</span>',
  );
  highlighted = highlighted.replace(
    numbers,
    '<span class="sql-number">$1</span>',
  );

  return highlighted;
}

function updateSQLHighlight() {
  highlightLayer.innerHTML = highlightSQL(queryInput.value);
  highlightLayer.scrollTop = queryInput.scrollTop;
  highlightLayer.scrollLeft = queryInput.scrollLeft;
}

// Add event listeners for real-time highlighting
queryInput.addEventListener("input", () => {
  updateSQLHighlight();

  // Auto-revert schema if input is empty
  if (!queryInput.value.trim()) {
    if (originalSchemaData) {
      renderSchemas(originalSchemaData);
    }
    // Reset results stats
    document.getElementById("resultCount").textContent = "0";
    document.getElementById("resultTime").textContent = "0ms";
    // Reset results container to empty state if truly cleared? User didn't explicitly ask for this but "it should be back to original when the input is cleared"
    // "also this too: 2 results in 185ms. it should be back to original when the input is cleared/or when no text in the query."
    resultsContainer.innerHTML =
      '<div class="results-empty"><div class="empty-state"><div class="empty-icon"><span class="material-symbols-outlined">query_stats</span></div><p class="empty-text">Upload a file and run a query to see results</p></div></div>';
  }
});
queryInput.addEventListener("scroll", () => {
  highlightLayer.scrollTop = queryInput.scrollTop;
  highlightLayer.scrollLeft = queryInput.scrollLeft;
});

// Format buttons
formatBtns.forEach((btn) => {
  btn.addEventListener("click", () => {
    formatBtns.forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    currentFormat = btn.dataset.format;
    if (currentData) {
      displayResults(currentData);
    }
  });
});

// Keyboard shortcut: Ctrl+Enter to run query
queryInput.addEventListener("keydown", (e) => {
  if (e.ctrlKey && e.key === "Enter") {
    e.preventDefault();
    runQueryBtn.click();
  }
});

// Run query
runQueryBtn.addEventListener("click", async () => {
  if (!currentFile || currentFile.length === 0) {
    alert("Please upload a file first");
    return;
  }

  const query = queryInput.value;
  if (!query) {
    alert("Please enter a query");
    return;
  }

  // Show processing status
  runQueryBtn.disabled = true;
  runQueryBtn.innerHTML =
    '<span class="material-symbols-outlined" style="animation: spin 1s linear infinite;">autorenew</span>Running...';

  try {
    const formData = new FormData();
    // Append all files
    currentFile.forEach((file) => {
      formData.append("file", file);
    });

    formData.append("query", query);
    formData.append("format", currentFormat);

    const response = await fetch("/query", {
      method: "POST",
      body: formData,
    });

    const data = await response.json();

    if (!response.ok || data.status !== "success") {
      alert(data.error || "Query execution failed");
      return;
    }

    currentData = data;

    // Update schema view to show result columns
    renderSchemas(data);

    displayResults(data);
  } catch (err) {
    alert(`Error: ${err.message}`);
  } finally {
    runQueryBtn.disabled = false;
    runQueryBtn.innerHTML =
      '<span class="material-symbols-outlined">play_arrow</span>Run Query';
  }
});

// Display results
function displayResults(data) {
  const resultCountElem = document.getElementById("resultCount");
  const resultTimeElem = document.getElementById("resultTime");

  resultCountElem.textContent = data.rows.length;
  resultTimeElem.textContent = `${data.time_ms}ms`;

  if (currentFormat === "json") {
    const jsonData = data.rows.map((row) => {
      const obj = {};
      data.columns.forEach((col, i) => {
        obj[col] = row[i];
      });
      return obj;
    });
    // Syntax Highlight JSON
    const jsonStr = JSON.stringify(jsonData, null, 2);
    resultsContainer.innerHTML = `<pre>${syntaxHighlightJSON(jsonStr)}</pre>`;
  } else if (currentFormat === "csv") {
    let csvHtml = "";

    // Header
    const headerLine = data.columns
      .map(
        (col, i) =>
          `<span class="csv-col-${i % 8}">${escapeHtml(JSON.stringify(col))}</span>`,
      )
      .join('<span class="csv-comma">,</span>');
    csvHtml += headerLine + "\n";

    // Rows
    data.rows.forEach((row) => {
      const rowLine = row
        .map(
          (cell, i) =>
            `<span class="csv-col-${i % 8}">${escapeHtml(JSON.stringify(cell))}</span>`,
        )
        .join('<span class="csv-comma">,</span>');
      csvHtml += rowLine + "\n";
    });

    resultsContainer.innerHTML = `<pre>${csvHtml}</pre>`;
  } else {
    let html = "<table><thead><tr>";
    data.columns.forEach((col) => {
      html += `<th>${escapeHtml(col)}</th>`;
    });
    html += "</tr></thead><tbody>";

    if (data.rows.length === 0) {
      html += `<tr><td colspan="${data.columns.length}" style="text-align: center; padding: 2rem; color: #737373;">No results</td></tr>`;
    } else {
      data.rows.forEach((row) => {
        html += "<tr>";
        row.forEach((cell) => {
          html += `<td>${escapeHtml(String(cell))}</td>`;
        });
        html += "</tr>";
      });
    }

    html += "</tbody></table>";
    resultsContainer.innerHTML = html;
  }
}

function syntaxHighlightJSON(json) {
  if (typeof json !== "string") {
    json = JSON.stringify(json, undefined, 2);
  }
  json = json
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
  return json.replace(
    /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
    function (match) {
      var cls = "json-number";
      if (/^"/.test(match)) {
        if (/:$/.test(match)) {
          cls = "json-key";
        } else {
          cls = "json-string";
        }
      } else if (/true|false/.test(match)) {
        cls = "json-boolean";
      } else if (/null/.test(match)) {
        cls = "json-null";
      }
      return '<span class="' + cls + '">' + match + "</span>";
    },
  );
}

function escapeHtml(text) {
  const div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

// Export functionality
document.getElementById("exportBtn").addEventListener("click", () => {
  if (!currentData || currentData.rows.length === 0) {
    alert("No data to export");
    return;
  }

  let content, filename, type;

  if (currentFormat === "json") {
    const jsonData = currentData.rows.map((row) => {
      const obj = {};
      currentData.columns.forEach((col, i) => {
        obj[col] = row[i];
      });
      return obj;
    });
    content = JSON.stringify(jsonData, null, 2);
    filename = "export.json";
    type = "application/json";
  } else {
    let csv = currentData.columns.join(",") + "\n";
    currentData.rows.forEach((row) => {
      csv += row.map((cell) => JSON.stringify(cell)).join(",") + "\n";
    });
    content = csv;
    filename = "export.csv";
    type = "text/csv";
  }

  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
});
