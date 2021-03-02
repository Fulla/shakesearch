const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}&m=${data.mode}`).then((response) => {
      response.json().then((results) => {
        Controller.updateTable(results);
      });
    });
  },

  updateTable: (results) => {
    const body = document.getElementById("result-body");
    body.innerHTML = ''
    for (let result of results) {
      const hr = document.createElement("hr");
      body.appendChild(hr);
      const res = document.createElement("div");
      // const work = document.createElement("h4")
      // work.innerText = `from ${result.work}:`
      const text = result.text.join("<br />")
      res.innerHTML = (`<h4>from ${result.work}</h4>`+
        `<p>${text}</p>`);
      res.className = "result";
      body.appendChild(res);
    }
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
