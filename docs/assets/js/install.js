(function () {
  "use strict";

  function initInstallCopy() {
    var buttons = Array.prototype.slice.call(
      document.querySelectorAll(".copy-btn[data-copy]")
    );

    buttons.forEach(function (btn) {
      btn.addEventListener("click", function () {
        var text = btn.getAttribute("data-copy");
        if (!navigator.clipboard) {
          return;
        }

        navigator.clipboard.writeText(text).then(function () {
          var originalLabel = btn.textContent;
          btn.textContent = "Copied!";
          btn.classList.add("is-copied");

          setTimeout(function () {
            btn.textContent = originalLabel;
            btn.classList.remove("is-copied");
          }, 1500);
        });
      });
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initInstallCopy);
  } else {
    initInstallCopy();
  }
})();
