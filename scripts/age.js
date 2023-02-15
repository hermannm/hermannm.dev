const birthday = new Date("1999-09-12");

const ageField = document.querySelector("#age-field");
if (ageField) {
  ageField.textContent = ageFromBirthday(birthday).toString();
}

/** @param {Date} birthday, @returns {number} */
function ageFromBirthday(birthday) {
  const now = new Date();

  let age = now.getFullYear() - birthday.getFullYear();

  if (
    now.getMonth() < birthday.getMonth() ||
    (now.getMonth() === birthday.getMonth() && now.getDate() < birthday.getDate())
  ) {
    age--;
  }

  return age;
}
