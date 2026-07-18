(function () {
  "use strict";

  /* ---------- mobile nav ---------- */
  var toggle = document.getElementById("navToggle");
  var nav = document.getElementById("nav");
  if (toggle && nav) {
    toggle.addEventListener("click", function () {
      var open = nav.classList.toggle("is-open");
      toggle.setAttribute("aria-expanded", open ? "true" : "false");
    });
    nav.querySelectorAll("a").forEach(function (link) {
      link.addEventListener("click", function () {
        nav.classList.remove("is-open");
        toggle.setAttribute("aria-expanded", "false");
      });
    });
  }

  /* ---------- scroll reveal ---------- */
  var reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  var revealables = document.querySelectorAll(
    ".about__panel, .stack__panel, .project-card, .footer__lead, .footer__links"
  );
  if (!reduceMotion && "IntersectionObserver" in window) {
    revealables.forEach(function (el) {
      el.style.opacity = "0";
      el.style.transform = "translateY(14px)";
      el.style.transition = "opacity .5s ease, transform .5s ease";
    });
    var io = new IntersectionObserver(
      function (entries) {
        entries.forEach(function (entry) {
          if (entry.isIntersecting) {
            entry.target.style.opacity = "1";
            entry.target.style.transform = "translateY(0)";
            io.unobserve(entry.target);
          }
        });
      },
      { threshold: 0.15, rootMargin: "0px 0px -40px 0px" }
    );
    revealables.forEach(function (el) { io.observe(el); });
  }

  /* ---------- lanyard badge: drag-to-swing ---------- */
  var lanyard = document.getElementById("lanyard");
  var badge = document.getElementById("badge");
  if (!lanyard || !badge || reduceMotion) return;

  var dragging = false;
  var pivotX = 0;
  var pivotY = 0;
  var currentAngle = 0;
  var velocity = 0;

  function pivotPoint() {
    var rect = lanyard.getBoundingClientRect();
    pivotX = rect.left + rect.width / 2;
    pivotY = rect.top;
  }

  function angleFromEvent(clientX, clientY) {
    var dx = clientX - pivotX;
    var dy = Math.max(clientY - pivotY, 1);
    var raw = Math.atan2(dx, dy) * (180 / Math.PI);
    return Math.max(-65, Math.min(65, raw));
  }

  function startDrag(clientX, clientY) {
    dragging = true;
    pivotPoint();
    lanyard.classList.add("is-dragging");
    lanyard.style.animation = "none";
    currentAngle = angleFromEvent(clientX, clientY);
    lanyard.style.transform = "rotate(" + currentAngle + "deg)";
  }

  function moveDrag(clientX, clientY) {
    if (!dragging) return;
    var next = angleFromEvent(clientX, clientY);
    velocity = next - currentAngle;
    currentAngle = next;
    lanyard.style.transform = "rotate(" + currentAngle + "deg)";
  }

  function endDrag() {
    if (!dragging) return;
    dragging = false;
    lanyard.classList.remove("is-dragging");

    // release into a damped spring back to rest, using the last
    // drag velocity as the initial "kick" for a touch of realism.
    var kick = Math.max(-18, Math.min(18, velocity * 3));
    var overshoot = currentAngle * -0.28 + kick;

    lanyard.style.transition = "transform .55s " + "cubic-bezier(.34,1.56,.64,1)";
    lanyard.style.transform = "rotate(" + overshoot + "deg)";

    setTimeout(function () {
      lanyard.style.transition = "transform .6s ease-out";
      lanyard.style.transform = "rotate(0deg)";
      setTimeout(function () {
        lanyard.style.transition = "";
        lanyard.style.transform = "";
        lanyard.style.animation = "badge-swing 5.5s ease-in-out infinite";
      }, 620);
    }, 260);
  }

  badge.addEventListener("mousedown", function (e) {
    startDrag(e.clientX, e.clientY);
    e.preventDefault();
  });
  window.addEventListener("mousemove", function (e) {
    if (dragging) moveDrag(e.clientX, e.clientY);
  });
  window.addEventListener("mouseup", endDrag);

  badge.addEventListener(
    "touchstart",
    function (e) {
      var t = e.touches[0];
      startDrag(t.clientX, t.clientY);
    },
    { passive: true }
  );
  window.addEventListener(
    "touchmove",
    function (e) {
      if (!dragging) return;
      var t = e.touches[0];
      moveDrag(t.clientX, t.clientY);
    },
    { passive: true }
  );
  window.addEventListener("touchend", endDrag);

  // keyboard nudge for accessibility: arrow keys swing the badge
  badge.addEventListener("keydown", function (e) {
    if (e.key !== "ArrowLeft" && e.key !== "ArrowRight") return;
    e.preventDefault();
    lanyard.style.animation = "none";
    currentAngle = e.key === "ArrowLeft" ? -20 : 20;
    lanyard.style.transition = "transform .4s ease-out";
    lanyard.style.transform = "rotate(" + currentAngle + "deg)";
    setTimeout(function () {
      lanyard.style.transform = "rotate(0deg)";
      setTimeout(function () {
        lanyard.style.transition = "";
        lanyard.style.transform = "";
        lanyard.style.animation = "badge-swing 5.5s ease-in-out infinite";
      }, 420);
    }, 60);
  });
})();
