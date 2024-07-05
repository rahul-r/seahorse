function start(containerName) {
  document.getElementById("spinner").style.display = 'block';
  startStop(containerName, true);
}

function stop(containerName) {
  document.getElementById("spinner").style.display = 'block';
  startStop(containerName, false);
}

function restart(containerName) {
  document.getElementById("spinner").style.display = 'block';
  startStop(containerName, false);
  startStop(containerName, true);
}

function startStop(containerName, isStart) {
  console.log(`startStop ${containerName} ${isStart}`)
  fetch(isStart ? "/start" : "/stop", {
    method: "POST",
    body: containerName,
  })
    .then((response) => response.text())
    .then((_resp) => window.location.reload(true));
}

function update(containerName) {
  document.getElementById("spinner").style.display = 'block';
  fetch("/update", {
    method: "POST",
    body: containerName,
  })
    .then((response) => response.text())
    .then((resp) => console.log(resp));
}

function install(containerName) {
  document.getElementById("spinner").style.display = 'block';
  fetch("/install", {
    method: "POST",
    body: containerName,
  })
    .then((response) => response.text())
    .then((resp) => console.log(resp));
}
