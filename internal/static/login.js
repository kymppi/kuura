function check(event) {
  event.preventDefault();
  alert(SRP_G, SRP_P);
  const username = document.getElementById('username').value;
  const password = document.getElementById('password').value;
  console.log('Username:', username);
  console.log('Password:', password);
}

document.getElementById('login-form').addEventListener('submit', check);
