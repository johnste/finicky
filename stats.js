function getCachedData(key, expiryMinutes = 60) {
  try {
    const cached = localStorage.getItem(key);
    if (!cached) return null;
    const { timestamp, data } = JSON.parse(cached);
    if (Date.now() < timestamp + expiryMinutes * 60 * 1000) return data;
  } catch (e) {}
  return null;
}

function setCachedData(key, data) {
  try {
    localStorage.setItem(key, JSON.stringify({ timestamp: Date.now(), data }));
  } catch (e) {}
}

async function fetchWithCache(url, cacheKey, expiryMinutes = 60) {
  const cached = getCachedData(cacheKey, expiryMinutes);
  if (cached) return cached;
  const response = await fetch(url, { mode: "cors" });
  if (!response.ok) throw new Error(`API error: ${response.status}`);
  const data = await response.json();
  setCachedData(cacheKey, data);
  return data;
}

async function showStars() {
  try {
    const data = await fetchWithCache(
      "https://api.github.com/repos/johnste/finicky",
      "finicky-repo", 180
    );
    const el = document.querySelector(".star-count");
    if (el) el.textContent = data.stargazers_count.toLocaleString();
  } catch (e) {
    const el = document.querySelector(".star-count");
    if (el) el.textContent = "—";
  }
}

async function showDownloads() {
  try {
    const releases = await fetchWithCache(
      "https://api.github.com/repos/johnste/finicky/releases",
      "finicky-releases", 120
    );
    const valid = releases.filter((r) => r?.assets?.[0]);
    const total = valid.reduce((s, r) => s + r.assets[0].download_count, 0);
    const rows = valid
      .map((r) => `<tr><td>${r.tag_name}</td><td>${r.assets[0].download_count.toLocaleString()}</td></tr>`)
      .join("");
    document.querySelector(".download-count tbody").innerHTML =
      `<tr class="total"><td>Total</td><td>${total.toLocaleString()}</td></tr>` + rows;
  } catch (e) {
    console.error("Downloads fetch failed", e);
  }
}

document.addEventListener("DOMContentLoaded", () => {
  showStars();
  showDownloads();
});
