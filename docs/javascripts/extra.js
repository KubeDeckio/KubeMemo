var kbTopButtonHandlerBound = false;

function bindKubeMemoTopButton() {
  var topButton = document.querySelector(".md-top");
  if (!topButton) {
    return;
  }

  var threshold = 240;

  function syncTopButton() {
    if (window.scrollY > threshold) {
      topButton.classList.add("kb-top-visible");
    } else {
      topButton.classList.remove("kb-top-visible");
    }
  }

  syncTopButton();

  if (!kbTopButtonHandlerBound) {
    window.addEventListener("scroll", syncTopButton, { passive: true });
    kbTopButtonHandlerBound = true;
  }
}

document.addEventListener("DOMContentLoaded", bindKubeMemoTopButton);

if (typeof document$ !== "undefined" && document$.subscribe) {
  document$.subscribe(function () {
    bindKubeMemoTopButton();
  });
}
