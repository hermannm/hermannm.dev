const MODAL_SUFFIX = "-modal";
const MODAL_OPENER_SUFFIX = "-modal-opener";
const MODAL_CLOSER_SUFFIX = "-modal-closer";

const MODAL_IDS = [
  "indok-web",
  "bfh",
  "coffeetalk",
  "corona-defense",
  "rov-sim",
  "export-control",
  "advent-of-rust",
  "minesweeper",
  "gruvbox-plain",
  "ignite",
  "ntnu-work",
  "fresh",
  "norlandia",
  "ntnu-study",
  "foss",
];

for (const modalId of MODAL_IDS) {
  /** @type {HTMLDialogElement | null} */
  const modal = document.getElementById(`${modalId}${MODAL_SUFFIX}`);

  /** @type {HTMLButtonElement | null} */
  const modalOpener = document.getElementById(`${modalId}${MODAL_OPENER_SUFFIX}`);

  modalOpener?.addEventListener("click", () => {
    modal?.showModal();
  });

  /** @type {HTMLButtonElement | null} */
  const modalCloser = document.getElementById(`${modalId}${MODAL_CLOSER_SUFFIX}`);

  modalCloser?.addEventListener("click", () => {
    modal?.close();
  });
}
