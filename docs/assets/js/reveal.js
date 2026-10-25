(function () {
  "use strict";

  function initSiteHeader() {
    var headerEl = document.querySelector(".site-header");

    if (!headerEl) {
      return;
    }

    window.addEventListener(
      "hero-demo:complete",
      function () {
        headerEl.classList.add("is-visible");
      },
      { once: true }
    );
  }

  function initReveal() {
    var revealEls = Array.prototype.slice.call(
      document.querySelectorAll("[data-reveal]")
    );
    var scrollCueEl = document.getElementById("demo-scroll-cue");
    var reduceMotion = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;

    initSiteHeader();

    if (reduceMotion || typeof window.IntersectionObserver !== "function") {
      revealEls.forEach(function (el) {
        el.classList.add("is-visible");
      });
      return;
    }

    var revealObserver = new IntersectionObserver(
      function (entries) {
        entries.forEach(function (entry) {
          if (entry.isIntersecting) {
            entry.target.classList.add("is-visible");
            revealObserver.unobserve(entry.target);
          }
        });
      },
      { threshold: 0.15 }
    );

    revealEls.forEach(function (el) {
      revealObserver.observe(el);
    });

    if (scrollCueEl) {
      var hideCueOnScroll = function () {
        scrollCueEl.classList.add("is-scrolled");
        window.removeEventListener("scroll", hideCueOnScroll);
      };
      window.addEventListener("scroll", hideCueOnScroll, { passive: true });
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initReveal);
  } else {
    initReveal();
  }

  window.initReveal = initReveal;
})();
