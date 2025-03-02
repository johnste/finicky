// Helper function to manage API cache
function getCachedData(key, expiryMinutes = 60) {
  const cachedData = localStorage.getItem(key);
  if (!cachedData) return null;

  try {
    const { timestamp, data } = JSON.parse(cachedData);
    const now = new Date().getTime();
    const expiryTime = timestamp + expiryMinutes * 60 * 1000;

    // Return cached data if it's still valid
    if (now < expiryTime) {
      console.log(`Using cached data for ${key}`);
      return data;
    }
  } catch (error) {
    console.error("Error parsing cached data:", error);
  }

  return null;
}

function setCachedData(key, data) {
  const cacheObject = {
    timestamp: new Date().getTime(),
    data: data,
  };

  try {
    localStorage.setItem(key, JSON.stringify(cacheObject));
    console.log(`Cached data for ${key}`);
  } catch (error) {
    console.error("Error caching data:", error);
  }
}

async function fetchWithCache(url, cacheKey, expiryMinutes = 60) {
  // Get cache data (with longer expiry time for fallback)
  const cachedData = getCachedData(cacheKey, expiryMinutes);

  // Check if we have a cache that's not too old (less than 10 minutes)
  const cacheIsRecent = checkCacheRecency(cacheKey, 10);

  // If we have very recent cache (< 10 min), use it without making a new request
  if (cachedData && cacheIsRecent) {
    console.log(`Using recent cached data for ${cacheKey} (< 10 minutes old)`);
    return cachedData;
  }

  // Otherwise try to fetch fresh data
  try {
    console.log(`Fetching fresh data for ${cacheKey}`);
    const response = await fetch(url, { mode: "cors" });
    if (!response.ok) {
      throw new Error(`API request failed: ${response.status}`);
    }

    const data = await response.json();

    // Cache the successful response
    setCachedData(cacheKey, data);
    return data;
  } catch (error) {
    console.error(`Error fetching ${url}:`, error);

    // If we have old cache, use it as fallback
    if (cachedData) {
      console.log(
        `Using cached data for ${cacheKey} as fallback after fetch error`
      );
      return cachedData;
    }

    // Last resort: check for expired cache
    const oldCache = localStorage.getItem(cacheKey);
    if (oldCache) {
      try {
        const { data } = JSON.parse(oldCache);
        console.log(`Using expired cache for ${cacheKey} due to fetch error`);
        return data;
      } catch (e) {
        console.error("Error parsing fallback cache:", e);
      }
    }

    throw error;
  }
}

// Helper function to check if cache is recent enough (within minutes threshold)
function checkCacheRecency(key, minutesThreshold = 10) {
  const cachedItem = localStorage.getItem(key);
  if (!cachedItem) return false;

  try {
    const { timestamp } = JSON.parse(cachedItem);
    const now = new Date().getTime();
    const ageInMinutes = (now - timestamp) / (1000 * 60);

    return ageInMinutes < minutesThreshold;
  } catch (error) {
    console.error("Error checking cache recency:", error);
    return false;
  }
}

async function showDownloads() {
  try {
    const result = await fetchWithCache(
      "https://api.github.com/repos/johnste/finicky/releases",
      "finicky-releases",
      120 // Cache for 2 hours
    );

    let validEntries = result.filter((v) => v?.assets?.[0]);

    let total = validEntries.reduce(
      (total, v) => total + v?.assets?.[0]?.download_count,
      0
    );

    let downloads = validEntries
      .filter((v) => v?.assets?.[0])
      .map((v) => [v.tag_name, v.assets[0].download_count]);

    let rows = downloads
      .map(([version, downloads]) => {
        return `<tr><td>${version}</td><td>${downloads.toLocaleString()}</td></tr>`;
      })
      .join("");

    document.querySelector(".download-count tbody").innerHTML =
      `<tr class="total"><td>total</td><td>${total.toLocaleString()}</td></tr>` +
      rows;

    // Return download data for possible use in charts
    return {
      total,
      versions: downloads,
    };
  } catch (error) {
    console.error("Error fetching downloads:", error);
    return null;
  }
}

