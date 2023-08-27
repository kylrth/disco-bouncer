import React, { useState } from "react";
import { HashRouter as Router, Routes, Route } from "react-router-dom";

import './App.css';
import Login from './component/Login';
import ManageUser from './component/ManageUser';
import HomePage from './component/HomePage';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  const handleLogin = (loggedInUsername) => {
    setIsLoggedIn(true);
  };

  return (
    <div className="App">
       <Router>
        <Routes>
          <Route 
            path="/index.html" 
            element={<HomePage isLoggedIn={isLoggedIn} /> } />
          <Route 
            path="/" 
            element={<HomePage isLoggedIn={isLoggedIn} /> } />
          <Route 
            path="/home" 
            element={<HomePage isLoggedIn={isLoggedIn} /> } />
          <Route 
            path="/manage-user" 
            element={<ManageUser isLoggedIn={isLoggedIn} /> } />
          <Route 
            path="/login" 
            element={<Login onLogin={handleLogin} />} />
        </Routes>
      </Router>
    </div>
  );
}

export default App;
