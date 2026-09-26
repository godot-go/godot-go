/* ============================================================
   godot-go — site interactivity
   Particle field, scroll reveal, copy-to-clipboard.
   No game, no konami, no sound.
   ============================================================ */
(function () {
  "use strict";

  var reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  /* ---------- Particle field ---------- */
  function initParticles() {
    var canvas = document.getElementById("particles");
    if (!canvas || reducedMotion) return;

    var ctx = canvas.getContext("2d");
    if (!ctx) return;

    var dpr = Math.min(window.devicePixelRatio || 1, 2);
    var particles = [];
    var COUNT = 70;
    var LINK_DIST = 130;
    var running = true;
    var rafId = null;

    function resize() {
      var w = window.innerWidth;
      var h = window.innerHeight;
      canvas.width = Math.floor(w * dpr);
      canvas.height = Math.floor(h * dpr);
      canvas.style.width = w + "px";
      canvas.style.height = h + "px";
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    }

    function spawn() {
      particles = [];
      for (var i = 0; i < COUNT; i++) {
        particles.push({
          x: Math.random() * window.innerWidth,
          y: Math.random() * window.innerHeight,
          vx: (Math.random() - 0.5) * 0.25,
          vy: (Math.random() - 0.5) * 0.25,
          r: Math.random() * 1.6 + 0.6,
          hue: Math.random() < 0.5 ? "34,211,238" : "139,92,246"
        });
      }
    }

    function step() {
      if (!running) return;
      var w = window.innerWidth;
      var h = window.innerHeight;
      ctx.clearRect(0, 0, w, h);

      var i, j, p, q, dx, dy, dist;
      for (i = 0; i < particles.length; i++) {
        p = particles[i];
        p.x += p.vx;
        p.y += p.vy;
        if (p.x < -10) p.x = w + 10;
        else if (p.x > w + 10) p.x = -10;
        if (p.y < -10) p.y = h + 10;
        else if (p.y > h + 10) p.y = -10;

        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        ctx.fillStyle = "rgba(" + p.hue + ",0.55)";
        ctx.fill();
      }

      for (i = 0; i < particles.length; i++) {
        p = particles[i];
        for (j = i + 1; j < particles.length; j++) {
          q = particles[j];
          dx = p.x - q.x;
          dy = p.y - q.y;
          dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < LINK_DIST) {
            var alpha = (1 - dist / LINK_DIST) * 0.14;
            ctx.beginPath();
            ctx.moveTo(p.x, p.y);
            ctx.lineTo(q.x, q.y);
            ctx.strokeStyle = "rgba(120,170,240," + alpha + ")";
            ctx.lineWidth = 1;
            ctx.stroke();
          }
        }
      }

      rafId = requestAnimationFrame(step);
    }

    resize();
    spawn();
    step();

    window.addEventListener("resize", function () {
      resize();
    });

    // Pause when the tab is hidden to save CPU.
    document.addEventListener("visibilitychange", function () {
      if (document.hidden) {
        running = false;
        if (rafId) cancelAnimationFrame(rafId);
      } else if (!running) {
        running = true;
        rafId = requestAnimationFrame(step);
      }
    });
  }

  /* ---------- Scroll reveal ---------- */
  function initReveal() {
    var targets = document.querySelectorAll(".reveal");
    if (!targets.length) return;

    if (reducedMotion || !("IntersectionObserver" in window)) {
      targets.forEach(function (el) { el.classList.add("revealed"); });
      return;
    }

    var io = new IntersectionObserver(
      function (entries) {
        entries.forEach(function (entry) {
          if (entry.isIntersecting) {
            entry.target.classList.add("revealed");
            io.unobserve(entry.target);
          }
        });
      },
      { threshold: 0.12 }
    );

    targets.forEach(function (el) { io.observe(el); });
  }

  /* ---------- Copy-to-clipboard ---------- */
  function initCopyButtons() {
    var buttons = document.querySelectorAll(".copy-btn");

    function flash(btn, ok) {
      var original = "Copy";
      btn.textContent = ok ? "Copied" : "Failed";
      btn.classList.toggle("copied", ok);
      setTimeout(function () {
        btn.textContent = original;
        btn.classList.remove("copied");
      }, 1500);
    }

    function legacyCopy(text) {
      var ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.style.position = "fixed";
      ta.style.opacity = "0";
      document.body.appendChild(ta);
      ta.select();
      var ok = false;
      try {
        ok = document.execCommand("copy");
      } catch (e) {
        ok = false;
      }
      document.body.removeChild(ta);
      return ok;
    }

    buttons.forEach(function (btn) {
      btn.addEventListener("click", function () {
        var text = btn.getAttribute("data-copy") || "";
        if (navigator.clipboard && navigator.clipboard.writeText) {
          navigator.clipboard
            .writeText(text)
            .then(function () { flash(btn, true); })
            .catch(function () { flash(btn, legacyCopy(text)); });
        } else {
          flash(btn, legacyCopy(text));
        }
      });
    });
  }

  /* ---------- OS tabs (Get Started) ---------- */
  function initOsTabs() {
    var tabs = Array.prototype.slice.call(document.querySelectorAll(".os-tab"));
    if (!tabs.length) return;

    function select(tab, focus) {
      tabs.forEach(function (t) {
        var selected = t === tab;
        t.setAttribute("aria-selected", selected ? "true" : "false");
        t.tabIndex = selected ? 0 : -1;
        var panel = document.getElementById(t.getAttribute("aria-controls"));
        if (panel) panel.hidden = !selected;
      });
      if (focus) tab.focus();
    }

    tabs.forEach(function (tab, i) {
      tab.addEventListener("click", function () {
        select(tab, false);
      });
      tab.addEventListener("keydown", function (e) {
        var idx = null;
        if (e.key === "ArrowRight") idx = (i + 1) % tabs.length;
        else if (e.key === "ArrowLeft") idx = (i - 1 + tabs.length) % tabs.length;
        else if (e.key === "Home") idx = 0;
        else if (e.key === "End") idx = tabs.length - 1;
        if (idx !== null) {
          e.preventDefault();
          select(tabs[idx], true);
        }
      });
    });

    // Auto-select the visitor's own OS when detectable.
    var platform =
      (navigator.userAgentData && navigator.userAgentData.platform) ||
      navigator.platform ||
      navigator.userAgent ||
      "";
    var guess = /win/i.test(platform)
      ? "windows"
      : /mac|iphone|ipad/i.test(platform)
        ? "macos"
        : "linux";
    var match = tabs.filter(function (t) {
      return t.getAttribute("data-os") === guess;
    })[0];
    if (match) select(match, false);
  }

  document.addEventListener("DOMContentLoaded", function () {
    initParticles();
    initReveal();
    initCopyButtons();
    initOsTabs();
  });
})();
