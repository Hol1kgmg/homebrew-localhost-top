(function () {
  "use strict";

  async function loadPartial(mountSelector, url) {
    var mount = document.querySelector(mountSelector);
    if (!mount) {
      return;
    }
    var response = await fetch(url);
    mount.innerHTML = await response.text();
  }

  window.loadPartial = loadPartial;
})();
