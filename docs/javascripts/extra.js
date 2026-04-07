var kbTopScrollBound = false;

function kubeMemoSyncTopButton() {
  var topButton = document.querySelector(".md-top");
  if (!topButton) {
    return;
  }

  if (window.scrollY > 240) {
    topButton.classList.add("kb-top-visible");
  } else {
    topButton.classList.remove("kb-top-visible");
  }
}

function kubeMemoBindTopButton() {
  kubeMemoSyncTopButton();

  if (!kbTopScrollBound) {
    window.addEventListener("scroll", kubeMemoSyncTopButton, { passive: true });
    kbTopScrollBound = true;
  }
}

document.addEventListener("DOMContentLoaded", kubeMemoBindTopButton);

if (typeof document$ !== "undefined" && document$.subscribe) {
  document$.subscribe(function () {
    kubeMemoBindTopButton();
  });
}