async function showStarGazers() {
  try {
    const result = await fetchWithCache(
      "https://api.github.com/repos/johnste/finicky",
      "finicky-repo",
      180 // Cache for 3 hours
    );

    const starCount = result.stargazers_count;
    const starCountElement = document.querySelector(".star-count");

    if (starCountElement) {
      animateCount(starCountElement, 0, starCount, 1000);
    }

    return result;
  } catch (error) {
    console.error("Failed to fetch star count:", error);
    document.querySelector(".star-count").textContent = "⭐";
    return null;
  }
}

// Fetch repo activity data
async function fetchRepoActivity() {
  try {
    // Fetch commit activity (last year of weekly commit counts)
    const commitActivity = await fetchWithCache(
      "https://api.github.com/repos/johnste/finicky/stats/commit_activity",
      "finicky-commits",
      240 // Cache for 4 hours
    );

    // Fetch contributors with timeline data
    const contributors = await fetchWithCache(
      "https://api.github.com/repos/johnste/finicky/stats/contributors",
      "finicky-contributors-stats",
      360 // Cache for 6 hours
    );

    // Fetch issues
    const issues = await fetchWithCache(
      "https://api.github.com/repos/johnste/finicky/issues?state=all&per_page=100",
      "finicky-issues",
      240 // Cache for 4 hours
    );

    // Fetch participation stats (commits by owner vs. non-owner)
    const participation = await fetchWithCache(
      "https://api.github.com/repos/johnste/finicky/stats/participation",
      "finicky-participation",
      240 // Cache for 4 hours
    );

    // Fetch code frequency (additions/deletions per week)
    const codeFrequency = await fetchWithCache(
      "https://api.github.com/repos/johnste/finicky/stats/code_frequency",
      "finicky-code-frequency",
      240 // Cache for 4 hours
    );

    // Process the contributors data to extract timeline
    const contributorTimeline = processContributorTimeline(contributors);

    return {
      commitActivity,
      contributorsCount: contributors.length || 0,
      issuesCount: issues.length || 0,
      participation,
      codeFrequency,
      contributorTimeline,
      // Extract just the data we need for the chart
      weeklyCommits: commitActivity.slice(-26).map((week) => ({
        week: week.week * 1000, // Convert to milliseconds for JS Date
        count: week.total,
      })),
    };
  } catch (error) {
    console.error("Failed to fetch repo activity:", error);
    return null;
  }
}

// Process contributor timeline data
function processContributorTimeline(contributors) {
  if (
    !contributors ||
    !Array.isArray(contributors) ||
    contributors.length === 0
  ) {
    return [];
  }

  // Create a map of weeks to contributor counts
  const weekMap = new Map();

  // Process each contributor's weekly data
  contributors.forEach((contributor) => {
    if (contributor.weeks) {
      contributor.weeks.forEach((week) => {
        const timestamp = week.w * 1000; // Convert to milliseconds
        const existing = weekMap.get(timestamp) || {
          count: 0,
          additions: 0,
          deletions: 0,
        };

        // If there were any changes this week, count this contributor as active
        if (week.a > 0 || week.d > 0 || week.c > 0) {
          existing.count += 1;
          existing.additions += week.a;
          existing.deletions += week.d;
        }

        weekMap.set(timestamp, existing);
      });
    }
  });

  // Convert map to sorted array
  const timeline = Array.from(weekMap.entries())
    .map(([week, data]) => ({ week, ...data }))
    .sort((a, b) => a.week - b.week);

  return timeline;
}

