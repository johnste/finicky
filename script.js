async function showDownloads() {
  let response = fetch("https://api.github.com/repos/johnste/finicky/releases");
  let result = await response.json();

  let total = 0;

  const downloads = Object.fromEntries(
    result
      .filter((v) => current?.assets?.[0])
      .map((v) => [current.tag_name, current.assets[0].download_count])
  );

  document.getElementsByClassName("download-count").innerHTML = JSON.stringify(
    { total, ...downloads },
    null,
    2
  );
}

async function showStarGazers() {
  let response = fetch("https://api.github.com/repos/johnste/finicky");
  let result = await response.json();

  document.getElementsByClassName(".star-count").innerHTML =
    result.stargazers_count;
}

showDownloads();
showStarGazers();
