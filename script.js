async function showDownloads() {
  let response = await fetch(
    "https://api.github.com/repos/johnste/finicky/releases",
    {
      mode: "cors",
    }
  );
  let result = await response.json();

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
      return `<tr><td>${version}</td><td>${downloads}</td></tr>`;
    })
    .join("");

  document.querySelector(".download-count tbody").innerHTML =
    `<tr class="total"><td>total</td><td>${total}</td></tr>` + rows;
}

async function showStarGazers() {
  let response = await fetch("https://api.github.com/repos/johnste/finicky", {
    mode: "cors",
  });
  let result = await response.json();

  document.getElementsByClassName("star-count")[0].innerHTML =
    result.stargazers_count;
}

showDownloads();
showStarGazers();
