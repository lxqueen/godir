function dropdown(element) {
  var target = document.getElementById(element.id.split('_')[1])
  if (element.checked) {
    target.style.display = "block";
    element.parentElement.classList.add('checked')
  } else {
    target.style.display = "none";
    element.parentElement.classList.remove('checked')
  }
}
