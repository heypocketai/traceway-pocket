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

  $("#btn-clear").on("click", function () {
    $("#log").text("Waiting for actions...");
  });
});