// Create activity visualization
async function createActivityChart() {
  // Fetch repository activity data
  const repoData = await fetchRepoActivity();
  if (!repoData) {
    console.error("No activity data available");
    return;
  }

  // Create the activity section if it doesn't exist
  let activitySection = document.querySelector(".activity-section");
  if (!activitySection) {
    activitySection = document.createElement("section");
    activitySection.className = "activity-section";
    activitySection.innerHTML = `
      <h3>Project Activity</h3>
      <div class="chart-container">
        <div class="chart-tabs">
          <button class="tab-button active" data-chart="commits">Commits</button>
          <button class="tab-button" data-chart="contributors">Contributors</button>
          <button class="tab-button" data-chart="code">Code Changes</button>
        </div>
        <canvas id="activity-chart"></canvas>
      </div>
      <div class="metrics-grid">
        <div class="metric-card">
          <span class="metric-value" id="contributor-count">-</span>
          <span class="metric-label">Contributors</span>
        </div>
        <div class="metric-card">
          <span class="metric-value" id="commits-count">-</span>
          <span class="metric-label">Total Commits</span>
        </div>
        <div class="metric-card">
          <span class="metric-value" id="issues-count">-</span>
          <span class="metric-label">Issues</span>
        </div>
      </div>
    `;

    // Insert before footer
    const footer = document.querySelector("footer");
    if (footer) {
      footer.parentNode.insertBefore(activitySection, footer);
    } else {
      document.querySelector(".inner").appendChild(activitySection);
    }
  }

  // Update metrics
  document.getElementById("contributor-count").textContent =
    repoData.contributorsCount || "-";

  // Calculate total commits across all weeks
  const totalCommits = (repoData.commitActivity || []).reduce(
    (total, week) => total + (week.total || 0),
    0
  );
  document.getElementById("commits-count").textContent =
    totalCommits.toLocaleString() || "-";
  document.getElementById("issues-count").textContent =
    repoData.issuesCount || "-";

  // Load Chart.js if not already loaded
  await loadScript(
    "https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js"
  );

  // Initialize the chart with commit data first
  const ctx = document.getElementById("activity-chart").getContext("2d");
  let activityChart = renderCommitChart(ctx, repoData);

  // Set up tab switching
  document.querySelectorAll(".tab-button").forEach((button) => {
    button.addEventListener("click", function () {
      // Remove active class from all buttons
      document.querySelectorAll(".tab-button").forEach((btn) => {
        btn.classList.remove("active");
      });

      // Add active class to clicked button
      this.classList.add("active");

      // Destroy existing chart
      if (activityChart) {
        activityChart.destroy();
      }

      // Render appropriate chart based on selected tab
      const chartType = this.getAttribute("data-chart");
      if (chartType === "commits") {
        activityChart = renderCommitChart(ctx, repoData);
      } else if (chartType === "contributors") {
        activityChart = renderContributorChart(ctx, repoData);
      } else if (chartType === "code") {
        activityChart = renderCodeChangeChart(ctx, repoData);
      }
    });
  });
}

