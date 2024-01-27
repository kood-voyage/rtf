import "../../styles/navbar.css";


export function NavbarNotLogged() {
  const mainContainer = document.querySelector(".root");
  mainContainer.innerHTML = `
    <nav class="navbar">

    <a href="#/home"> <h1 class="logo">VOYAGE</h1></a>
    </div>
    <div class="navbar-auth">

        <a href="#/login"><div class="navbar-button"><p>Sign in</p></div></a>

        <a href="#/register">
        <div class="navbar-button"><p>Sign up</p></div>
      </a>
      </div>
  </nav>`;
}