function start(containerName) {
  startStop(containerName, true);
}

function stop(containerName) {
  startStop(containerName, false);
}

function restart(containerName) {
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
  fetch("/update", {
    method: "POST",
    body: containerName,
  })
    .then((response) => response.text())
    .then((resp) => console.log(resp));
}

function install(containerName) {
  fetch("/install", {
    method: "POST",
    body: containerName,
  })
    .then((response) => response.text())
    .then((resp) => console.log(resp));
}
