var examples = [
  { amount: "15.00", currency: "USD", merchant: "Render", purpose: "Deploy backend service" },
  { amount: "75.00", currency: "USD", merchant: "AWS", purpose: "GPU server" },
  { amount: "20.00", currency: "USD", merchant: "Steam Video Games", purpose: "Database renewal" },
];

function el(id) {
  return document.getElementById(id);
}

function fillExample(i) {
  var e = examples[i];
  el("amount").value = e.amount;
  el("currency").value = e.currency;
  el("merchant").value = e.merchant;
  el("purpose").value = e.purpose;
}

function sendRequest() {
  var payload = {
    amount: el("amount").value.trim(),
    currency: el("currency").value.trim(),
    merchant: el("merchant").value.trim(),
    purpose: el("purpose").value.trim(),
  };

  var btn = el("send");
  btn.disabled = true;
  btn.textContent = "Checking...";

  fetch("/api/v1/orders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
    .then(function (res) {
      return res.text().then(function (text) {
        var data;
        try {
          data = JSON.parse(text);
        } catch (e) {
          data = { error: text };
        }
        showResult(payload, res.status, data);
      });
    })
    .catch(function () {
      showResult(payload, 0, { error: "Could not reach the server." });
    })
    .then(function () {
      btn.disabled = false;
      btn.textContent = "Send request";
    });
}

function showResult(payload, status, data) {
  var verdict, detail;

  if (status === 201) {
    verdict = "Approved";
    detail = "PayPal order " + data.order_id + " created.";
  } else if (status === 403) {
    verdict = "Blocked";
    detail = data.error || "Blocked.";
  } else {
    verdict = "Not completed";
    detail = data.error || "Unexpected response.";
  }

  var box = el("result");
  box.style.display = "block";
  box.innerHTML = "";

  var v = document.createElement("div");
  v.className = "verdict";
  v.textContent = verdict + " (HTTP " + status + ")";
  box.appendChild(v);

  var d = document.createElement("div");
  d.textContent = detail;
  box.appendChild(d);

  if (status === 201 && data.checkout_url && data.checkout_url.indexOf("https://") === 0) {
    var p = document.createElement("div");
    p.style.marginTop = "6px";
    var a = document.createElement("a");
    a.href = data.checkout_url;
    a.target = "_blank";
    a.rel = "noopener";
    a.textContent = "Open PayPal checkout to approve";
    p.appendChild(a);
    box.appendChild(p);
  }

  addHistory(payload, verdict, detail);
}

function addHistory(payload, verdict, detail) {
  var empty = el("empty");
  if (empty) empty.remove();

  var tr = document.createElement("tr");
  var cells = [new Date().toLocaleTimeString(), payload.merchant, payload.amount + " " + payload.currency, payload.purpose, verdict, detail];
  cells.forEach(function (text) {
    var td = document.createElement("td");
    td.textContent = text;
    tr.appendChild(td);
  });

  var body = el("history");
  body.insertBefore(tr, body.firstChild);
}
