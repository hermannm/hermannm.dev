const ageField = document.querySelector("#age-field");
if (ageField) {
  ageField.textContent = ageFromBirthday(new Date("1999-09-12")).toString();
}

/** @param {Date} birthday, @returns {number} */
function ageFromBirthday(birthday) {
  const now = new Date();

  let age = now.getFullYear() - birthday.getFullYear();

  const birthdayCelebratedThisYear =
    now.getMonth() > birthday.getMonth() ||
    (now.getMonth() === birthday.getMonth() && now.getDate() >= birthday.getDate());

  if (!birthdayCelebratedThisYear) {
    age--;
  }

  return age;
}
