(function () {
  "use strict";

  var CHARS = "01#{}[]/<>•·:=+-*";
  var FONT_SIZE = 15;
  var FONT_FAMILY = "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace";

  function initHeroCanvas() {
    var hero = document.getElementById("hero-demo");
    var reduceMotion = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;

    if (!hero || reduceMotion) {
      return;
    }

    var canvas = document.createElement("canvas");
    canvas.className = "hero-canvas";
    canvas.setAttribute("aria-hidden", "true");
    hero.insertBefore(canvas, hero.firstChild);

    var ctx = canvas.getContext("2d");
    var dpr = Math.min(window.devicePixelRatio || 1, 2);
    var columns = 0;
    var drops = [];
    var running = false;
    var rafId = null;

    function resize() {
      var rect = hero.getBoundingClientRect();
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
      canvas.style.width = rect.width + "px";
      canvas.style.height = rect.height + "px";
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

      columns = Math.max(1, Math.floor(rect.width / FONT_SIZE));
      drops = new Array(columns).fill(0).map(function () {
        return Math.random() * -50;
      });
    }

    function draw() {
      var rect = hero.getBoundingClientRect();

      ctx.fillStyle = "rgba(10, 10, 10, 0.15)";
      ctx.fillRect(0, 0, rect.width, rect.height);
      ctx.font = FONT_SIZE + "px " + FONT_FAMILY;

      for (var i = 0; i < columns; i++) {
        var char = CHARS[Math.floor(Math.random() * CHARS.length)];
        var x = i * FONT_SIZE;
        var y = drops[i] * FONT_SIZE;

        ctx.fillStyle =
          Math.random() < 0.02 ? "#c8ffc8" : "rgba(95, 215, 95, 0.55)";
        ctx.fillText(char, x, y);

        if (y > rect.height && Math.random() > 0.975) {
          drops[i] = 0;
        }
        drops[i] += 0.4 + Math.random() * 0.4;
      }

      if (running) {
        rafId = requestAnimationFrame(draw);
      }
    }

    function start() {
      if (running) {
        return;
      }
      running = true;
      rafId = requestAnimationFrame(draw);
    }

    function stop() {
      running = false;
      if (rafId) {
        cancelAnimationFrame(rafId);
        rafId = null;
      }
    }

    resize();
    window.addEventListener("resize", resize);

    if (typeof window.IntersectionObserver === "function") {
      var heroObserver = new IntersectionObserver(
        function (entries) {
          entries.forEach(function (entry) {
            if (entry.isIntersecting) {
              start();
            } else {
              stop();
            }
          });
        },
        { threshold: 0 }
      );
      heroObserver.observe(hero);
    } else {
      start();
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initHeroCanvas);
  } else {
    initHeroCanvas();
  }
})();
