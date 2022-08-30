/** @type {HTMLDetailsElement | null} */
const aboutSection = document.querySelector("#about-section");

const MOBILE_BREAKPOINT = "600px";

const mediaQuery = window.matchMedia(`(min-width: ${MOBILE_BREAKPOINT})`);
openAboutSection(mediaQuery);
mediaQuery.addEventListener("change", openAboutSection);

/** @param {MediaQueryList} mediaQuery */
function openAboutSection(mediaQuery) {
  if (!aboutSection) return;

  // Opens the about section by default on desktop, and keeps it closed on mobile.
  aboutSection.open = mediaQuery.matches;
}
