async function parseData() {
  let response = fetch("https://api.github.com/repos/johnste/finicky/releases");
  let data = await response.json();

  data.done(function (result) {
    let total = 0;
    const downloads = result.reduce(function (prev, current) {
      if (current && current.assets && current.assets[0]) {
        total += current.assets[0].download_count;
        return {
          ...prev,
          [current.tag_name]: current.assets[0].download_count,
        };
      }
    }, {});
    $(".download-count").html(JSON.stringify({ total, ...downloads }, null, 2));
  });
}

// $(function () {
//   $.ajax("https://api.github.com/repos/johnste/finicky").done(function (
//     result
//   ) {
//     $(".star-count").html(result.stargazers_count);
//   });
// });
