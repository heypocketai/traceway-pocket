import { init } from "@tracewayapp/jquery";

init("frontend-dev-token@http://localhost:8082/api/report");

function addLog(msg) {
  var time = new Date().toLocaleTimeString();
  var $log = $("#log");
  var text = $log.text();
  if (text === "Waiting for actions...") text = "";
  $log.text(text + "[" + time + "] " + msg + "\n");
}

$(function () {
  $("#btn-error").on("click", function () {
    addLog("Calling GET /api/test-error ...");
    $.ajax({
      url: "/api/test-error",
      method: "GET",
      complete: function (jqXHR) {
        var traceId = jqXHR.getResponseHeader("traceway-trace-id");
        addLog(
          "Response: " +
            jqXHR.status +
            " \u2014 traceway-trace-id: " +
            (traceId || "not present"),
        );
        if (jqXHR.status >= 400) {
          addLog("Exception auto-captured by @tracewayapp/jquery with distributedTraceId=" + traceId);
        }
      },
    });
  });

  $("#btn-success").on("click", function () {
    addLog("Calling GET /api/test-success ...");
    $.ajax({
      url: "/api/test-success",
      method: "GET",
      complete: function (jqXHR) {
        var traceId = jqXHR.getResponseHeader("traceway-trace-id");
        addLog(
          "Response: " +
            jqXHR.status +
            " \u2014 traceway-trace-id: " +
            (traceId || "not present"),
        );
      },
    });
  });

  $("#btn-logs").on("click", function () {
    addLog("Emitting console logs at every level ...");
    console.debug("traceway-demo: debug ping", { step: 1 });
    console.log("traceway-demo: cart loaded", { items: 3, total: 49.99 });
    console.info("traceway-demo: user resumed session", { userId: "u_42" });
    console.warn("traceway-demo: slow response from /api/test-success");
    console.error("traceway-demo: payment retry exhausted", { attempts: 3 });
    addLog("5 console.* calls captured into the rolling log buffer.");
  });

  $("#btn-network").on("click", function () {
    addLog("Firing 4 network requests (mix of fetch + jQuery, success + 404) ...");

    fetch("/api/test-success")
      .then(function (r) { addLog("fetch GET /api/test-success -> " + r.status); });

    fetch("/api/test-log-levels")
      .then(function (r) { addLog("fetch GET /api/test-log-levels -> " + r.status); });

    fetch("/api/does-not-exist")
      .then(function (r) { addLog("fetch GET /api/does-not-exist -> " + r.status); });

    $.ajax({
      url: "/api/test-spans-with-logs",
      method: "GET",
      complete: function (jqXHR) {
        addLog("jQuery GET /api/test-spans-with-logs -> " + jqXHR.status);
      },
    });
  });

  $("#btn-clear").on("click", function () {
    $("#log").text("Waiting for actions...");
  });
});
