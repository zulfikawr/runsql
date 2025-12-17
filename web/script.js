let currentData = null;
let currentFormat = "table";
let currentFile = null;

// DOM Elements
const fileInput = document.getElementById("fileInput");
const queryInput = document.getElementById("queryInput");
const runQueryBtn = document.getElementById("runQueryBtn");
const fileInfo = document.getElementById("fileInfo");
const fileInfoDiv = document.querySelector(".file-info");
const fileInfoContent = document.querySelector(".file-info-content");
const schemaList = document.getElementById("schemaList");
const clearFileBtn = document.getElementById("clearFileBtn");
const resultsContainer = document.getElementById("resultsContainer");
const formatBtns = document.querySelectorAll(".format-btn");
const themeToggle = document.getElementById("themeToggle");
const clearBtn = document.getElementById("clearBtn");
const formatBtn = document.getElementById("formatBtn");
const fileNameElem = document.getElementById("fileName");
const fileStatusElem = document.getElementById("fileStatus");

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
fileInput.addEventListener("change", async (e) => {
  if (e.target.files.length > 0) {
    currentFile = e.target.files[0];
    fileInfo.style.display = "flex";

    // Show file info content during processing
    fileInfoContent.style.display = "flex";
    fileStatusElem.textContent = "Processing...";

    // Fetch schema
    try {
      const formData = new FormData();
      formData.append("file", currentFile);

      const response = await fetch("/schema", {
        method: "POST",
        body: formData,
      });

      const data = await response.json();

      if (response.ok && data.status === "success") {
        // Update schema display
        if (data.columns) {
          schemaList.innerHTML = data.columns
            .map(
              (col) => `
                        <div class="schema-item">
                            <span class="schema-col-name">${escapeHtml(col)}</span>
                            <span class="schema-col-type">TEXT</span>
                        </div>
                    `,
            )
            .join("");
          document.getElementById("colCount").textContent =
            `${data.columns.length} cols`;
        }
        // Update status to ready
        fileStatusElem.textContent = "Ready";
      } else {
        fileNameElem.textContent = currentFile.name;
        fileStatusElem.textContent = "Error";
      }
    } catch (err) {
      fileNameElem.textContent = currentFile.name;
      fileStatusElem.textContent = "Error";
    }
  }
});

clearFileBtn.addEventListener("click", () => {
  currentFile = null;
  fileInput.value = "";
  fileInfo.style.display = "none";
  fileInfoDiv.style.display = "flex";
  schemaList.innerHTML =
    '<div class="schema-empty">Upload a file to see schema</div>';
  document.getElementById("colCount").textContent = "0 cols";
  resultsContainer.innerHTML =
    '<div class="results-empty"><div class="empty-state"><div class="empty-icon"><span class="material-symbols-outlined">query_stats</span></div><p class="empty-text">Upload a file and run a query to see results</p></div></div>';
});

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
queryInput.addEventListener("input", updateSQLHighlight);
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
  if (!currentFile) {
    alert("Please upload a file first");
    return;
  }

  const query = queryInput.value;
  if (!query) {
    alert("Please enter a query");
    return;
  }

  // Show processing status
  fileStatusElem.textContent = "Processing...";
  runQueryBtn.disabled = true;
  runQueryBtn.innerHTML =
    '<span class="material-symbols-outlined" style="animation: spin 1s linear infinite;">autorenew</span>Running...';

  try {
    const formData = new FormData();
    formData.append("file", currentFile);
    formData.append("query", query);
    formData.append("format", currentFormat);

    const response = await fetch("/query", {
      method: "POST",
      body: formData,
    });

    const data = await response.json();

    if (!response.ok || data.status !== "success") {
      alert(data.error || "Query execution failed");
      fileStatusElem.textContent = "Ready";
      return;
    }

    currentData = data;
    fileStatusElem.textContent = "Ready";

    // Update schema
    if (data.columns) {
      schemaList.innerHTML = data.columns
        .map(
          (col) => `
                <div class="schema-item">
                    <span class="schema-col-name">${escapeHtml(col)}</span>
                    <span class="schema-col-type">TEXT</span>
                </div>
            `,
        )
        .join("");
      document.getElementById("colCount").textContent =
        `${data.columns.length} cols`;
    }

    displayResults(data);
  } catch (err) {
    alert(`Error: ${err.message}`);
    fileStatusElem.textContent = "Ready";
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
    resultsContainer.innerHTML = `<pre>${escapeHtml(JSON.stringify(jsonData, null, 2))}</pre>`;
  } else if (currentFormat === "csv") {
    let csv = data.columns.join(",") + "\n";
    data.rows.forEach((row) => {
      csv += row.map((cell) => JSON.stringify(cell)).join(",") + "\n";
    });
    resultsContainer.innerHTML = `<pre>${escapeHtml(csv)}</pre>`;
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
