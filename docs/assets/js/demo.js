(function () {
  "use strict";

  var COMMAND = "localhost-top";

  function initHeroDemo() {
    var promptPane = document.getElementById("demo-prompt-pane");
    var terminal = document.getElementById("demo-terminal");
    var tuiPane = document.getElementById("demo-tui-pane");
    var typedEl = document.getElementById("demo-typed");
    var enterEl = document.getElementById("demo-enter");
    var cursorEl = document.getElementById("demo-cursor");
    var scrollCueEl = document.getElementById("demo-scroll-cue");

    if (!promptPane || !terminal || !tuiPane || !typedEl || !enterEl || !cursorEl) {
      return;
    }

    var rows = Array.prototype.slice.call(tuiPane.querySelectorAll("tbody tr"));
    var reduceMotion = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    var hasAnime = typeof window.anime === "function";

    function sleep(ms) {
      return new Promise(function (resolve) {
        setTimeout(resolve, ms);
      });
    }

    function animate(targets, props) {
      if (!hasAnime) {
        return Promise.resolve();
      }
      var instance = window.anime(
        Object.assign({ targets: targets }, props)
      );
      return instance.finished || Promise.resolve();
    }

    function resetDemo() {
      terminal.hidden = true;
      terminal.style.opacity = 0;
      promptPane.hidden = false;
      typedEl.textContent = "";
      enterEl.hidden = true;
      cursorEl.classList.remove("is-hidden");
      if (scrollCueEl) {
        scrollCueEl.hidden = true;
        scrollCueEl.classList.remove("is-visible");
      }
    }

    async function showScrollCue() {
      if (!scrollCueEl) {
        return;
      }
      scrollCueEl.hidden = false;

      if (reduceMotion || !hasAnime) {
        scrollCueEl.classList.add("is-visible");
        return;
      }

      await sleep(400);
      await animate(scrollCueEl, {
        opacity: [0, 1],
        duration: 320,
        easing: "easeOutQuad",
      });
      scrollCueEl.classList.add("is-visible");
    }

    function typeCommand() {
      typedEl.textContent = "";
      var i = 0;
      return new Promise(function (resolve) {
        (function next() {
          if (i >= COMMAND.length) {
            resolve();
            return;
          }
          typedEl.textContent += COMMAND[i];
          i += 1;
          setTimeout(next, 70 + Math.random() * 60);
        })();
      });
    }

    async function pressEnter() {
      await sleep(250);
      cursorEl.classList.add("is-hidden");
      enterEl.hidden = false;
      await sleep(350);
    }

    async function showTerminal() {
      promptPane.hidden = true;
      terminal.hidden = false;

      if (!hasAnime) {
        terminal.style.opacity = 1;
        return;
      }

      window.anime.set(terminal, { opacity: 0, scale: 0.98 });
      window.anime.set(rows, { opacity: 0, translateY: 8 });

      await animate(terminal, {
        opacity: [0, 1],
        scale: [0.98, 1],
        duration: 320,
        easing: "easeOutQuad",
      });

      await animate(rows, {
        opacity: [0, 1],
        translateY: [8, 0],
        delay: window.anime.stagger(60),
        duration: 260,
        easing: "easeOutQuad",
      });
    }

    async function playOnce() {
      resetDemo();

      if (reduceMotion) {
        typedEl.textContent = COMMAND;
        cursorEl.classList.add("is-hidden");
        enterEl.hidden = false;
        await sleep(200);
        await showTerminal();
        await showScrollCue();
        window.dispatchEvent(new CustomEvent("hero-demo:complete"));
        return;
      }

      await sleep(600);
      await typeCommand();
      await pressEnter();
      await showTerminal();
      await showScrollCue();
      window.dispatchEvent(new CustomEvent("hero-demo:complete"));
    }

    playOnce();
  }

  window.initHeroDemo = initHeroDemo;
})();
