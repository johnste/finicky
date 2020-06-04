async function showDownloads() {
  let response = await fetch(
    "https://api.github.com/repos/johnste/finicky/releases"
  );
  let result = await response.json();

  let total = 0;

  const downloads = Object.fromEntries(
    result
      .filter((v) => v?.assets?.[0])
      .map((v) => [v.tag_name, v.assets[0].download_count])
  );

  console.log(downloads);
  document.getElementsByClassName(
    "download-count"
  )[0].innerHTML = JSON.stringify({ total, ...downloads }, null, 2);
}

async function showStarGazers() {
  let response = await fetch("https://api.github.com/repos/johnste/finicky");
  let result = await response.json();

  console.log(result.stargazers_count);

  document.getElementsByClassName("star-count")[0].innerHTML =
    result.stargazers_count;
}

showDownloads();
showStarGazers();