function renderCommitChart(ctx, repoData) {
  // Process commit data to get weekly commits
  const commitWeeks = repoData.commitActivity || [];
  const labels = [];
  const commitCounts = [];

  // Get the last 26 weeks (6 months) of data
  const recentWeeks = commitWeeks.slice(-26);

  // Format the data for the chart
  for (let i = 0; i < recentWeeks.length; i++) {
    const week = recentWeeks[i];
    if (week) {
      // Create date label (Month YYYY)
      const date = new Date(week.week * 1000);
      const monthYear = date.toLocaleDateString("en-US", {
        month: "short",
        year: "2-digit",
      });

      if (i % 4 === 0 || i === recentWeeks.length - 1) {
        labels.push(monthYear);
      } else {
        labels.push("");
      }

      // Get total commits for the week
      commitCounts.push(week.total);
    }
  }

  // Create gradient for the chart
  const gradient = ctx.createLinearGradient(0, 0, 0, 350);
  gradient.addColorStop(0, "rgba(181, 82, 251, 0.7)");
  gradient.addColorStop(1, "rgba(66, 50, 255, 0.1)");

  // Create and return the chart
  return new Chart(ctx, {
    type: "line",
    data: {
      labels: labels,
      datasets: [
        {
          label: "Commits per Week",
          data: commitCounts,
          borderColor: "#b552fb",
          backgroundColor: gradient,
          borderWidth: 2,
          tension: 0.4,
          fill: true,
          pointBackgroundColor: "#b552fb",
          pointRadius: 3,
          pointHoverRadius: 5,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: {
        duration: 1000,
      },
      plugins: {
        legend: {
          display: true,
          labels: {
            color: "rgba(255, 255, 255, 0.7)",
          },
        },
        tooltip: {
          backgroundColor: "rgba(20, 20, 40, 0.95)",
          titleColor: "#fff",
          bodyColor: "#fff",
          borderColor: "rgba(255, 255, 255, 0.1)",
          borderWidth: 1,
        },
      },
      scales: {
        x: {
          grid: {
            color: "rgba(255, 255, 255, 0.05)",
          },
          ticks: {
            color: "rgba(255, 255, 255, 0.6)",
          },
        },
        y: {
          beginAtZero: true,
          grid: {
            color: "rgba(255, 255, 255, 0.05)",
          },
          ticks: {
            color: "rgba(255, 255, 255, 0.6)",
          },
        },
      },
    },
  });
}

function renderContributorChart(ctx, repoData) {
  const timeline = repoData.contributorTimeline || [];
  const labels = [];
  const contributorCounts = [];

  // Get the last year of data
  const recentData = timeline.slice(-52);

  // Format the data for the chart (every 4 weeks)
  for (let i = 0; i < recentData.length; i++) {
    const entry = recentData[i];
    if (entry) {
      // Create date label
      const date = new Date(entry.week);
      const monthYear = date.toLocaleDateString("en-US", {
        month: "short",
        year: "2-digit",
      });

      if (i % 4 === 0 || i === recentData.length - 1) {
        labels.push(monthYear);
      } else {
        labels.push("");
      }

      // Get active contributors for the week
      contributorCounts.push(entry.count);
    }
  }

  // Create gradient for the chart
  const gradient = ctx.createLinearGradient(0, 0, 0, 350);
  gradient.addColorStop(0, "rgba(66, 50, 255, 0.7)");
  gradient.addColorStop(1, "rgba(66, 50, 255, 0.1)");

  // Create and return the chart
  return new Chart(ctx, {
    type: "line",
    data: {
      labels: labels,
      datasets: [
        {
          label: "Active Contributors per Week",
          data: contributorCounts,
          borderColor: "#4232ff",
          backgroundColor: gradient,
          borderWidth: 2,
          tension: 0.4,
          fill: true,
          pointBackgroundColor: "#4232ff",
          pointRadius: 3,
          pointHoverRadius: 5,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: {
        duration: 1000,
      },
      plugins: {
        legend: {
          display: true,
          labels: {
            color: "rgba(255, 255, 255, 0.7)",
          },
        },
        tooltip: {
          backgroundColor: "rgba(20, 20, 40, 0.95)",
          titleColor: "#fff",
          bodyColor: "#fff",
          borderColor: "rgba(255, 255, 255, 0.1)",
          borderWidth: 1,
        },
      },
      scales: {
        x: {
          grid: {
            color: "rgba(255, 255, 255, 0.05)",
          },
          ticks: {
            color: "rgba(255, 255, 255, 0.6)",
          },
        },
        y: {
          beginAtZero: true,
          grid: {
            color: "rgba(255, 255, 255, 0.05)",
          },
          ticks: {
            color: "rgba(255, 255, 255, 0.6)",
          },
        },
      },
    },
  });
}

function renderCodeChangeChart(ctx, repoData) {
  const codeFrequency = repoData.codeFrequency || [];
  const labels = [];
  const additions = [];
  const deletions = [];

  // Get the last 26 weeks (6 months) of data
  const recentData = codeFrequency.slice(-26);

  // Format the data for the chart
  for (let i = 0; i < recentData.length; i++) {
    const entry = recentData[i];
    if (entry) {
      // Create date label
      const date = new Date(entry[0] * 1000);
      const monthYear = date.toLocaleDateString("en-US", {
        month: "short",
        year: "2-digit",
      });

      if (i % 4 === 0 || i === recentData.length - 1) {
        labels.push(monthYear);
      } else {
        labels.push("");
      }

      // Get additions and deletions (make deletions positive for charting)
      additions.push(entry[1]);
      deletions.push(Math.abs(entry[2]));
    }
  }

  // Create and return the chart
  return new Chart(ctx, {
    type: "bar",
    data: {
      labels: labels,
      datasets: [
        {
          label: "Additions",
          data: additions,
          backgroundColor: "rgba(75, 192, 192, 0.7)",
          borderColor: "rgba(75, 192, 192, 1)",
          borderWidth: 1,
        },
        {
          label: "Deletions",
          data: deletions,
          backgroundColor: "rgba(255, 99, 132, 0.7)",
          borderColor: "rgba(255, 99, 132, 1)",
          borderWidth: 1,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: {
        duration: 1000,
      },
      plugins: {
        legend: {
          display: true,
          labels: {
            color: "rgba(255, 255, 255, 0.7)",
          },
        },
        tooltip: {
          backgroundColor: "rgba(20, 20, 40, 0.95)",
          titleColor: "#fff",
          bodyColor: "#fff",
          borderColor: "rgba(255, 255, 255, 0.1)",
          borderWidth: 1,
        },
      },
      scales: {
        x: {
          grid: {
            color: "rgba(255, 255, 255, 0.05)",
          },
          ticks: {
            color: "rgba(255, 255, 255, 0.6)",
          },
        },
        y: {
          beginAtZero: true,
          grid: {
            color: "rgba(255, 255, 255, 0.05)",
          },
          ticks: {
            color: "rgba(255, 255, 255, 0.6)",
          },
        },
      },
    },
  });
}

// Helper function to load external scripts
function loadScript(src) {
  return new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = src;
    script.onload = resolve;
    script.onerror = reject;
    document.head.appendChild(script);
  });
}

// Animate counting up
function animateCount(element, start, end, duration) {
  let startTimestamp = null;
  const step = (timestamp) => {
    if (!startTimestamp) startTimestamp = timestamp;
    const progress = Math.min((timestamp - startTimestamp) / duration, 1);
    const currentCount = Math.floor(progress * (end - start) + start);
    element.textContent = currentCount.toLocaleString();
    if (progress < 1) {
      window.requestAnimationFrame(step);
    }
  };
  window.requestAnimationFrame(step);
}

function setupCodeDisplay() {
  // No longer needed as line numbers and styling are now static in HTML

  // Add typing cursor to the last line if needed
  const lastLine = document.querySelector(
    ".config-example code .line:last-child"
  );
  if (lastLine) {
    // Only add cursor if it doesn't already exist
    if (!lastLine.querySelector(".typing-cursor")) {
      const cursor = document.createElement("span");
      cursor.className = "typing-cursor";
      lastLine.appendChild(cursor);
    }
  }
}

function addPageAnimations() {
  // Create staggered animations for better visual impact
  const elements = [
    { selector: ".inner", delay: 50 },
    { selector: "header", delay: 150 },
    { selector: "section:nth-of-type(1)", delay: 200 },
    { selector: ".config-example", delay: 250 },
    { selector: ".activity-section", delay: 300 },
    { selector: "footer", delay: 350 },
    { selector: ".code", delay: 400 },
  ];

  // Apply initial styles
  elements.forEach((item) => {
    const element = document.querySelector(item.selector);
    if (element) {
      element.style.opacity = "0";
      element.style.transform = "translateY(20px)";
      element.style.transition = "opacity 0.4s ease, transform 0.4s ease";
    }
  });

  // Trigger animations with staggered delays
  elements.forEach((item) => {
    setTimeout(() => {
      const element = document.querySelector(item.selector);
      if (element) {
        element.style.opacity = "1";
        element.style.transform = "translateY(0)";
      }
    }, item.delay);
  });

  // Add background pulse animation
  addBackgroundEffect();
}

function addBackgroundEffect() {
  const style = document.createElement("style");
  style.textContent = `
    @keyframes gradientPulse {
      0% { background-position: 0% 50%; }
      50% { background-position: 100% 50%; }
      100% { background-position: 0% 50%; }
    }

    body:after {
      content: '';
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      z-index: -1;
      opacity: 0.4;
      background: linear-gradient(-45deg, rgba(66, 50, 255, 0.05), rgba(181, 82, 251, 0.05), rgba(66, 50, 255, 0.05));
      background-size: 400% 400%;
      animation: gradientPulse 15s ease infinite;
      pointer-events: none;
    }
  `;
  document.head.appendChild(style);
}

// Initialize everything
async function init() {
  try {
    await Promise.all([showDownloads(), showStarGazers()]);

    setupCodeDisplay();
    // Add activity chart after basic content is loaded
    createActivityChart();
    addPageAnimations();

    // Make links open in new tab
    document.querySelectorAll("a").forEach((link) => {
      link.setAttribute("target", "_blank");
      link.setAttribute("rel", "noopener noreferrer");
    });
  } catch (error) {
    console.error("Error initializing page:", error);
  }
}

// Start when page is loaded
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", init);
} else {
  init();
}
