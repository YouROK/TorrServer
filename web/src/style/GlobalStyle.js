import { createGlobalStyle } from 'styled-components'

export default createGlobalStyle`
  *,
  *::before,
  *::after {  
    margin: 0;
    padding: 0;
    box-sizing: inherit;
  }

  body {  
    font-family: "Open Sans", sans-serif;
    box-sizing: border-box;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    letter-spacing: -0.1px;
    -webkit-tap-highlight-color: transparent;

    @media screen and (display-mode: standalone) {
      height: 100vh;
    }
  }

  button {
    font-family: "Open Sans", sans-serif;
    letter-spacing: -0.1px;
  }
`
