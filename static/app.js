const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    console.log("data", data)
    const response = fetch(`/search?q=${data.query}&m=${data.mode}`).then((response) => {
      response.json().then((results) => {
        Controller.updateTable(results);
      });
    });
  },

  updateTable: (results) => {
    const body = document.getElementById("result-body");
    for (let result of results) {
      const hr = document.createElement("hr")
      body.appendChild(hr);
      const res = document.createElement("div")
      res.innerHTML = `<p>${result}</p>`
      res.className = "result"
      body.appendChild(res);
    }
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
