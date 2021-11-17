import React from 'react';
import './App.less';
import {Solver} from "./components/Solver/solver";

function App() {
  return (
    <div className="App">
      <header className="AppHeader">
        <p>
          SAT Solver <a href="https://github.com/styczynski/go-sat-solver">(view on Github)</a>
        </p>
        <Solver />
      </header>
    </div>
  );
}

export default App;
